package utils

import "strings"

// RemoveFirstLines removes the first n lines from a string
func RemoveFirstLines(s string, n int) string {
	if n <= 0 {
		return s
	}

	lines := strings.Split(s, "\n")
	if n >= len(lines) {
		return ""
	}

	return strings.Join(lines[n:], "\n")
}
