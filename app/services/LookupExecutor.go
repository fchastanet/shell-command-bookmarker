package services

import (
	"os/exec"
)

type LookupExecutorInterface interface {
	// ExecuteCommandWithStdin executes a command with stdin and returns the output and error if any.
	LookPath(path string) (string, error)
}

type DefaultLookupExecutor struct{}

func (c *DefaultLookupExecutor) LookPath(path string) (string, error) {
	return exec.LookPath(path)
}
