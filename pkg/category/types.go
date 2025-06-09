package category

import (
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/sort"
)

// Type defines the type of command category
type Type int

const (
	// CategoryAll represents all commands
	CategoryAll Type = iota
	// CategoryFiltered represents filtered commands
	CategoryFiltered
	// CategoryFavorites represents favorite commands
	CategoryFavorites
	// CategoryTerminalHistory represents history commands
	CategoryTerminalHistory
)

// String returns the string representation of a category type
func (c Type) String() string {
	switch c {
	case CategoryAll:
		return "All"
	case CategoryFiltered:
		return "Filtered"
	case CategoryFavorites:
		return "Favorites"
	case CategoryTerminalHistory:
		return "History"
	default:
		return "Unknown"
	}
}

// Tab represents a command category tab
type Tab[CommandStatus any] struct {
	Type         Type
	Title        string
	Count        int
	CommandTypes []CommandStatus
	FilterState  FilterSortState
}

// AdapterInterface provides methods for category-specific operations
type AdapterInterface[V resource.Identifiable, CommandStatus any] interface {
	// GetCategoryTabs returns the list of category tabs
	GetCategoryTabs() []Tab[CommandStatus]
	// GetCategoryTabConfiguration returns the configuration for a specific category
	GetCategoryTabConfiguration(category Type) Tab[CommandStatus]
	// GetCategoryCounts returns the counts of commands in each category
	GetCategoryCounts() (map[Type]int, error)
	// FilterCategory filters the rows based on the selected category and filter value
	FilterCategory(rows []V, category Type, filterValue string) ([]V, int)
	// GetFilter returns the filter component for the category tabs
	GetFilter() interface{}
}

// FilterSortState holds the current filter value and sort state
type FilterSortState struct {
	FilterValue string
	SortState   *sort.State
}
