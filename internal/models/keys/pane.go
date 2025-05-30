package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type PaneNavigationKeyMap struct {
	SwitchBottomPane *key.Binding
	SwitchPaneBack   *key.Binding
	LeftPane         *key.Binding
	TopPane          *key.Binding
	BottomPane       *key.Binding
	ShrinkPaneHeight *key.Binding
	GrowPaneHeight   *key.Binding
	ShrinkPaneWidth  *key.Binding
	GrowPaneWidth    *key.Binding
}

// Navigation returns key bindings for navigation.
func GetPaneNavigationKeyMap() *PaneNavigationKeyMap {
	switchBottomPane := key.NewBinding(
		key.WithKeys("enter", "tab"),
		key.WithHelp("⏎/⭾", "bottom pane"),
	)
	switchPaneBack := key.NewBinding(
		key.WithKeys("esc", "shift+tab"),
		key.WithHelp("␛/Shift-⭾", "back to top pane"),
	)
	topPane := key.NewBinding(
		key.WithKeys("alt+1", "alt+&"),
		key.WithHelp("Alt+1/Alt+&", "top pane"),
	)
	bottomPane := key.NewBinding(
		key.WithKeys("alt+2", "alt+é"),
		key.WithHelp("Alt+2/Alt+é", "bottom right pane"),
	)
	leftPane := key.NewBinding(
		key.WithKeys("alt+3", "alt+\""),
		key.WithHelp("Alt+3/Alt+\"", "left pane"),
	)
	shrinkPaneHeight := key.NewBinding(
		key.WithKeys("alt+up"),
		key.WithHelp("Alt+⬆", "reduce height"),
	)
	growPaneHeight := key.NewBinding(
		key.WithKeys("alt+down"),
		key.WithHelp("Alt+⬇", "increase height"),
	)
	shrinkPaneWidth := key.NewBinding(
		key.WithKeys("alt+left"),
		key.WithHelp("Alt+⬅", "reduce width"),
	)
	growPaneWidth := key.NewBinding(
		key.WithKeys("alt+right"),
		key.WithHelp("Alt+⮕", "increase width"),
	)

	return &PaneNavigationKeyMap{
		SwitchBottomPane: &switchBottomPane,
		SwitchPaneBack:   &switchPaneBack,
		LeftPane:         &leftPane,
		TopPane:          &topPane,
		BottomPane:       &bottomPane,
		ShrinkPaneHeight: &shrinkPaneHeight,
		GrowPaneHeight:   &growPaneHeight,
		ShrinkPaneWidth:  &shrinkPaneWidth,
		GrowPaneWidth:    &growPaneWidth,
	}
}
