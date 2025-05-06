package services

type CommandExecutorInterface interface {
	// ExecuteCommandWithStdin executes a command with stdin and returns the output and error if any.
	ExecuteCommandWithStdin(cmd string, args []string, stdin string) (
		output string, errOutput string, err error,
	)
}

type LookupExecutorInterface interface {
	LookPath(path string) (string, error)
}
