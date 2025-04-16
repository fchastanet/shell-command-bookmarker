package processors

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const FileMode = 0o644

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantTime      time.Time
		wantElapsed   int
		wantErrString string
	}{
		{
			name:          "Valid timestamp",
			input:         "1618246940:5",
			wantTime:      time.Unix(1618246940, 0),
			wantElapsed:   5,
			wantErrString: "",
		},
		{
			name:          "Invalid format - too few parts",
			input:         "1618246940",
			wantErrString: "invalid timestamp format",
			wantTime:      time.Time{},
			wantElapsed:   0,
		},
		{
			name:          "Invalid format - too many parts",
			input:         "1618246940:5:extra",
			wantErrString: "invalid timestamp format",
			wantTime:      time.Time{},
			wantElapsed:   0,
		},
		{
			name:          "Invalid timestamp",
			input:         "invalid:5",
			wantErrString: "invalid timestamp",
			wantTime:      time.Time{},
			wantElapsed:   0,
		},
		{
			name:          "Invalid elapsed",
			input:         "1618246940:invalid",
			wantErrString: "invalid elapsed format",
			wantTime:      time.Time{},
			wantElapsed:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTime, gotElapsed, err := ParseTimestamp(tt.input)

			if tt.wantErrString != "" {
				if err == nil || err.Error() != tt.wantErrString {
					t.Errorf("ParseTimestamp() error = %v, wantErr %v", err, tt.wantErrString)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseTimestamp() unexpected error: %v", err)
				return
			}

			if !gotTime.Equal(tt.wantTime) {
				t.Errorf("ParseTimestamp() gotTime = %v, want %v", gotTime, tt.wantTime)
			}

			if gotElapsed != tt.wantElapsed {
				t.Errorf("ParseTimestamp() gotElapsed = %v, want %v", gotElapsed, tt.wantElapsed)
			}
		})
	}
}

func TestParseFirstHistoryLine(t *testing.T) {
	tests := []struct {
		name             string
		line             string
		isExtendedFormat bool
		wantCommand      string
		wantTimestamp    time.Time
		wantElapsed      int
		checkTimestamp   bool
	}{
		{
			name:             "Simple format",
			line:             "ls -la",
			isExtendedFormat: false,
			wantCommand:      "ls -la",
			wantElapsed:      0,
			checkTimestamp:   false,
			wantTimestamp:    time.Time{},
		},
		{
			name:             "Extended format",
			line:             ": 1618246940:5;cd /home",
			isExtendedFormat: true,
			wantCommand:      "cd /home",
			wantTimestamp:    time.Unix(1618246940, 0),
			wantElapsed:      5,
			checkTimestamp:   true,
		},
		{
			name:             "Extended format detected but invalid",
			line:             ": aaaaa;ls -l",
			isExtendedFormat: false,
			wantCommand:      ": aaaaa;ls -l",
			wantElapsed:      0,
			checkTimestamp:   false,
			wantTimestamp:    time.Time{},
		},
		{
			name:             "Extended format with missing semicolon",
			line:             ": 1618246940:5 ls -l",
			isExtendedFormat: false,
			wantCommand:      ": 1618246940:5 ls -l",
			wantElapsed:      0,
			checkTimestamp:   false,
			wantTimestamp:    time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamp, elapsed, commandPart, isExtendedFormat := parseFirstHistoryLine(tt.line)

			if commandPart != tt.wantCommand {
				t.Errorf("parseFirstHistoryLine() Command = %v, want %v", commandPart, tt.wantCommand)
			}

			if tt.checkTimestamp && !timestamp.Equal(tt.wantTimestamp) {
				t.Errorf("parseFirstHistoryLine() Timestamp = %v, want %v", timestamp, tt.wantTimestamp)
			}

			if elapsed != tt.wantElapsed {
				t.Errorf("parseFirstHistoryLine() Elapsed = %v, want %v", elapsed, tt.wantElapsed)
			}
			assert.Equal(t, tt.isExtendedFormat, isExtendedFormat, "Expected extended format mismatch")
		})
	}
}

func TestParseBashHistory(t *testing.T) {
	// Create temporary test files
	testDir := t.TempDir()

	// Test case 1: Empty file
	emptyFile := filepath.Join(testDir, "empty_history")
	if err := os.WriteFile(emptyFile, []byte(""), FileMode); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test case 2: Simple format
	simpleFormatFile := filepath.Join(testDir, "simple_history")
	simpleContent := "ls -la\ncd /home\npwd\n"
	if err := os.WriteFile(simpleFormatFile, []byte(simpleContent), FileMode); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test case 3: Extended format
	extendedFormatFile := filepath.Join(testDir, "extended_history")
	extendedContent := ": 1618246940:5;ls -la\n: 1618247000:3;cd /home\n: 1618247060:1;pwd\n"
	if err := os.WriteFile(extendedFormatFile, []byte(extendedContent), FileMode); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test case 4: Mixed format
	mixedFormatFile := filepath.Join(testDir, "mixed_history")
	mixedContent := "ls -la\n: 1618247000:3;cd /home\npwd\n"
	if err := os.WriteFile(mixedFormatFile, []byte(mixedContent), FileMode); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name         string
		historyFile  string
		wantCommands []string
		wantErr      bool
	}{
		{
			name:         "Empty file",
			historyFile:  emptyFile,
			wantCommands: []string{},
			wantErr:      false,
		},
		{
			name:         "Simple format",
			historyFile:  simpleFormatFile,
			wantCommands: []string{"ls -la", "cd /home", "pwd"},
			wantErr:      false,
		},
		{
			name:         "Extended format",
			historyFile:  extendedFormatFile,
			wantCommands: []string{"ls -la", "cd /home", "pwd"},
			wantErr:      false,
		},
		{
			name:         "Mixed format",
			historyFile:  mixedFormatFile,
			wantCommands: []string{"ls -la", "cd /home", "pwd"},
			wantErr:      false,
		},
		{
			name:         "Non-existent file",
			historyFile:  filepath.Join(testDir, "nonexistent"),
			wantCommands: []string{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var commands []string

			historyIngestor := &HistoryIngestor{}
			err := historyIngestor.ParseBashHistory(tt.historyFile, func(cmd HistoryCommand) error {
				commands = append(commands, cmd.Command)
				return nil
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBashHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(commands) != len(tt.wantCommands) {
				t.Errorf("ParseBashHistory() got %d commands, want %d", len(commands), len(tt.wantCommands))
				return
			}

			for i, cmd := range commands {
				if cmd != tt.wantCommands[i] {
					t.Errorf("ParseBashHistory() command[%d] = %v, want %v", i, cmd, tt.wantCommands[i])
				}
			}
		})
	}
}

type callbackError struct{}

func (c *callbackError) Error() string {
	return "callback error"
}

func TestParseBashHistoryWithCallbackError(t *testing.T) {
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "history_file")
	content := "command1\ncommand2\ncommand3\n"

	if err := os.WriteFile(testFile, []byte(content), FileMode); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	expectedErr := &callbackError{}
	callCount := 0

	historyIngestor := &HistoryIngestor{}
	err := historyIngestor.ParseBashHistory(testFile, func(_ HistoryCommand) error {
		callCount++
		if callCount == 2 {
			return expectedErr
		}
		return nil
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("ParseBashHistory() error = %v, want %v", err, expectedErr)
	}

	if callCount != 2 {
		t.Errorf("Callback should have been called %d times, but was called %d times", 2, callCount)
	}
}

func TestParseBashHistoryDefaultPath(t *testing.T) {
	// This test can only verify that providing an empty path doesn't cause errors
	// Skip test if home dir cannot be accessed
	_, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Skipping test: could not get user home directory")
	}

	// Just test that the function doesn't error with empty path
	// We're not actually validating the contents since we don't want to modify the real ~/.bash_history
	historyIngestor := &HistoryIngestor{}
	err = historyIngestor.ParseBashHistory("", func(_ HistoryCommand) error {
		// Don't process commands from actual history file
		return nil
	})

	if err != nil && !os.IsNotExist(err) { // It's OK if ~/.bash_history doesn't exist
		t.Errorf("ParseBashHistory() with default path error = %v", err)
	}
}

func createTempHistoryFile(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "testHistory*.log")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())
	return tmpFile.Name()
}

func TestParseBashHistory_MultiLineError(t *testing.T) {
	ingestor := NewHistoryIngestor()

	//nolint:exhaustruct // This is a test case struct for the test function
	tests := []struct {
		name           string
		historyContent string
		expectedCmds   []HistoryCommand
		expectError    bool
	}{
		{
			name: "Simple format single line",
			historyContent: `echo "hello"
ls -l`,
			expectedCmds: []HistoryCommand{
				{Command: `echo "hello"`},
				{Command: `ls -l`},
			},
		},
		{
			name: "Simple format multi-line",
			historyContent: `echo "line1" \
line2 \
line3
ls -l`,
			expectedCmds: []HistoryCommand{
				{Command: `echo "line1" \
line2 \
line3`},
				{Command: `ls -l`},
			},
		},
		{
			name: "Extended format single line",
			historyContent: `: 1678886400:5;git status
: 1678886410:2;docker ps`,
			expectedCmds: []HistoryCommand{
				{Command: `git status`, Timestamp: time.Unix(1678886400, 0), Elapsed: 5},
				{Command: `docker ps`, Timestamp: time.Unix(1678886410, 0), Elapsed: 2},
			},
		},
		{
			name: "Extended format multi-line",
			historyContent: `: 1678886400:5;git commit -m "multi \
line \
message"
: 1678886410:2;docker ps`,
			expectedCmds: []HistoryCommand{
				{Command: `git commit -m "multi \
line \
message"`, Timestamp: time.Unix(1678886400, 0), Elapsed: 5},
				{Command: `docker ps`, Timestamp: time.Unix(1678886410, 0), Elapsed: 2},
			},
		},
		{
			name: "Mixed formats with multi-line",
			historyContent: `simple command
: 1678886400:5;git commit -m "multi \
line \
message"
another simple \
multi-line
: 1678886410:2;docker ps`,
			expectedCmds: []HistoryCommand{
				{Command: `simple command`},
				{Command: `git commit -m "multi \
line \
message"`, Timestamp: time.Unix(1678886400, 0), Elapsed: 5},
				{Command: `another simple \
multi-line`},
				{Command: `docker ps`, Timestamp: time.Unix(1678886410, 0), Elapsed: 2},
			},
		},
		{
			name: "Multi-line ending with backslash EOF",
			historyContent: `echo "part1" \
part2 \`,
			expectedCmds: []HistoryCommand{
				{Command: `echo "part1" \
part2 \
`}, // Note the trailing space might occur depending on interpretation
			},
		},
		{
			name: "Extended Multi-line ending with backslash EOF",
			historyContent: `: 1678886400:5;git commit \
-m "incomplete" \`,
			expectedCmds: []HistoryCommand{
				{Command: `git commit \
-m "incomplete" \
`, Timestamp: time.Unix(1678886400, 0), Elapsed: 5},
			},
		},
		{
			name:           "Empty file",
			historyContent: ``,
			expectedCmds:   nil,
		},
		{
			name: "File with only empty lines",
			historyContent: `

`,
			expectedCmds: []HistoryCommand{},
		},
		{
			name:           "Invalid extended format (no semicolon)",
			historyContent: `: 1678886400:5 git status`, // Missing semicolon
			expectedCmds: []HistoryCommand{
				{Command: `: 1678886400:5 git status`}, // Treated as simple command
			},
		},
		{
			name:           "Invalid extended format (bad timestamp)",
			historyContent: `: not_a_time:5;git status`,
			expectedCmds: []HistoryCommand{
				{Command: `: not_a_time:5;git status`}, // Treated as simple command
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			historyFilePath := createTempHistoryFile(t, tc.historyContent)
			defer os.Remove(historyFilePath) // Clean up

			var actualCmds []HistoryCommand
			callback := func(cmd HistoryCommand) error {
				// Ignore timestamp for simple format comparison if not set in expected
				if cmd.Timestamp.IsZero() && len(tc.expectedCmds) > len(actualCmds) && tc.expectedCmds[len(actualCmds)].Timestamp.IsZero() {
					// Use a fixed time for comparison or skip timestamp check for simple format
					cmd.Timestamp = time.Time{} // Zero out for comparison
				}
				//nolint:exhaustruct // This is a test case struct for the test function
				if cmd != (HistoryCommand{}) {
					actualCmds = append(actualCmds, cmd)
				}
				return nil
			}

			err := ingestor.ParseBashHistory(historyFilePath, callback)

			require.NoError(t, err)
			// Adjust expected commands' timestamps if they are zero (simple format)
			if len(tc.expectedCmds) > 0 {
				adjustedExpectedCmds := make([]HistoryCommand, len(tc.expectedCmds))
				for i, expected := range tc.expectedCmds {
					adjustedExpectedCmds[i] = expected
					if expected.Timestamp.IsZero() {
						// Find corresponding actual command to ignore its timestamp too
						if i < len(actualCmds) {
							actualCmds[i].Timestamp = time.Time{}
						}
					}
				}
				assert.Equal(t, adjustedExpectedCmds, actualCmds)
			} else {
				assert.Empty(t, actualCmds)
			}
		})
	}
}

func TestParseBashHistory_FileNotFound(t *testing.T) {
	ingestor := NewHistoryIngestor()
	err := ingestor.ParseBashHistory(
		"/non/existent/path/to/.bash_history",
		func(_ HistoryCommand) error {
			t.Fatal("Callback should not be called")
			return nil
		},
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, os.ErrNotExist) // Or the specific error type returned by os.Open
}
