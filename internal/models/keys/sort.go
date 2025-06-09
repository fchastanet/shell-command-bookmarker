package keys

import (
	"github.com/fchastanet/shell-command-bookmarker/pkg/sort"
)

// GetSortKeyMap returns the key map for sort operations
func GetSortKeyMap() *sort.KeyMap {
	return sort.DefaultKeyMap()
}
