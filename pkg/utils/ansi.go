package utils

import (
	"regexp"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func RemoveAnsiCodes(input string) string {
	// Remove any other ANSI escape sequences
	return ansiRegexp.ReplaceAllString(input, "")
}
