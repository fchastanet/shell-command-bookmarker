package executors

import (
	"os/exec"
)

type DefaultLookupExecutor struct{}

func (c *DefaultLookupExecutor) LookPath(path string) (string, error) {
	return exec.LookPath(path)
}
