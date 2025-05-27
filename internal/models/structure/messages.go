package structure

import (
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
)

type Position int

const (
	// TopPane occupies the top area of the terminal.
	TopPane Position = iota
	// BottomPane occupies the bottom area of the terminal.
	BottomPane
	// LeftPane occupies the left side of the terminal.
	LeftPane
)

// NavigationMsg is an instruction to navigate to a page.
type NavigationMsg struct {
	Page         Page
	Position     Position
	DisableFocus bool
}

func NewNavigationMsg(kind resource.Kind, opts ...NavigateOption) NavigationMsg {
	msg := NavigationMsg{
		Page:         Page{Kind: kind, ID: 0},
		Position:     LeftPane, // Default to left pane
		DisableFocus: false,    // Default to enabling focus
	}
	for _, fn := range opts {
		fn(&msg)
	}
	return msg
}

type NavigateOption func(msg *NavigationMsg)

func WithPosition(position Position) NavigateOption {
	return func(msg *NavigationMsg) {
		msg.Position = position
	}
}

type FocusedPaneChangedMsg struct {
	From Position
	To   Position
}

// CommandSelectedForShellMsg is sent when a command is selected for pasting to shell
type CommandSelectedForShellMsg struct {
	Command string
}
