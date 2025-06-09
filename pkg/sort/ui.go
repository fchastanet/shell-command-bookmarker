package sort

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
)

func (state *State[ElementType, FieldType]) Init() tea.Cmd {
	return func() tea.Msg {
		return MsgSortEditModeChanged[ElementType, FieldType]{
			State: state,
		}
	}
}

func (state *State[ElementType, FieldType]) Update(msg tea.Msg) (cmd tea.Cmd, forward bool) {
	// Handle sorting key pressed when sort mode is active

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if state.IsEditActive {
			return state.handleSortEditModeKeyMsg(keyMsg), false
		} else if tui.CheckKey(keyMsg, state.KeyMap.Sort) {
			return state.handleActivateSort(), false
		}
	}

	return nil, true
}

// handleSortEditModeKeyMsg handles key pressed when in sort edit mode
func (state *State[ElementType, FieldType]) handleSortEditModeKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch {
	case tui.CheckKey(msg, state.KeyMap.Apply):
		return state.handleApplySort()

	case tui.CheckKey(msg, state.KeyMap.Cancel):
		return state.handleCancelSort()

	case tui.CheckKey(msg, state.KeyMap.NextField):
		state.handleTabNavigation(true)
		return nil

	case tui.CheckKey(msg, state.KeyMap.PreviousField):
		state.handleTabNavigation(false)
		return nil

	case tui.CheckKey(msg, state.KeyMap.NextComboValue):
		// When down is pressed, cycle through next combo options for the selected field
		state.handleNextComboOption(DirectionAsc)
		return nil

	case tui.CheckKey(msg, state.KeyMap.PreviousComboValue):
		// When up is pressed, cycle through previous combo options for the selected field
		state.handleNextComboOption(DirectionDesc)
		return nil
	}

	return nil
}

// handleTabNavigation handles tab and shift+tab navigation between sort options
func (state *State[ElementType, FieldType]) handleTabNavigation(
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

// handleActivateSort activates the sort mode
func (state *State[ElementType, FieldType]) handleActivateSort() tea.Cmd {
	state.IsEditActive = true
	state.SelectedField = 0 // Start with primary field selected

	// Return an info message
	return tea.Batch(
		func() tea.Msg {
			return tui.InfoMsg("Sort mode activated - use Tab/Shift+Tab to navigate, Enter to apply, Esc to cancel")
		},
		func() tea.Msg {
			return MsgSortEditModeChanged[ElementType, FieldType]{
				State: state,
			}
		},
	)
}

// handleApplySort applies the current sort settings
func (state *State[ElementType, FieldType]) handleApplySort() tea.Cmd {
	state.FixDuplicateSortKey()
	state.IsEditActive = false

	// Return a reload message with sort state included
	message := fmt.Sprintf("Applied sort: %s %s", state.PrimarySort.Field, state.PrimarySort.Direction)
	if state.SecondarySort != nil {
		message += fmt.Sprintf(", %s %s", state.SecondarySort.Field, state.SecondarySort.Direction)
	}

	infoMsg := tui.InfoMsg(message)
	return func() tea.Msg {
		return Msg[ElementType, FieldType]{
			State:   state,
			InfoMsg: &infoMsg,
		}
	}
}

// handleCancelSort cancels the current sort operation
func (state *State[ElementType, FieldType]) handleCancelSort() tea.Cmd {
	state.IsEditActive = false

	return tea.Batch(
		func() tea.Msg {
			return tui.InfoMsg("Sort operation cancelled")
		},
		func() tea.Msg {
			return MsgSortEditModeChanged[ElementType, FieldType]{
				State: state,
			}
		},
	)
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

// SelectNextOption cycles through the available options for the currently selected field
func (state *State[ElementType, FieldType]) handleNextComboOption(
	comboDirection Direction,
) {
	if !state.IsEditActive {
		return
	}

	switch state.SelectedField {
	case SelectedFieldPrimaryField:
		state.UpdateSortField(state.PrimarySort, comboDirection)

		// If primary field is changed from ID, initialize secondary sort
		if state.PrimarySort.Field != state.IDField && state.SecondarySort == nil {
			state.SecondarySort = &Option[FieldType]{
				Field:     state.IDField,
				Direction: DirectionAsc,
			}
		}
	case SelectedFieldPrimaryDirection:
		state.UpdateDirectionField(state.PrimarySort, comboDirection)

	case SelectedFieldSecondaryField:
		if state.SecondarySort == nil {
			return
		}
		state.UpdateSortField(state.SecondarySort, comboDirection)
	case SelectedFieldSecondaryDirection:
		if state.SecondarySort == nil {
			return
		}
		state.UpdateDirectionField(state.SecondarySort, comboDirection)
	}
}

func (state *State[ElementType, FieldType]) View() string {
	if state.IsEditActive {
		return state.renderStateEditMode()
	}

	return state.renderStateReadOnly()
}

func (state *State[ElementType, FieldType]) renderStateReadOnly() string {
	var sb strings.Builder
	sb.WriteString("Sort By: ")
	sb.WriteString(string(state.PrimarySort.Field))
	sb.WriteString(" ")
	sb.WriteString(string(state.PrimarySort.Direction))

	if state.SecondarySort != nil {
		sb.WriteString(" ")
		sb.WriteString(string(state.SecondarySort.Field))
		sb.WriteString(" ")
		sb.WriteString(string(state.SecondarySort.Direction))
	}

	return sb.String()
}

func (state *State[ElementType, FieldType]) renderStateEditMode() string {
	// Render active sort interface
	var sb strings.Builder
	sb.WriteString("Sort By: ")

	activeStyle := state.EditorSortStyles.GetActiveStyle()
	inactiveStyle := state.EditorSortStyles.GetInactiveStyle()

	// Primary field selector
	sb.WriteString(renderField(
		string(state.PrimarySort.Field),
		state.SelectedField == SelectedFieldPrimaryField,
		activeStyle,
		inactiveStyle,
	))

	// Primary direction selector
	sb.WriteString(renderField(
		string(state.PrimarySort.Direction),
		state.SelectedField == SelectedFieldPrimaryDirection,
		activeStyle,
		inactiveStyle,
	))

	// Secondary selectors (if primary is not ID)
	if state.SecondarySort != nil {
		sb.WriteString(renderField(
			string(state.SecondarySort.Field),
			state.SelectedField == SelectedFieldSecondaryField,
			activeStyle,
			inactiveStyle,
		))

		// Secondary direction selector
		sb.WriteString(renderField(
			string(state.SecondarySort.Direction),
			state.SelectedField == SelectedFieldSecondaryDirection,
			activeStyle,
			inactiveStyle,
		))
	}

	return sb.String()
}

func renderField(fieldText string, active bool, activeStyle, inactiveStyle *lipgloss.Style) string {
	if active {
		return activeStyle.Render("["+fieldText+"]") + " "
	}
	return inactiveStyle.Render("["+fieldText+"]") + " "
}
