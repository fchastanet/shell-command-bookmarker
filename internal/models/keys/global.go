package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type GlobalKeyMap struct {
	Search           key.Binding
	Filter           key.Binding
	ShrinkPaneHeight key.Binding
	GrowPaneHeight   key.Binding
	ShrinkPaneWidth  key.Binding
	GrowPaneWidth    key.Binding
	ClosePane        key.Binding
	Quit             key.Binding
	Help             key.Binding
}

func GetGlobalKeyMap() *GlobalKeyMap {
	return &GlobalKeyMap{
		Search: key.NewBinding(
			key.WithKeys("ctrl+f", "f3", "ctrl+r"),
			key.WithHelp("ctrl+f/ctrl+r/F3", "search"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp(`/`, "filter"),
		),
		ShrinkPaneHeight: key.NewBinding(
			key.WithKeys("-"),
			key.WithHelp("-", "reduce height"),
		),
		GrowPaneHeight: key.NewBinding(
			key.WithKeys("+"),
			key.WithHelp("+", "increase height"),
		),
		ShrinkPaneWidth: key.NewBinding(
			key.WithKeys("<"),
			key.WithHelp("<", "reduce width"),
		),
		GrowPaneWidth: key.NewBinding(
			key.WithKeys(">"),
			key.WithHelp(">", "increase width"),
		),
		ClosePane: key.NewBinding(
			key.WithKeys("X"),
			key.WithHelp("X", "close pane"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "exit"),
		),
		Help: key.NewBinding(
			key.WithKeys("h", "?"),
			key.WithHelp("h/?", "close help"),
		),
	}
}
