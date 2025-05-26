package keys

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui/table"
)

type TableCustomActionKeyMap struct {
	ComposeCommand *key.Binding
}

func GetTableCustomActionKeyMap() *TableCustomActionKeyMap {
	composeCommand := key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "compose command"),
	)
	return &TableCustomActionKeyMap{
		ComposeCommand: &composeCommand,
	}
}

// Navigation returns key bindings for table.
func GetTableNavigationKeyMap() *table.Navigation {
	return table.GetDefaultNavigation()
}

func GetTableActionKeyMap() *table.Action {
	return table.GetDefaultAction()
}
