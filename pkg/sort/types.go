package sort

import (
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
)

// Field represents a field that can be sorted
type Field string

// Direction represents the direction of sorting
type Direction string

const (
	// Sort directions
	DirectionAsc  Direction = "▲"
	DirectionDesc Direction = "▼"
)

// Option represents a sort option (field and direction)
type Option[FieldType string] struct {
	Field     FieldType
	Direction Direction
}

// State contains the current sort configuration
type State[ElementType resource.Identifiable, FieldType string] struct {
	EditorSortStyles   EditorSortStyles
	IDField            FieldType
	PrimarySort        *Option[FieldType]
	SecondarySort      *Option[FieldType]
	CompareBySortField CompareBySortFieldFunc[ElementType, FieldType]
	Fields             []FieldType
	SelectedField      SelectedField // Currently selected field
	IsEditActive       bool          // Whether sort edit mode is active
}

type CompareBySortFieldFunc[ElementType resource.Identifiable, FieldType string] func(
	i, j ElementType, field FieldType,
) int

type SelectedField int

const (
	SelectedFieldPrimaryField       SelectedField = iota // Primary sort field
	SelectedFieldPrimaryDirection                        // Primary sort direction
	SelectedFieldSecondaryField                          // Secondary sort field
	SelectedFieldSecondaryDirection                      // Secondary sort direction
)

// Msg is sent when sorting is applied
type Msg[ElementType resource.Identifiable, FieldType string] struct {
	State   *State[ElementType, FieldType]
	InfoMsg *tui.InfoMsg
}

// NewDefaultState creates a new default sort state
func NewDefaultState[ElementType resource.Identifiable, FieldType string](
	editorStyles EditorSortStyles,
	idField FieldType,
	fields []FieldType,
	compareBySortFieldFunc CompareBySortFieldFunc[ElementType, FieldType],
) *State[ElementType, FieldType] {
	return &State[ElementType, FieldType]{
		PrimarySort: &Option[FieldType]{
			Field:     idField,
			Direction: DirectionAsc,
		},
		SecondarySort:      nil,
		IsEditActive:       false,
		SelectedField:      0,
		EditorSortStyles:   editorStyles,
		Fields:             fields,
		IDField:            idField,
		CompareBySortField: compareBySortFieldFunc,
	}
}

// GetDirectionOptions returns the available sort direction options
func GetDirectionOptions() []Direction {
	return []Direction{
		DirectionAsc,
		DirectionDesc,
	}
}

func (state *State[ElementType, FieldType]) FixDuplicateSortKey() {
	if state.SecondarySort != nil && state.PrimarySort.Field == state.SecondarySort.Field {
		state.SecondarySort = nil
	}

	// Secondary selector (if primary is not ID)
	if state.PrimarySort.Field != state.IDField {
		// Initialize secondary sort if needed
		if state.SecondarySort == nil {
			state.SecondarySort = &Option[FieldType]{
				Field:     state.IDField,
				Direction: DirectionAsc,
			}
		}
	}
}

// HandleTabNavigation handles tab and shift+tab navigation between sort options
func HandleTabNavigation[ElementType resource.Identifiable, FieldType string](
	state *State[ElementType, FieldType],
	forward bool,
) {
	if !state.IsEditActive {
		return
	}

	maxFields := 2 // Primary field and direction
	if state.PrimarySort.Field != state.IDField && state.SecondarySort != nil {
		maxFields = 4 // Include secondary field and direction
	}

	if forward {
		// Tab - move to next field
		state.SelectedField = SelectedField((int(state.SelectedField) + 1) % maxFields)
	} else {
		// Shift+Tab - move to previous field
		state.SelectedField = SelectedField((int(state.SelectedField) - 1 + maxFields) % maxFields)
	}
}

func (state *State[ElementType, FieldType]) UpdateSortField(
	option *Option[FieldType],
	direction Direction,
) {
	for i, field := range state.Fields {
		if field == option.Field {
			newIndex := computeNextIndex(i, len(state.Fields), direction)
			option.Field = state.Fields[newIndex]
			break
		}
	}
}

func (state *State[ElementType, FieldType]) UpdateDirectionField(
	option *Option[FieldType],
	direction Direction,
) {
	directions := GetDirectionOptions()
	for i, dir := range directions {
		if dir == option.Direction {
			newIndex := computeNextIndex(i, len(directions), direction)
			option.Direction = directions[newIndex]
			break
		}
	}
}

func computeNextIndex(index int, fieldsLength int, comboDirection Direction) int {
	comboDir := -1
	if comboDirection == DirectionAsc {
		comboDir = 1
	}

	// Handle negative index properly when cycling
	newIndex := (index + comboDir) % fieldsLength
	if newIndex < 0 {
		newIndex += fieldsLength
	}
	return newIndex
}
