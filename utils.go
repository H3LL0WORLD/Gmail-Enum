package main

import (
	"regexp"
	"strings"
)

// RemoveInvalidChars removes gmail address' invalid chars
func RemoveInvalidChars(addr string) string {
	// Gmail only allows words (a-z), numbers and dots
	re := regexp.MustCompile(`[^a-z\d\.]`)

	if i := strings.Index(addr, "@"); i != -1 {
		username := re.ReplaceAllString(addr[:i], "")
		return username + addr[i:]
	}

	return re.ReplaceAllString(addr, "")
}
