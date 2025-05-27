package keys

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui/table"
)

type TableCustomActionKeyMap struct {
	ComposeCommand  *key.Binding
	CopyToClipboard *key.Binding
	SelectForShell  *key.Binding
}

func GetTableCustomActionKeyMap(shellSelectionMode bool) *TableCustomActionKeyMap {
	composeCommand := key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "compose command"),
	)
	composeCommand.SetEnabled(!shellSelectionMode) // Disable compose command in shell selection mode
	copyToClipboard := key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "copy to clipboard"),
	)
	selectForShell := key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select for shell"),
	)
	selectForShell.SetEnabled(shellSelectionMode) // Enable select for shell only in shell selection mode
	return &TableCustomActionKeyMap{
		ComposeCommand:  &composeCommand,
		CopyToClipboard: &copyToClipboard,
		SelectForShell:  &selectForShell,
	}
}

// Navigation returns key bindings for table.
func GetTableNavigationKeyMap() *table.Navigation {
	return table.GetDefaultNavigation()
}

func GetTableActionKeyMap(shellSelectionMode bool) *table.Action {
	tableKeyBindings := table.GetDefaultAction()
	if shellSelectionMode {
		tableKeyBindings.Select.SetEnabled(false)
		tableKeyBindings.SelectAll.SetEnabled(false)
		tableKeyBindings.SelectClear.SetEnabled(false)
		tableKeyBindings.SelectRange.SetEnabled(false)
		tableKeyBindings.Enter.SetEnabled(false)
		tableKeyBindings.Delete.SetEnabled(false)
	}
	return tableKeyBindings
}
