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
		key.WithHelp("shift+tab", "previous pane"),
	)
	topPane := key.NewBinding(
		key.WithKeys("alt+1", "alt+&"),
		key.WithHelp("alt+1/alt+&", "top pane"),
	)
	bottomPane := key.NewBinding(
		key.WithKeys("alt+2", "alt+é"),
		key.WithHelp("alt+2/alt+é", "bottom right pane"),
	)
	leftPane := key.NewBinding(
		key.WithKeys("alt+3", "alt+\""),
		key.WithHelp("alt+3/alt+\"", "left pane"),
	)
	closePane := key.NewBinding(
		key.WithKeys("alt+X", "alt+x"),
		key.WithHelp("alt+x", "close pane"),
	)
	shrinkPaneHeight := key.NewBinding(
		key.WithKeys("alt+up"),
		key.WithHelp("alt+⬆", "reduce height"),
	)
	growPaneHeight := key.NewBinding(
		key.WithKeys("alt+down"),
		key.WithHelp("alt+⬇", "increase height"),
	)
	shrinkPaneWidth := key.NewBinding(
		key.WithKeys("alt+left"),
		key.WithHelp("alt+⬅", "reduce width"),
	)
	growPaneWidth := key.NewBinding(
		key.WithKeys("alt+right"),
		key.WithHelp("alt+⮕", "increase width"),
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
