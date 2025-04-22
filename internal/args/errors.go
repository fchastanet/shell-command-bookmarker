package args

import "errors"

// Common errors that can occur during argument processing
var (
	// ErrMissingDBPath is returned when the required db-path argument is missing
	ErrMissingDBPath = errors.New("missing required argument: db-path")

	// ErrFileDoesNotExist is returned when the specified file doesn't exist
	ErrFileDoesNotExist = errors.New("file does not exist")

	// ErrPermissionDenied is returned when access is denied to a file
	ErrPermissionDenied = errors.New("permission denied to access file")

	// ErrAccessingFile is returned for general file access errors
	ErrAccessingFile = errors.New("error accessing file")
)
