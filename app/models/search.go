package models

import (
	"github.com/charmbracelet/bubbles/table"
	customTable "github.com/fchastanet/shell-command-bookmarker/internal/components/table"
)

// jscpd:ignore-start
//
//nolint:all
func SearchTableModel() customTable.Model {
	columns := []table.Column{
		{Title: "Rank", Width: 4},
		{Title: "City", Width: 10},
		{Title: "Country", Width: 10},
		{Title: "Population", Width: 10},
	}
	rows := []table.Row{}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)
	return *customTable.NewModel(&t)
}

// jscpd:ignore-end
