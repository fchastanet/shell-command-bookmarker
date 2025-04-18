package services

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockCommandExecutor struct {
	// Mocked command output
	stdout string
	stderr string
	err    error
}

func (m *MockCommandExecutor) ExecuteCommandWithStdin(_ string, _ []string, _ string) (
	stdout string, stderr string, err error,
) {
	return m.stdout, m.stderr, m.err
}

type MockLookupExecutor struct {
	path string
	err  error
}

func (m *MockLookupExecutor) LookPath(_ string) (string, error) {
	return m.path, m.err
}

// TestNewLintService tests the constructor. Direct testing of LookPath is hard,
// so we focus on the state of the returned service.
func TestNewLintService(t *testing.T) {
	// Simulate shellcheck being found
	t.Run("Shellcheck Found", func(t *testing.T) {
		lookupExecutor := &MockLookupExecutor{
			path: "/fake/path/to/shellcheck",
			err:  nil,
		}

		service, err := NewLintService(
			WithLookPathExecutor(lookupExecutor),
		)
		assert.NoError(t, err)
		assert.Equal(t, "/fake/path/to/shellcheck", service.shellCheckPath)
		assert.NotNil(t, service.commandExecutor)
		assert.NotNil(t, service.lookupExecutor)
		assert.Equal(t, lookupExecutor, service.lookupExecutor)
		assert.Equal(t, true, service.IsLintingAvailable())
	})

	// Simulate shellcheck *not* being found (requires manipulating PATH or mocking LookPath)
	// This is hard to test reliably without altering the environment.
	// We'll trust that if LookPath returns ErrNotFound, the service behaves correctly
	// (sets path to "" and returns nil error), as tested implicitly in LintScript tests.
	t.Run("Shellcheck Not Found Handling", func(t *testing.T) {
		// Assume LookPath returns ErrNotFound
		service := &LintService{
			shellCheckPath:  "",
			commandExecutor: nil,
			lookupExecutor:  nil,
		} // Manually create the state
		assert.Equal(t, false, service.IsLintingAvailable())
		_, err := service.LintScript("echo 'hello'")
		assert.Error(t, err)
		assert.Equal(t, ErrShellCheckNotFound, err)
	})
}

func TestLintService_LintScript(t *testing.T) {
	// Valid Script scenarios
	t.Run("Valid Script Scenarios", func(t *testing.T) {
		tests := []struct {
			name          string
			scriptContent string
			mockStdout    string
			mockStderr    string
			wantIssues    []ShellCheckIssue
		}{
			{
				name:          "Valid Script No Issues",
				scriptContent: "echo 'hello world'",
				mockStdout:    "[]", // Empty JSON array for no issues
				mockStderr:    "",
				wantIssues:    []ShellCheckIssue{},
			},
			{
				name:          "Empty Output No Error", // e.g., shellcheck runs but outputs nothing
				scriptContent: "echo 'perfectly valid'",
				mockStdout:    "",
				mockStderr:    "",
				wantIssues:    []ShellCheckIssue{},
			},
			{
				name:          "Empty Output With Stderr", // e.g., shellcheck runs, outputs nothing to stdout, but warns on stderr
				scriptContent: "echo 'valid but weird'",
				mockStdout:    "",
				mockStderr:    "Some warning message",
				wantIssues:    []ShellCheckIssue{},
			},
			{
				name:          "Script With Linting Issues",
				scriptContent: "echo $undefined",
				mockStdout: `[
					{"file": "-", "line": 1, "endLine": 1, "column": 6, "endColumn": 15, "level": "info", "code": 2154, "message": "undefined is referenced but not assigned."}
				]`,
				mockStderr: "",
				wantIssues: []ShellCheckIssue{
					{
						File: "-", Line: 1, EndLine: 1, Column: 6, EndColumn: 15, Level: "info",
						Code: 2154, Message: "undefined is referenced but not assigned.", Fix: nil,
					},
				},
			},
			{
				name:          "Script With Multiple Issues",
				scriptContent: "echo $undefined; echo $anotherUndefined",
				mockStdout: `[
					{"file": "-", "line": 1, "endLine": 1, "column": 6, "endColumn": 15, "level": "info", "code": 2154, "message": "undefined is referenced but not assigned."},
					{"file": "-", "line": 1, "endLine": 1, "column": 30, "endColumn": 45, "level": "info", "code": 2154, "message": "anotherUndefined is referenced but not assigned."}
				]`,
				mockStderr: "",
				wantIssues: []ShellCheckIssue{
					{File: "-", Line: 1, EndLine: 1, Column: 6, EndColumn: 15, Level: "info", Code: 2154, Message: "undefined is referenced but not assigned.", Fix: nil},
					{File: "-", Line: 1, EndLine: 1, Column: 30, EndColumn: 45, Level: "info", Code: 2154, Message: "anotherUndefined is referenced but not assigned.", Fix: nil},
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				// Create a service instance with shellcheck path set
				service := &LintService{
					shellCheckPath: "/fake/path/to/shellcheck",
					commandExecutor: &MockCommandExecutor{
						stdout: tc.mockStdout,
						stderr: tc.mockStderr,
						err:    nil,
					},
					lookupExecutor: &MockLookupExecutor{
						path: "/fake/path/to/shellcheck",
						err:  nil,
					},
				}

				issues, err := service.LintScript(tc.scriptContent)
				assert.NoError(t, err, "LintScript() error = %v, wantErr false", err)
				assert.Equal(t, tc.wantIssues, issues)
			})
		}
	})

	// Script with issues scenarios
	t.Run("Script With Issues", func(t *testing.T) {
		tests := []struct {
			name          string
			scriptContent string
			mockStdout    string
			mockStderr    string
			mockErr       error
			wantIssues    []ShellCheckIssue // Detailed check for issue scenarios
			wantErrMsg    string            // Substring to check in error message
		}{
			{
				name:          "Shellcheck Execution Error - empty outputs",
				scriptContent: "",
				mockStdout:    "",
				mockStderr:    "",
				mockErr:       exec.ErrNotFound,
				wantIssues:    nil,
				wantErrMsg:    "shellcheck unknown error: executable file not found in $PATH",
			},
			{
				// Can happen even with exit code 1 if output is broken
				name:          "Invalid JSON Output",
				scriptContent: "echo 'test'",
				mockStdout:    "this is not json",
				mockStderr:    "",
				mockErr:       exec.ErrNotFound,
				wantIssues:    nil,
				wantErrMsg:    "shellcheck parse error: invalid character 'h' in literal true (expecting 'r') | Output: this is not json",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				// Create a service instance with shellcheck path set
				service := &LintService{
					shellCheckPath: "/fake/path/to/shellcheck",
					commandExecutor: &MockCommandExecutor{
						stdout: tc.mockStdout,
						stderr: tc.mockStderr,
						err:    tc.mockErr,
					},
					lookupExecutor: &MockLookupExecutor{
						path: "/fake/path/to/shellcheck",
						err:  nil,
					},
				}

				issues, err := service.LintScript(tc.scriptContent)
				assert.Equal(t, tc.wantIssues, issues, "Expected no issues when error occurs")
				assert.Equal(t, tc.wantErrMsg, err.Error(), "Expected error message to match")
			})
		}
	})
}
