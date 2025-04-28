package table

import (
	"github.com/charmbracelet/bubbles/key"
)

type Navigation struct {
	LineUp          *key.Binding
	LineDown        *key.Binding
	PageUp          *key.Binding
	PageDown        *key.Binding
	HalfPageUp      *key.Binding
	HalfPageDown    *key.Binding
	GotoTop         *key.Binding
	GotoBottom      *key.Binding
	SwitchPane      *key.Binding
	SwitchPaneBack  *key.Binding
	LeftPane        *key.Binding
	TopRightPane    *key.Binding
	BottomRightPane *key.Binding
	Select          *key.Binding
	SelectAll       *key.Binding
	SelectClear     *key.Binding
	SelectRange     *key.Binding
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
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "½ page up"),
	)
	halfPageDown := key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "½ page down"),
	)
	gotoTop := key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("g/home", "go to start"),
	)
	gotoBottom := key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("G/end", "go to end"),
	)
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

	return &Navigation{
		LineUp:          &lineUp,
		LineDown:        &lineDown,
		PageUp:          &pageUp,
		PageDown:        &pageDown,
		HalfPageUp:      &halfPageUp,
		HalfPageDown:    &halfPageDown,
		GotoTop:         &gotoTop,
		GotoBottom:      &gotoBottom,
		SwitchPane:      &switchPane,
		SwitchPaneBack:  &switchPaneBack,
		LeftPane:        &leftPane,
		TopRightPane:    &topRightPane,
		BottomRightPane: &bottomRightPane,
		Select:          &selectKey,
		SelectAll:       &selectAll,
		SelectClear:     &selectClear,
		SelectRange:     &selectRange,
	}
}
