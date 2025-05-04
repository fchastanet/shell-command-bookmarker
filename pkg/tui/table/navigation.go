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
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	)
	lineDown := key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
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
		key.WithHelp("ctrl+⇞", "½ page up"),
	)
	halfPageDown := key.NewBinding(
		key.WithKeys("ctrl+pgdown"),
		key.WithHelp("ctrl+⇟", "½ page down"),
	)
	gotoTop := key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("g/↖", "go to start"),
	)
	gotoBottom := key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("G/end", "go to end"),
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
