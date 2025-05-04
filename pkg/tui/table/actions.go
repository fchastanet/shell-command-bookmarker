package table

import "github.com/charmbracelet/bubbles/key"

type Action struct {
	Select      *key.Binding
	SelectAll   *key.Binding
	SelectClear *key.Binding
	SelectRange *key.Binding
	Filter      *key.Binding
	Reload      *key.Binding
	Enter       *key.Binding
	Delete      *key.Binding
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
		key.WithKeys("ctrl+r", "f5"),
		key.WithHelp("f5/ctrl+r", "reload"),
	)
	enter := key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "view resource"),
	)
	deleteKey := key.NewBinding(
		key.WithKeys("delete"),
		key.WithHelp("delete", "delete row"),
	)

	return &Action{
		Select:      &selectKey,
		SelectAll:   &selectAll,
		SelectClear: &selectClear,
		SelectRange: &selectRange,
		Filter:      &filter,
		Reload:      &reload,
		Enter:       &enter,
		Delete:      &deleteKey,
	}
}
