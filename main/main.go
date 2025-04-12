package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fchastanet/shell-command-bookmarker/components/tabs"
)

type model struct {
	keys          keyMap
	mouseEvent    tea.MouseEvent
	lastKey       string
	width         int
	height        int
	help          help.Model
	TabsComponent *tabs.Tabs
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
	return []key.Binding{m.keys.Help, m.keys.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (m model) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keys.Up, m.keys.Down}, // first column
		m.TabsComponent.GetKeyBindings(),
		{m.keys.Help, m.keys.Quit}, // second column
	}
}

var keys = keyMap{
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

func (m model) Init() tea.Cmd {
	return m.TabsComponent.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	tabsModel, tabsCmd := m.TabsComponent.Update(msg)
	tabsModelConverted := tabsModel.(tabs.Tabs)
	m.TabsComponent = &tabsModelConverted
	cmds = append(cmds, tabsCmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// If we set a width on the help menu it can gracefully truncate
		// its view as needed.
		m.help.Width = msg.Width
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			m.lastKey = m.keys.Up.Help().Key
		case key.Matches(msg, m.keys.Down):
			m.lastKey = m.keys.Down.Help().Key
		case key.Matches(msg, m.keys.Quit):
			m.lastKey = m.keys.Quit.Help().Key
			cmds = append(cmds, tea.Quit)
		case key.Matches(msg, m.keys.Help):
			m.lastKey = m.keys.Help.Help().Key
			m.help.ShowAll = !m.help.ShowAll
		}
	}

	return m, tea.Batch(cmds...)
}

var (
	docStyle       = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	windowStyle    = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

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
	tabs := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs)
	doc.WriteString(tabs)
	doc.WriteString(strings.Repeat("\n", height) + helpView)
	return docStyle.Render(doc.String())
}

func main() {
	tabsTitle := []string{"Search", "History", "Bookmarks"}
	tabContent := []string{"Search Tab", "History Tab", "Bookmarks Tab"}
	m := model{
		keys: keys,
		help: help.New(),
		TabsComponent: tabs.NewTabs(
			tabsTitle,
			tabContent,
			windowStyle,
			highlightColor,
		),
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
