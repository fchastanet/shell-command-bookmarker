package services

import "fmt"

type ShellcheckUnknownError struct {
	Err error
}

func (e *ShellcheckUnknownError) Error() string {
	return "shellcheck unknown error: " + e.Err.Error()
}

type ShellcheckParseError struct {
	Err    error
	Output string
}

func (e *ShellcheckParseError) Error() string {
	return fmt.Sprintf("shellcheck parse error: %v | Output: %s", e.Err, e.Output)
}
