package command

import (
	"strings"

	"github.com/fchastanet/shell-command-bookmarker/pkg/sort"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui/table"
)

// getSortFieldByColumnKey maps column keys to sort fields
func getSortFieldByColumnKey(columnKey table.ColumnKey) sort.Field {
	switch columnKey {
	case "id":
		return sort.FieldID
	case "title":
		return sort.FieldTitle
	case "script":
		return sort.FieldScript
	case "status":
		return sort.FieldStatus
	case "lintStatus":
		return sort.FieldLintStatus
	default:
		return ""
	}
}

// updateColumnHeadersWithSortIndicators updates table column headers to display sort indicators
func updateColumnHeadersWithSortIndicators(columns []table.Column, sortState *sort.State) []table.Column {
	if sortState == nil {
		return columns
	}

	// Make a copy of the columns to avoid modifying the originals
	updatedColumns := make([]table.Column, len(columns))
	copy(updatedColumns, columns)

	// Reset all column titles to their original values without indicators
	for i := range updatedColumns {
		// Remove any existing sort indicators
		updatedColumns[i].Title = strings.TrimSuffix(strings.TrimSuffix(
			strings.TrimSuffix(updatedColumns[i].Title, " 1▲"),
			" 1▼"), " 2▲")
		updatedColumns[i].Title = strings.TrimSuffix(updatedColumns[i].Title, " 2▼")
	}

	// Apply primary sort indicator
	primaryField := getSortFieldByColumnKey(getPrimaryColumnKey(sortState))
	for i := range updatedColumns {
		columnField := getSortFieldByColumnKey(updatedColumns[i].Key)
		if columnField == primaryField {
			updatedColumns[i].Title += " 1" + string(sortState.PrimarySort.Direction)
			break
		}
	}

	// Apply secondary sort indicator if applicable
	if sortState.SecondarySort != nil {
		secondaryField := getSortFieldByColumnKey(getSecondaryColumnKey(sortState))
		for i := range updatedColumns {
			columnField := getSortFieldByColumnKey(updatedColumns[i].Key)
			if columnField == secondaryField {
				updatedColumns[i].Title += " 2" + string(sortState.SecondarySort.Direction)
				break
			}
		}
	}

	return updatedColumns
}

// getPrimaryColumnKey returns the column key for the primary sort field
func getPrimaryColumnKey(sortState *sort.State) table.ColumnKey {
	switch sortState.PrimarySort.Field {
	case sort.FieldID:
		return "id"
	case sort.FieldTitle:
		return "title"
	case sort.FieldScript:
		return "script"
	case sort.FieldStatus:
		return "status"
	case sort.FieldLintStatus:
		return "lintStatus"
	default:
		return ""
	}
}

// getSecondaryColumnKey returns the column key for the secondary sort field
func getSecondaryColumnKey(sortState *sort.State) table.ColumnKey {
	if sortState.SecondarySort == nil {
		return ""
	}

	switch sortState.SecondarySort.Field {
	case sort.FieldID:
		return "id"
	case sort.FieldTitle:
		return "title"
	case sort.FieldScript:
		return "script"
	case sort.FieldStatus:
		return "status"
	case sort.FieldLintStatus:
		return "lintStatus"
	default:
		return ""
	}
}
