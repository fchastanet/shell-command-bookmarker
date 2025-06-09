package sort

import (
	"log/slog"
	"strings"
	"time"

	"github.com/fchastanet/shell-command-bookmarker/internal/services/models"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
)

// CommandSortFunc returns a sort function based on the current sort state
func CommandSortFunc(state *State) func(i, j *models.Command) int {
	return func(i, j *models.Command) int {
		// Primary sort
		primary := compareBySortField(i, j, state.PrimarySort.Field)
		if primary != 0 {
			// Apply sort direction
			if state.PrimarySort.Direction == DirectionDesc {
				return -primary
			}
			return primary
		}

		// If primary fields are equal and secondary sort is defined
		if state.SecondarySort != nil {
			secondary := compareBySortField(i, j, state.SecondarySort.Field)
			if secondary != 0 {
				// Apply sort direction
				if state.SecondarySort.Direction == DirectionDesc {
					return -secondary
				}
				return secondary
			}
		}

		// Fall back to ID comparison if everything else is equal
		return models.CommandSorter(i, j)
	}
}

// CommandSortFuncDynamic returns a sort function that dynamically gets the sort state
// from a getter function whenever sorting is performed. This ensures the sort always
// uses the current sort state rather than a cached state.
func CommandSortFuncDynamic(getState func() *State) func(i, j *models.Command) int {
	return func(i, j *models.Command) int {
		// Get the current state whenever sorting is performed
		state := getState()
		if state == nil {
			// Fall back to default sorting if no state is available
			return models.CommandSorter(i, j)
		}

		// Primary sort
		primary := compareBySortField(i, j, state.PrimarySort.Field)
		if primary != 0 {
			// Apply sort direction
			if state.PrimarySort.Direction == DirectionDesc {
				return -primary
			}
			return primary
		}

		// If primary fields are equal and secondary sort is defined
		if state.SecondarySort != nil {
			secondary := compareBySortField(i, j, state.SecondarySort.Field)
			if secondary != 0 {
				// Apply sort direction
				if state.SecondarySort.Direction == DirectionDesc {
					return -secondary
				}
				return secondary
			}
		}

		// Fall back to ID comparison if everything else is equal
		return models.CommandSorter(i, j)
	}
}

// compareBySortField compares two commands by the given field
func compareBySortField(i, j *models.Command, field Field) int {
	switch field {
	case FieldID:
		return compareInt(i.GetID(), j.GetID())
	case FieldTitle:
		return strings.Compare(i.Title, j.Title)
	case FieldScript:
		return strings.Compare(i.Script, j.Script)
	case FieldStatus:
		return strings.Compare(string(i.Status), string(j.Status))
	case FieldLintStatus:
		return strings.Compare(string(i.LintStatus), string(j.LintStatus))
	case FieldCreationDate:
		return compareTime(i.CreationDatetime, j.CreationDatetime)
	case FieldModificationDate:
		return compareTime(i.ModificationDatetime, j.ModificationDatetime)
	default:
		slog.Warn("Unknown sort field", "field", field)
		return 0
	}
}

func compareInt(i, j resource.ID) int {
	if i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

// compareTime compares two time values
func compareTime(t1, t2 time.Time) int {
	switch {
	case t1.Before(t2):
		return -1
	case t1.After(t2):
		return 1
	default:
		return 0
	}
}
