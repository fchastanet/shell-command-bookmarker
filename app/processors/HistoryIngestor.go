package processors

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// HistoryCommand represents a single command entry from bash history
type HistoryCommand struct {
	Command   string
	Timestamp time.Time
	Elapsed   int // elapsed time in seconds
}

// timestampFieldsCount is the number of fields in the extended format
const timestampFieldsCount = 2

// ParseBashHistory reads and parses the bash history file
// It supports both simple format (just commands) and extended format (`:start:elapsed;command`)
func ParseBashHistory(historyFilePath string, callback func(HistoryCommand) error) error {
	// If no specific path is provided, use the default ~/.bash_history
	if historyFilePath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		historyFilePath = filepath.Join(homeDir, ".bash_history")
	}

	file, err := os.Open(historyFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Detect format and parse all lines
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Check if this line is in extended format
		isExtendedFormat := strings.HasPrefix(line, ":")
		cmd := parseHistoryLine(line, isExtendedFormat)
		err := callback(cmd)
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// parseHistoryLine parses a single line from the history file
// based on the detected format
func parseHistoryLine(line string, isExtendedFormat bool) HistoryCommand {
	timestamp := time.Now()
	elapsed := 0
	commandPart := line

	if isExtendedFormat {
		// Parse extended format: ":start:elapsed;command"
		semicolonPos := strings.Index(line, ";")
		if semicolonPos == -1 {
			// No semicolon found, treat as simple format
			return HistoryCommand{
				Timestamp: time.Now(),
				Elapsed:   0,
				Command:   line,
			}
		}

		// Extract timestamps part and command part
		timestampsPart := line[2:semicolonPos] // Skip the leading colon and space
		commandPart = line[semicolonPos+1:]

		// Parse timestamps
		var err error
		timestamp, elapsed, err = ParseTimestamp(timestampsPart)
		if err != nil {
			// Parsing failed, treat as simple format
			return HistoryCommand{
				Timestamp: time.Now(),
				Elapsed:   0,
				Command:   line,
			}
		}
	}

	// Simple format: just the command
	return HistoryCommand{
		Command:   commandPart,
		Timestamp: timestamp,
		Elapsed:   elapsed,
	}
}

var (
	errInvalidTimestampFormat = errors.New("invalid timestamp format")
	errInvalidTimestamp       = errors.New("invalid timestamp")
	errInvalidElapsedFormat   = errors.New("invalid elapsed format")
)

func ParseTimestamp(timestampsPart string) (time.Time, int, error) {
	// Split timestamps part by colon
	timestampFields := strings.Split(timestampsPart, ":")
	if len(timestampFields) != timestampFieldsCount {
		// Invalid format
		return time.Time{}, 0, errInvalidTimestampFormat
	}

	// Parse start timestamp
	startTimestamp, err := strconv.ParseInt(timestampFields[0], 10, 64)
	if err != nil {
		// Parsing failed
		return time.Time{}, 0, errInvalidTimestamp
	}

	// Parse elapsed time
	elapsed, err := strconv.Atoi(timestampFields[1])
	if err != nil {
		// Parsing failed
		return time.Unix(startTimestamp, 0), 0, errInvalidElapsedFormat
	}

	return time.Unix(startTimestamp, 0), elapsed, nil
}
