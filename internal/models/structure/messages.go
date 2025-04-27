package structure

import (
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
)

type Position int

const (
	// TopRightPane occupies the top right area of the terminal. Mutually
	// exclusive with RightPane.
	TopRightPane Position = iota
	// BottomRightPane occupies the bottom right area of the terminal. Mutually
	// exclusive with RightPane.
	BottomRightPane
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
