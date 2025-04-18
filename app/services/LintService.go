package services

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os/exec"
)

// ErrShellCheckNotFound indicates that the shellcheck command was not found in the system's PATH.
var ErrShellCheckNotFound = errors.New("shellcheck command not found")

// ShellCheckIssue represents a single issue reported by shellcheck.
// Fields correspond to the JSON output format of shellcheck.
type ShellCheckIssue struct {
	File      string    `json:"file"`
	Line      int       `json:"line"`
	EndLine   int       `json:"endLine"`
	Column    int       `json:"column"`
	EndColumn int       `json:"endColumn"`
	Level     string    `json:"level"` // e.g., "error", "warning", "info", "style"
	Code      int       `json:"code"`  // e.g., SC2086
	Message   string    `json:"message"`
	Fix       *struct { // Optional fix information
		Replacements []struct {
			Line           int    `json:"line"`
			Column         int    `json:"column"`
			EndLine        int    `json:"endLine"`
			EndColumn      int    `json:"endColumn"`
			InsertionPoint string `json:"insertionPoint"` // "beginning", "end", etc.
			Replacement    string `json:"replacement"`
		} `json:"replacements"`
	} `json:"fix"`
}

// LintService provides functionality to lint shell scripts using shellcheck.
type LintService struct {
	shellCheckPath  string
	commandExecutor CommandExecutorInterface
	lookupExecutor  LookupExecutorInterface
}

type LintStatus string

const (
	LintStatusNotAvailable     LintStatus = "NOT_AVAILABLE"
	LintStatusOK               LintStatus = "OK"
	LintStatusWarning          LintStatus = "WARNING"
	LintStatusError            LintStatus = "ERROR"
	LintStatusShellcheckFailed LintStatus = "SHELLCHECK_FAILED"
)

type LintServiceOption func(*LintService)

func WithCustomCommandExecutor(commandExecutor CommandExecutorInterface) LintServiceOption {
	return func(p *LintService) {
		p.commandExecutor = commandExecutor
	}
}

func WithLookPathExecutor(lookupExecutor LookupExecutorInterface) LintServiceOption {
	return func(p *LintService) {
		p.lookupExecutor = lookupExecutor
	}
}

// NewLintService creates a new LintService instance.
// It checks for the presence of the shellcheck command during initialization.
func NewLintService(options ...LintServiceOption) (*LintService, error) {
	defaultCommandExecutor := &DefaultCommandExecutor{}
	lookupExecutor := &DefaultLookupExecutor{}
	service := &LintService{
		shellCheckPath:  "",
		commandExecutor: defaultCommandExecutor,
		lookupExecutor:  lookupExecutor,
	}
	for _, option := range options {
		option(service)
	}

	path, err := service.lookupExecutor.LookPath("shellcheck")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			slog.Warn("shellcheck command not found in PATH. Linting will be disabled.", "error", err)
			// Return the service anyway, but LintScript will return ErrShellCheckNotFound
			return service, nil
		}
		slog.Error("Error looking up shellcheck path", "error", err)
		return nil, err
	}
	slog.Info("Found shellcheck executable", "path", path)
	service.shellCheckPath = path
	return service, nil
}

// LintScript runs shellcheck on the provided script content and returns the issues found.
// It returns ErrShellCheckNotFound if shellcheck was not found during service initialization.
func (s *LintService) LintScript(scriptContent string) ([]ShellCheckIssue, error) {
	if s.shellCheckPath == "" {
		return nil, ErrShellCheckNotFound
	}

	// Use "--" to indicate end of options and treat subsequent args as filenames (or stdin in this case)
	output, outputErr, err := s.commandExecutor.ExecuteCommandWithStdin(
		s.shellCheckPath,
		[]string{"-s", "bash", "-f", "json", "-x", "--", "-"},
		scriptContent,
	)

	// shellcheck exits with status 1 if there are warnings/errors,
	// status 0 if clean, and other non-zero for operational errors.
	// We only care about operational errors here. JSON parsing handles lint issues.
	if err != nil && output == "" && outputErr == "" {
		slog.Error("Failed to run shellcheck, unknown error", "error", err)
		return nil, &ShellcheckUnknownError{Err: err}
	}

	// Even if shellcheck exits with 1 (lint issues found), stdout should contain the JSON report.
	if output == "" {
		// No output likely means no issues or an operational error already caught.
		if outputErr != "" {
			slog.Warn("shellcheck produced stderr output but no stdout", "stderr", outputErr)
		}
		return []ShellCheckIssue{}, nil // No issues found or reported
	}

	var issues []ShellCheckIssue
	if err := json.Unmarshal([]byte(output), &issues); err != nil {
		slog.Error("Failed to parse shellcheck JSON output", "error", err, "output", output)
		return nil, &ShellcheckParseError{
			Err:    err,
			Output: output,
		}
	}

	slog.Debug("Shellcheck analysis complete", "issueCount", len(issues))
	return issues, nil
}

// IsLintingAvailable checks if the shellcheck tool is available.
func (s *LintService) IsLintingAvailable() bool {
	return s.shellCheckPath != ""
}

func (s *LintService) FormatLintIssuesAsJSON(issues []ShellCheckIssue) string {
	str, err := json.Marshal(issues)
	if err != nil {
		slog.Error("Error formatting lint issues as JSON", "error", err)
		return "[]"
	}
	return string(str)
}

func (s *LintService) GetLintResultingStatus(issues []ShellCheckIssue) LintStatus {
	if len(issues) == 0 {
		return LintStatusOK
	}
	allInfo := true
	for _, issue := range issues {
		if issue.Level == "error" {
			return LintStatusError
		}
		if issue.Level != "info" {
			allInfo = false
		}
	}
	if allInfo {
		return LintStatusOK
	}
	return LintStatusWarning
}
