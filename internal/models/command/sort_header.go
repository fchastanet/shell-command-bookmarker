package command

import (
	"strings"

	dbmodels "github.com/fchastanet/shell-command-bookmarker/internal/services/models"
	"github.com/fchastanet/shell-command-bookmarker/pkg/sort"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui/table"
)

// updateColumnHeadersWithSortIndicators updates table column headers to display sort indicators
func updateColumnHeadersWithSortIndicators(
	columns []table.Column,
	sortState *sort.State[*dbmodels.Command, string],
) []table.Column {
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
	primaryField := table.ColumnKey(sortState.PrimarySort.Field)
	for i := range updatedColumns {
		columnField := updatedColumns[i].Key
		if columnField == primaryField {
			updatedColumns[i].Title += " 1" + string(sortState.PrimarySort.Direction)
			break
		}
	}

	// Apply secondary sort indicator if applicable
	if sortState.SecondarySort != nil {
		secondaryField := table.ColumnKey(sortState.SecondarySort.Field)
		for i := range updatedColumns {
			columnField := updatedColumns[i].Key
			if columnField == secondaryField {
				updatedColumns[i].Title += " 2" + string(sortState.SecondarySort.Direction)
				break
			}
		}
	}

	return updatedColumns
}
