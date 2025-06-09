package category

import (
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/sort"
)

// Type defines the type of command category
type Type int

// Tab represents a command category tab
type Tab[
	CategoryType Type,
	CommandStatus any,
	ElementType resource.Identifiable,
	FieldType string,
] struct {
	FilterState  *FilterSortState[ElementType, FieldType]
	Type         CategoryType
	Title        string
	CommandTypes []CommandStatus
	Count        int
}

// FilterSortState holds the current filter value and sort state
type FilterSortState[ElementType resource.Identifiable, FieldType string] struct {
	SortState   *sort.State[ElementType, FieldType]
	FilterValue string
}
