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

type Action struct {
	Select      *key.Binding
	SelectAll   *key.Binding
	SelectClear *key.Binding
	SelectRange *key.Binding
	Filter      *key.Binding
	Reload      *key.Binding
	Enter       *key.Binding
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
		key.WithHelp("pgup", "page up"),
	)
	pageDown := key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdn", "page down"),
	)
	halfPageUp := key.NewBinding(
		key.WithKeys("ctrl+pgup"),
		key.WithHelp("ctrl+pgup", "½ page up"),
	)
	halfPageDown := key.NewBinding(
		key.WithKeys("ctrl+pgdown"),
		key.WithHelp("ctrl+pgdown", "½ page down"),
	)
	gotoTop := key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("g/home", "go to start"),
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

func GetDefaultAction() *Action {
	selectKey := key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("<space>", "select"),
	)
	selectAll := key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("ctrl+a", "select all"),
	)
	selectClear := key.NewBinding(
		key.WithKeys(`ctrl+\`),
		key.WithHelp(`ctrl+\`, "clear selection"),
	)
	selectRange := key.NewBinding(
		key.WithKeys(`ctrl+@`),
		key.WithHelp(`ctrl+<space>`, "select range"),
	)
	filter := key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp(`/`, "filter"),
	)
	reload := key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload"),
	)
	enter := key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "view resource"),
	)

	return &Action{
		Select:      &selectKey,
		SelectAll:   &selectAll,
		SelectClear: &selectClear,
		SelectRange: &selectRange,
		Filter:      &filter,
		Reload:      &reload,
		Enter:       &enter,
	}
}
