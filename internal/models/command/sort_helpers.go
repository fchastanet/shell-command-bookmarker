package command

import (
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/structure"
	"github.com/fchastanet/shell-command-bookmarker/internal/services/models"
	"github.com/fchastanet/shell-command-bookmarker/pkg/sort"
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

// compareBySortField compares two commands by the given field
func compareBySortField(i, j *models.Command, field structure.Field) int {
	switch field {
	case structure.FieldID:
		return sort.CompareInt(i.GetID(), j.GetID())
	case structure.FieldTitle:
		return strings.Compare(i.Title, j.Title)
	case structure.FieldScript:
		return strings.Compare(i.Script, j.Script)
	case structure.FieldStatus:
		return strings.Compare(string(i.Status), string(j.Status))
	case structure.FieldLintStatus:
		return strings.Compare(string(i.LintStatus), string(j.LintStatus))
	case structure.FieldCreationDate:
		return sort.CompareTime(i.CreationDatetime, j.CreationDatetime)
	case structure.FieldModificationDate:
		return sort.CompareTime(i.ModificationDatetime, j.ModificationDatetime)
	default:
		slog.Warn("Unknown sort field", "field", field)
		return 0
	}
}
