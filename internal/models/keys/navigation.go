package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type navigation struct {
	SwitchPane      key.Binding
	SwitchPaneBack  key.Binding
	LeftPane        key.Binding
	TopRightPane    key.Binding
	BottomRightPane key.Binding
}

// Navigation returns key bindings for navigation.
var Navigation = navigation{
	SwitchPane: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next pane"),
	),
	SwitchPaneBack: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "last pane"),
	),
	LeftPane: key.NewBinding(
		key.WithKeys("0"),
		key.WithHelp("0", "left pane"),
	),
	TopRightPane: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "top right pane"),
	),
	BottomRightPane: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "bottom right pane"),
	),
}
