package command

import (
	"github.com/charmbracelet/bubbles/key"
)

// IsSortModeActive returns true if the sort mode is currently active
func (m *commandsList) IsSortModeActive() bool {
	sortState := m.categoryTabs.GetActiveSortState()
	return sortState != nil && sortState.IsEditActive
}

// GetSortKeyBindings returns the key bindings for sort mode
func (m *commandsList) GetSortKeyBindings() []*key.Binding {
	return []*key.Binding{
		m.sortKeyMap.Apply,
		m.sortKeyMap.Cancel,
		m.sortKeyMap.PreviousField,
		m.sortKeyMap.NextField,
	}
}
