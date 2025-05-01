package keys

import (
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui/table"
)

// Navigation returns key bindings for table.
func GetTableNavigationKeyMap() *table.Navigation {
	return table.GetDefaultNavigation()
}

func GetTableActionKeyMap() *table.Action {
	return table.GetDefaultAction()
}
