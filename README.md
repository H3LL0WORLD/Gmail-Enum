## Gmail Enum
A fairly descent/fast go program to enumerate gmail accounts using a glitch by [@x0rz](https://twitter.com/x0rz) as described [here](https://blog.0day.rocks/abusing-gmail-to-get-previously-unlisted-e-mail-addresses-41544b62b2)

### Requirements:
- [Golang](https://golang.org)

### Usage:
```sh
$ go build

$ ./Gmail_Enum
Usage of ./Gmail_Enum:
  -i string
        List of accounts to test
  -o string
        Output file (default: Stdout)
  -stdin
        Grab accounts from stdin
  -t int
        Number of threads (default 10)
```
