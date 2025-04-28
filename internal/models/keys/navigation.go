package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type PaneNavigationKeyMap struct {
	SwitchPane      *key.Binding
	SwitchPaneBack  *key.Binding
	LeftPane        *key.Binding
	TopRightPane    *key.Binding
	BottomRightPane *key.Binding
}

// Navigation returns key bindings for navigation.
func GetPaneNavigationKeyMap() *PaneNavigationKeyMap {
	switchPane := key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next pane"),
	)
	switchPaneBack := key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "last pane"),
	)
	leftPane := key.NewBinding(
		key.WithKeys("0"),
		key.WithHelp("0", "left pane"),
	)
	topRightPane := key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "top right pane"),
	)
	bottomRightPane := key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "bottom right pane"),
	)

	return &PaneNavigationKeyMap{
		SwitchPane:      &switchPane,
		SwitchPaneBack:  &switchPaneBack,
		LeftPane:        &leftPane,
		TopRightPane:    &topRightPane,
		BottomRightPane: &bottomRightPane,
	}
}
