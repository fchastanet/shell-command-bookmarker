package command

import (
	"fmt"

	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
)

// ComposeCommandError represents an error when composing a command fails
type ComposeCommandError struct {
	Err error
}

func (e *ComposeCommandError) Error() string {
	return fmt.Sprintf("failed to compose command: %v", e.Err)
}

// ErrNoCommandsSelected is returned when no commands are selected for an operation
type ErrNoCommandsSelected struct{}

func (e *ErrNoCommandsSelected) Error() string {
	return "no commands selected to copy"
}

// ErrClipboardCopyFailed is returned when copying to clipboard fails
type ErrClipboardCopyFailed struct {
	Err error
}

func (e *ErrClipboardCopyFailed) Error() string {
	return "failed to copy to clipboard: " + e.Err.Error()
}

type ErrCommandLoadingFailure struct {
	Err       error
	CommandID resource.ID
}

func (e *ErrCommandLoadingFailure) Error() string {
	return fmt.Sprintf("failed to load command with ID %d: %v", e.CommandID, e.Err)
}
