package sort

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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
