package models

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	customTable "github.com/fchastanet/shell-command-bookmarker/pkg/components/table"
)

// jscpd:ignore-start
//
//nolint:all

type HistoryTableModel struct {
	tableModel     customTable.Model
	styles         *styles.TableStyle
	historyService *services.HistoryService
	width          int
	height         int
}

type Styles interface {
	GetTableStyle() *styles.TableStyle
	GetTableContentStyle() *table.Styles
}

const (
	idColumnPercentWidth     = 3
	titleColumnPercentWidth  = 19
	scriptColumnPercentWidth = 74
	statusColumnPercentWidth = 7
	// rowsDisplayLimit is the number of rows to display in the table
	rowsDisplayLimit = 20
	percent          = 100
	sidesCount       = 2
)

func NewHistoryTableModel(
	myStyles *styles.TableStyle,
	historyService *services.HistoryService,
) *HistoryTableModel {
	t := table.New(
		table.WithColumns([]table.Column{}), // will be initialized in Init
		table.WithRows([]table.Row{}),       // will be initialized in Update
		table.WithFocused(false),
		table.WithHeight(rowsDisplayLimit),
	)
	t.Focus()
	return &HistoryTableModel{
		tableModel:     *customTable.NewModel(&t, myStyles),
		styles:         myStyles,
		historyService: historyService,
		width:          0,
		height:         0,
	}
}

func (m *HistoryTableModel) getColumns(width int) []table.Column {
	slog.Debug("getColumns", "width", width)
	const columnsCount = 4
	w := width -
		columnsCount*m.styles.Content.Cell.GetHorizontalPadding()*sidesCount
	return []table.Column{
		{Title: "Id", Width: idColumnPercentWidth * w / percent},
		{Title: "Title", Width: titleColumnPercentWidth * w / percent},
		{Title: "Script", Width: scriptColumnPercentWidth * w / percent},
		{Title: "Status", Width: statusColumnPercentWidth * w / percent},
	}
}

func (m *HistoryTableModel) Init() tea.Cmd {
	m.tableModel.SetColumns(m.getColumns(m.width))
	return nil
}

func (m *HistoryTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		tableFrameWidth := m.styles.Border.GetHorizontalFrameSize()
		m.width = msg.Width - tableFrameWidth
		m.tableModel.SetColumns(m.getColumns(m.width))
		m.tableModel.SetWidth(m.width)

		tableFrameHeight := m.styles.Border.GetVerticalFrameSize()
		m.height = msg.Height - tableFrameHeight
		m.tableModel.SetHeight(m.height)

	case tea.BlurMsg:
		m.tableModel.Blur()

	case tea.FocusMsg:
		rows, err := m.historyService.GetHistoryRows()
		if err != nil {
			slog.Error("Error getting history rows", "error", err)
			return m, nil
		}
		m.tableModel.SetRows(rows)
		m.tableModel.Focus()
	}
	_, cmd := m.tableModel.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *HistoryTableModel) View() string {
	return m.tableModel.View()
}

// jscpd:ignore-end
