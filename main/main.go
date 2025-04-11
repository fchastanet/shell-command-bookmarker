package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	keys       keyMap
	mouseEvent tea.MouseEvent
	lastKey    string
	width      int
	height     int
	help       help.Model
	Tabs       []string
	TabContent []string
	activeTab  int
}

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Help  key.Binding
	Quit  key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right}, // first column
		{k.Help, k.Quit},                // second column
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
	Left: key.NewBinding(
		key.WithKeys("left", "p", "shift+tab"),
		key.WithHelp("←/p/Shift-↔", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "n", "tab"),
		key.WithHelp("→/n/↔", "move right"),
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
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			return m, tea.Quit
		case key.Matches(msg, m.keys.Right):
			m.lastKey = m.keys.Right.Help().Key
			if m.activeTab == len(m.Tabs)-1 {
				m.activeTab = 0
			} else {
				m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
			}
		case key.Matches(msg, m.keys.Left):
			m.lastKey = m.keys.Left.Help().Key
			if m.activeTab == 0 {
				m.activeTab = len(m.Tabs) - 1
			} else {
				m.activeTab = max(m.activeTab-1, 0)
			}
		case key.Matches(msg, m.keys.Help):
			m.lastKey = m.keys.Help.Help().Key
			m.help.ShowAll = !m.help.ShowAll
		}
	}

	return m, nil
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

func (m model) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == m.activeTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		tabsCount := len(m.Tabs)
		style = style.Width(
			m.width/tabsCount -
				(style.GetBorderLeftSize() + style.GetBorderRightSize() + m.width%2),
		)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	helpView := m.help.View(m.keys)
	height := 0
	if m.help.ShowAll {
		height = strings.Count(helpView, "\n")
	} else {
		height = 2
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(m.TabContent[m.activeTab]))
	doc.WriteString(strings.Repeat("\n", height) + helpView)
	return docStyle.Render(doc.String())
}

func main() {
	tabs := []string{"Search", "History", "Bookmarks"}
	tabContent := []string{"Search Tab", "History Tab", "Bookmarks Tab"}
	m := model{
		keys:       keys,
		help:       help.New(),
		Tabs:       tabs,
		TabContent: tabContent,
	}
	if _, err := tea.NewProgram(m, tea.WithMouseAllMotion()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
