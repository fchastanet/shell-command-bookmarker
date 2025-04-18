package models

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/app/services"
	"github.com/fchastanet/shell-command-bookmarker/internal/framework/style"
)

// jscpd:ignore-start
//
//nolint:all

type HistoryTableModel struct {
	tableModel     table.Model
	styleManager   *style.Manager
	historyService *services.HistoryService
}

const (
	idColumnWidth     = 6
	titleColumnWidth  = 20
	scriptColumnWidth = 20
	statusColumnWidth = 10
	// rowsDisplayLimit is the number of rows to display in the table
	rowsDisplayLimit = 20
)

func NewHistoryTableModel(
	styleManager *style.Manager,
	historyService *services.HistoryService,
) *HistoryTableModel {
	columns := []table.Column{
		{Title: "Id", Width: idColumnWidth},
		{Title: "Title", Width: titleColumnWidth},
		{Title: "Script", Width: scriptColumnWidth},
		{Title: "Status", Width: statusColumnWidth},
	}
	rows := []table.Row{}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(rowsDisplayLimit),
	)
	t.Focus()
	t.SetRows(rows)
	return &HistoryTableModel{
		tableModel:     t,
		styleManager:   styleManager,
		historyService: historyService,
	}
}

func (m *HistoryTableModel) Init() tea.Cmd {
	return nil
}

func (m *HistoryTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	rows, err := m.historyService.GetHistoryRows()
	if err != nil {
		slog.Error("Error getting history rows", "error", err)
		return m, nil
	}
	m.tableModel.SetRows(rows)

	m.tableModel, cmd = m.tableModel.Update(msg)

	return m, cmd
}

func (m *HistoryTableModel) View() string {
	return m.tableModel.View()
}

// jscpd:ignore-end
