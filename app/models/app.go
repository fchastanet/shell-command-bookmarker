package models

import (
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/davecgh/go-spew/spew"
	"github.com/fchastanet/shell-command-bookmarker/app/services"
	"github.com/fchastanet/shell-command-bookmarker/internal/components/help"
	"github.com/fchastanet/shell-command-bookmarker/internal/framework/style"
)

type AppModel struct {
	width  int
	height int
	// debug
	dump io.Writer
	// sub components
	appHelpModel *help.Model
	historyModel *HistoryTableModel
	styleManager *style.Manager
}

func NewAppModel(
	historyService *services.HistoryService,
	dump io.Writer,
) *AppModel {
	styleManager := style.NewManager()
	appHelpModel := help.NewAppHelpModel(
		styleManager,
	)
	historyModel := NewHistoryTableModel(styleManager, historyService)
	m := AppModel{
		width:        0,
		height:       0,
		dump:         dump,
		appHelpModel: appHelpModel,
		historyModel: historyModel,
		styleManager: styleManager,
	}

	shortHelp := func() []key.Binding {
		return m.appHelpModel.GetKeyBindings()
	}

	fullHelp := func() [][]key.Binding {
		helpKeyMap := m.appHelpModel.GetKeyBindings()
		return [][]key.Binding{
			helpKeyMap,
		}
	}

	m.appHelpModel.SetShortHelpHandler(shortHelp)
	m.appHelpModel.SetFullHelpHandler(fullHelp)

	return &m
}

func (m *AppModel) Init() tea.Cmd {
	// Initialize sub-models
	return tea.Batch(
		m.appHelpModel.Init(),
		m.historyModel.Init(),
	)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}
	var cmds []tea.Cmd

	// Update help model
	_, cmd := m.appHelpModel.Update(msg)
	cmds = append(cmds, cmd)

	_, cmd = m.historyModel.Update(msg)
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
	doc.WriteString(m.appHelpModel.View())
	return m.styleManager.DocStyle.Render(doc.String())
}
