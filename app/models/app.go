package models

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/app/services"
	"github.com/fchastanet/shell-command-bookmarker/internal/components/help"
	"github.com/fchastanet/shell-command-bookmarker/internal/components/tabs"
	"github.com/fchastanet/shell-command-bookmarker/internal/framework/style"
)

type AppModel struct {
	width  int
	height int
	// sub components
	appHelpModel  *help.Model
	TabsComponent *tabs.Tabs
	styleManager  *style.Manager
}

func NewAppModel(
	historyService *services.HistoryService,
) AppModel {
	styleManager := style.NewManager()

	myTabs := []tabs.Tab{
		{
			Title: "Search",
			Model: NewSearchTableModel(styleManager),
		},
		{
			Title: "History",
			Model: NewHistoryTableModel(styleManager, historyService),
		},
		{
			Title: "Bookmarks",
			Model: NewBookmarksTableModel(styleManager),
		},
	}

	tabsModel := tabs.NewTabs(
		myTabs,
		styleManager,
	)
	appHelpModel := help.NewAppHelpModel(
		styleManager,
	)

	m := AppModel{
		width:         0,
		height:        0,
		appHelpModel:  &appHelpModel,
		TabsComponent: tabsModel,
		styleManager:  styleManager,
	}

	shortHelp := func() []key.Binding {
		return m.appHelpModel.GetKeyBindings()
	}

	fullHelp := func() [][]key.Binding {
		helpKeyMap := m.appHelpModel.GetKeyBindings()
		return [][]key.Binding{
			helpKeyMap,
			m.TabsComponent.GetKeyBindings(),
		}
	}

	m.appHelpModel.SetShortHelpHandler(shortHelp)
	m.appHelpModel.SetFullHelpHandler(fullHelp)

	return m
}

func (m AppModel) Init() tea.Cmd {
	// Initialize sub-models
	return tea.Batch(
		m.appHelpModel.Init(),
		m.TabsComponent.Init(),
	)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		msg.Width -= m.styleManager.WindowStyle.GetHorizontalFrameSize()
		msg.Height -= m.styleManager.WindowStyle.GetVerticalFrameSize()
	}

	// Update tabs model
	tabsTeaModel, cmd := m.TabsComponent.Update(msg)
	tabsModel := tabsTeaModel.(tabs.Tabs)
	m.TabsComponent = &tabsModel
	cmds = append(cmds, cmd)

	// Update help model
	helpModel, cmd := m.appHelpModel.Update(msg)
	helpModelInstance := helpModel.(help.Model)
	m.appHelpModel = &helpModelInstance
	cmds = append(cmds, cmd)

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, tea.Batch(cmds...)
}

func (m AppModel) View() string {
	doc := strings.Builder{}

	renderedTabs := m.TabsComponent.View()
	doc.WriteString(renderedTabs)
	doc.WriteString(m.appHelpModel.View())
	return m.styleManager.DocStyle.Render(doc.String())
}
