package sort

import (
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
)

// Field represents a field that can be sorted
type Field string

// Direction represents the direction of sorting
type Direction string

const (
	// Sort fields
	FieldID               Field = "ID"
	FieldTitle            Field = "Title"
	FieldScript           Field = "Script"
	FieldStatus           Field = "Status"
	FieldLintStatus       Field = "Lint Status"
	FieldCreationDate     Field = "Creation Date"
	FieldModificationDate Field = "Modification Date"

	// Sort directions
	DirectionAsc  Direction = "▲"
	DirectionDesc Direction = "▼"
)

// Option represents a sort option (field and direction)
type Option struct {
	Field     Field
	Direction Direction
}

// State contains the current sort configuration
type State struct {
	PrimarySort      *Option
	SecondarySort    *Option
	IsEditActive     bool          // Whether sort edit mode is active
	SelectedField    SelectedField // Currently selected field
	EditorSortStyles EditorSortStyles
}

type SelectedField int

const (
	SelectedFieldPrimaryField       SelectedField = iota // Primary sort field
	SelectedFieldPrimaryDirection                        // Primary sort direction
	SelectedFieldSecondaryField                          // Secondary sort field
	SelectedFieldSecondaryDirection                      // Secondary sort direction
)

// Msg is sent when sorting is applied
type Msg struct {
	State   *State
	InfoMsg *tui.InfoMsg
}

// NewDefaultState creates a new default sort state
func NewDefaultState(editorStyles EditorSortStyles) *State {
	return &State{
		PrimarySort: &Option{
			Field:     FieldID,
			Direction: DirectionAsc,
		},
		SecondarySort:    nil,
		IsEditActive:     false,
		SelectedField:    0,
		EditorSortStyles: editorStyles,
	}
}

// GetFieldOptions returns the available sort field options
func GetFieldOptions() []Field {
	return []Field{
		FieldID,
		FieldTitle,
		FieldScript,
		FieldStatus,
		FieldLintStatus,
		FieldCreationDate,
		FieldModificationDate,
	}
}

// GetDirectionOptions returns the available sort direction options
func GetDirectionOptions() []Direction {
	return []Direction{
		DirectionAsc,
		DirectionDesc,
	}
}

func (state *State) FixDuplicateSortKey() {
	if state.PrimarySort.Field == state.SecondarySort.Field {
		state.SecondarySort = nil
	}

	// Secondary selector (if primary is not ID)
	if state.PrimarySort.Field != FieldID {
		// Initialize secondary sort if needed
		if state.SecondarySort == nil {
			state.SecondarySort = &Option{
				Field:     FieldID,
				Direction: DirectionAsc,
			}
		}
	}
}

// HandleTabNavigation handles tab and shift+tab navigation between sort options
func HandleTabNavigation(state *State, forward bool) {
	if !state.IsEditActive {
		return
	}

	maxFields := 2 // Primary field and direction
	if state.PrimarySort.Field != FieldID && state.SecondarySort != nil {
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

func (state *State) UpdateSortField(option *Option, direction Direction) {
	fields := GetFieldOptions()
	for i, field := range fields {
		if field == option.Field {
			newIndex := computeNextIndex(i, len(fields), direction)
			option.Field = fields[newIndex]
			break
		}
	}
}

func (state *State) UpdateDirectionField(option *Option, direction Direction) {
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
