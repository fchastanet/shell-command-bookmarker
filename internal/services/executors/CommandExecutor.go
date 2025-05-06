package executors

import (
	"bytes"
	"log/slog"
	"os/exec"
	"strings"
)

type DefaultCommandExecutor struct{}

func (*DefaultCommandExecutor) ExecuteCommandWithStdin(cmd string, args []string, stdin string) (
	output string, errOutput string, err error,
) {
	command := exec.Command(cmd, args...)
	command.Stdin = strings.NewReader(stdin)

	slog.Debug(
		"Executing command",
		"command", cmd+" "+strings.Join(args, " "),
		"stdin", stdin,
	)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err = command.Run()
	stdoutStr := stdout.String()
	stderrStr := stderr.String()
	if err != nil {
		slog.Error(
			"Failed to run command",
			"error", err,
			"stdout", stdoutStr,
			"stderr", stderrStr,
		)
	} else {
		slog.Debug(
			"Command executed successfully",
			"stdout", stdoutStr,
			"stderr", stderrStr,
		)
	}
	return stdoutStr, stderrStr, err
}
