package models

import (
	"github.com/charmbracelet/bubbles/table"
	customTable "github.com/fchastanet/shell-command-bookmarker/internal/components/table"
	"github.com/fchastanet/shell-command-bookmarker/internal/framework/style"
)

// jscpd:ignore-start
//
//nolint:all
func NewSearchTableModel(
	styleManager *style.Manager,
) customTable.Model {
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
	return *customTable.NewModel(&t, styleManager)
}

// jscpd:ignore-end
