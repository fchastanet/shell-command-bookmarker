package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type global struct {
	Select           key.Binding
	SelectAll        key.Binding
	SelectClear      key.Binding
	SelectRange      key.Binding
	Filter           key.Binding
	ShrinkPaneHeight key.Binding
	GrowPaneHeight   key.Binding
	ShrinkPaneWidth  key.Binding
	GrowPaneWidth    key.Binding
	ClosePane        key.Binding
	Quit             key.Binding
	Help             key.Binding
}

var Global = global{
	Select: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("<space>", "select"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("ctrl+a", "select all"),
	),
	SelectClear: key.NewBinding(
		key.WithKeys(`ctrl+\`),
		key.WithHelp(`ctrl+\`, "clear selection"),
	),
	SelectRange: key.NewBinding(
		key.WithKeys(`ctrl+@`),
		key.WithHelp(`ctrl+<space>`, "select range"),
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
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "exit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "close help"),
	),
}
