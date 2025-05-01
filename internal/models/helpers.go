package models

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
)

// Define static errors
var (
	// ErrCommandPanic is returned when a tea.Cmd panics
	ErrCommandPanic = errors.New("command panic")
)

// Helper methods for easily surfacing info in the TUI.
//
// TODO: leverage a cache to enhance performance, particularly if we introduce
// sqlite at some stage. These helpers are invoked on every render, which for a
// table with, say 40 visible rows, means they are invoked 40 times a render,
// which is 40 lookups.
type Helpers struct {
	AppService *services.AppService
}

// SafeCmd wraps a tea.Cmd with panic recovery to ensure terminal is properly reset
// even if the command panics
func SafeCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}

	return func() (msg tea.Msg) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic
				slog.Error("Recovered from panic in tea.Cmd",
					"error", r,
					"stack", string(debug.Stack()))

				// Set the error message that can be handled in the Update method
				msg = tui.ErrorMsg(fmt.Errorf("%w: %v", ErrCommandPanic, r))
			}
		}()

		// Execute the original command
		msg = cmd()
		return
	}
}
