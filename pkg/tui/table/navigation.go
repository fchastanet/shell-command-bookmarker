package table

import (
	"github.com/charmbracelet/bubbles/key"
)

type Navigation struct {
	LineUp       *key.Binding
	LineDown     *key.Binding
	PageUp       *key.Binding
	PageDown     *key.Binding
	HalfPageUp   *key.Binding
	HalfPageDown *key.Binding
	GotoTop      *key.Binding
	GotoBottom   *key.Binding
}

// GetNavigation returns key bindings for navigation.
func GetDefaultNavigation() *Navigation {
	lineUp := key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	)
	lineDown := key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	)
	pageUp := key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("⇞", "page up"),
	)
	pageDown := key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("⇟", "page down"),
	)
	halfPageUp := key.NewBinding(
		key.WithKeys("ctrl+pgup"),
		key.WithHelp("Ctrl+⇞", "½ page up"),
	)
	halfPageDown := key.NewBinding(
		key.WithKeys("ctrl+pgdown"),
		key.WithHelp("Ctrl+⇟", "½ page down"),
	)
	gotoTop := key.NewBinding(
		key.WithKeys("home"),
		key.WithHelp("↖", "go to start"),
	)
	gotoBottom := key.NewBinding(
		key.WithKeys("end"),
		key.WithHelp("End", "go to end"),
	)

	return &Navigation{
		LineUp:       &lineUp,
		LineDown:     &lineDown,
		PageUp:       &pageUp,
		PageDown:     &pageDown,
		HalfPageUp:   &halfPageUp,
		HalfPageDown: &halfPageDown,
		GotoTop:      &gotoTop,
		GotoBottom:   &gotoBottom,
	}
}
