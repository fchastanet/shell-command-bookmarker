package models

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fchastanet/shell-command-bookmarker/internal/components/help"
	"github.com/fchastanet/shell-command-bookmarker/internal/components/tabs"
	"github.com/fchastanet/shell-command-bookmarker/internal/framework/focus"
)

type styles struct {
	docStyle       lipgloss.Style
	windowStyle    lipgloss.Style
	highlightColor lipgloss.AdaptiveColor
}

type AppModel struct {
	width  int
	height int
	styles *styles
	// sub components
	appHelpModel  *help.Model
	TabsComponent *tabs.Tabs
	FocusManager  *focus.Manager
}

func NewAppModel(
	focusManager *focus.Manager,
) AppModel {
	myTabs := []tabs.Tab{
		{
			Title: "Search",
			Model: SearchTableModel(),
		},
		{
			Title: "History",
			Model: BookmarksTableModel(),
		},
		{
			Title: "Bookmarks",
			Model: BookmarksTableModel(),
		},
	}
	styles := getDefaultStyles(nil)

	tabsModel := tabs.NewTabs(
		myTabs,
		focusManager,
		&styles.highlightColor,
		&styles.windowStyle,
	)
	appHelpModel := help.NewAppHelpModel(
		focusManager,
		&styles.docStyle,
	)

	m := AppModel{
		width:         0,
		height:        0,
		appHelpModel:  &appHelpModel,
		TabsComponent: tabsModel,
		FocusManager:  focusManager,
		styles:        styles,
	}

	shortHelp := func() []key.Binding {
		return m.appHelpModel.GetKeyBindings()
	}

	fullHelp := func() [][]key.Binding {
		helpKeyMap := m.appHelpModel.GetKeyBindings()
		return [][]key.Binding{
			helpKeyMap,
			m.FocusManager.GetKeyBindings(),
			m.TabsComponent.GetKeyBindings(),
		}
	}

	m.appHelpModel.SetShortHelpHandler(shortHelp)
	m.appHelpModel.SetFullHelpHandler(fullHelp)

	return m
}

func (m AppModel) IsFocusable() bool {
	return true
}

func (m AppModel) GetInnerFocusableComponents() []focus.Focusable {
	return []focus.Focusable{
		m.TabsComponent,
	}
}

func (m AppModel) GetFocusableUniqueID() string {
	return "main"
}

func (m AppModel) Init() tea.Cmd {
	// Initialize sub-models
	return tea.Batch(
		m.appHelpModel.Init(),
		m.FocusManager.Init(),
		m.TabsComponent.Init(),
	)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	// Update focus manager model
	focusManagerModel, cmd := m.FocusManager.Update(msg)
	focusModel := focusManagerModel.(focus.Manager)
	m.FocusManager = &focusModel
	cmds = append(cmds, cmd)

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

//nolint:mnd // no need to check magic numbers
func getDefaultStyles(
	highlightColor *lipgloss.AdaptiveColor,
) *styles {
	if highlightColor == nil {
		highlightColor = &lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	}
	return &styles{
		highlightColor: *highlightColor,
		docStyle:       lipgloss.NewStyle().Padding(1, 2, 1, 2),
		windowStyle: lipgloss.NewStyle().
			BorderForeground(highlightColor).
			Padding(2, 0).
			Align(lipgloss.Center).
			Border(lipgloss.NormalBorder()).
			UnsetBorderTop(),
	}
}

func (m AppModel) View() string {
	doc := strings.Builder{}

	renderedTabs := m.TabsComponent.View()
	tabsStr := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs)
	doc.WriteString(tabsStr)
	doc.WriteString(m.appHelpModel.View())
	return m.styles.docStyle.Render(doc.String())
}
