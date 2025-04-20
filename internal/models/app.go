package models

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
)

type AppModel struct {
	width  int
	height int
	// debug
	loggerService *services.LoggerService
	// sub components
	historyModel *HistoryTableModel
	styles       *styles.Styles
}

func NewAppModel(
	historyService *services.HistoryService,
	loggerService *services.LoggerService,
) *AppModel {
	styles := styles.NewStyles()
	historyModel := NewHistoryTableModel(styles, historyService)
	m := AppModel{
		width:        0,
		height:       0,
		historyModel: historyModel,
		styles:       styles,
	}

	return &m
}

func (m *AppModel) Init() tea.Cmd {
	// Initialize sub-models
	return tea.Batch(
		m.historyModel.Init(),
	)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.loggerService.LogTeaMsg(msg)
	var cmds []tea.Cmd

	_, cmd := m.historyModel.Update(msg)
	cmds = append(cmds, cmd)

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height

		m.historyModel.height = m.height
		m.historyModel.width = m.width
	}

	return m, tea.Batch(cmds...)
}

func (m *AppModel) View() string {
	doc := strings.Builder{}

	doc.WriteString(m.historyModel.View())
	return m.styles.GetDocStyle().Render(doc.String())
}
