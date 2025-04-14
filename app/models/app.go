package models

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fchastanet/shell-command-bookmarker/internal/components/tabs"
	"github.com/fchastanet/shell-command-bookmarker/internal/framework/focus"
)

type settings struct {
	keys keyMap
}

type styles struct {
	docStyle       lipgloss.Style
	windowStyle    lipgloss.Style
	highlightColor lipgloss.AdaptiveColor
}

type model struct {
	width         int
	height        int
	help          *help.Model
	TabsComponent *tabs.Tabs
	FocusManager  *focus.Manager
	settings      *settings
	styles        *styles
}

func NewAppModel(
	focusManager *focus.Manager,
) model {
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

	helpModel := help.New()
	m := model{
		width:         0,
		height:        0,
		help:          &helpModel,
		TabsComponent: tabsModel,
		FocusManager:  focusManager,
		settings: &settings{
			keys: defaultKeyMap(),
		},
		styles: styles,
	}
	return m
}

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keyMap struct {
	Up   key.Binding
	Down key.Binding
	Help key.Binding
	Quit key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (m model) ShortHelp() []key.Binding {
	return []key.Binding{m.settings.keys.Help, m.settings.keys.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (m model) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.settings.keys.Up, m.settings.keys.Down}, // first column
		m.FocusManager.GetKeyBindings(),
		m.TabsComponent.GetKeyBindings(),
		{m.settings.keys.Help, m.settings.keys.Quit}, // last column
	}
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Help: key.NewBinding(
			key.WithKeys("?", "h"),
			key.WithHelp("?/h", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q/Escape", "quit"),
		),
	}
}

func (m model) IsFocusable() bool {
	return true
}

func (m model) GetInnerFocusableComponents() []focus.Focusable {
	return []focus.Focusable{
		m.TabsComponent,
	}
}

func (m model) GetFocusableUniqueID() string {
	return "main"
}

func (m model) Init() tea.Cmd {
	// Initialize sub-models
	return tea.Batch(
		m.FocusManager.Init(),
		m.TabsComponent.Init(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// If we set a width on the help menu it can gracefully truncate
		// its view as needed.
		m.help.Width = msg.Width
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.settings.keys.Quit):
			cmds = append(cmds, tea.Quit)
		case key.Matches(msg, m.settings.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}
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

func (m model) View() string {
	doc := strings.Builder{}

	helpView := m.help.View(m)
	height := 0
	if m.help.ShowAll {
		height = strings.Count(helpView, "\n")
	} else {
		height = 2
	}

	renderedTabs := m.TabsComponent.View()
	tabsStr := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs)
	doc.WriteString(tabsStr)
	doc.WriteString(strings.Repeat("\n", height) + helpView)
	return m.styles.docStyle.Render(doc.String())
}
