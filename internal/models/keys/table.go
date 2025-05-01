package keys

import (
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui/table"
)

// Navigation returns key bindings for table.
func GetTableKeyMap() *table.Navigation {
	return table.GetDefaultNavigation()
}
