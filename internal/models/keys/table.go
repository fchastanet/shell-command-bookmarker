package keys

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui/table"
)

type TableCustomActionKeyMap struct {
	ComposeCommand  *key.Binding
	CopyToClipboard *key.Binding
}

func GetTableCustomActionKeyMap() *TableCustomActionKeyMap {
	composeCommand := key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "compose command"),
	)
	copyToClipboard := key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "copy to clipboard"),
	)
	return &TableCustomActionKeyMap{
		ComposeCommand:  &composeCommand,
		CopyToClipboard: &copyToClipboard,
	}
}

// Navigation returns key bindings for table.
func GetTableNavigationKeyMap() *table.Navigation {
	return table.GetDefaultNavigation()
}

func GetTableActionKeyMap() *table.Action {
	return table.GetDefaultAction()
}
