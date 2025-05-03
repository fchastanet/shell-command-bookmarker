package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type PaneNavigationKeyMap struct {
	SwitchPane       *key.Binding
	SwitchPaneBack   *key.Binding
	LeftPane         *key.Binding
	TopPane          *key.Binding
	BottomPane       *key.Binding
	ShrinkPaneHeight *key.Binding
	GrowPaneHeight   *key.Binding
	ShrinkPaneWidth  *key.Binding
	GrowPaneWidth    *key.Binding
	ClosePane        *key.Binding
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
	topPane := key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "top pane"),
	)
	bottomPane := key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "bottom right pane"),
	)
	closePane := key.NewBinding(
		key.WithKeys("X"),
		key.WithHelp("X", "close pane"),
	)
	shrinkPaneHeight := key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "reduce height"),
	)
	growPaneHeight := key.NewBinding(
		key.WithKeys("+"),
		key.WithHelp("+", "increase height"),
	)
	shrinkPaneWidth := key.NewBinding(
		key.WithKeys("<"),
		key.WithHelp("<", "reduce width"),
	)
	growPaneWidth := key.NewBinding(
		key.WithKeys(">"),
		key.WithHelp(">", "increase width"),
	)

	return &PaneNavigationKeyMap{
		SwitchPane:       &switchPane,
		SwitchPaneBack:   &switchPaneBack,
		LeftPane:         &leftPane,
		TopPane:          &topPane,
		BottomPane:       &bottomPane,
		ShrinkPaneHeight: &shrinkPaneHeight,
		GrowPaneHeight:   &growPaneHeight,
		ShrinkPaneWidth:  &shrinkPaneWidth,
		GrowPaneWidth:    &growPaneWidth,
		ClosePane:        &closePane,
	}
}
