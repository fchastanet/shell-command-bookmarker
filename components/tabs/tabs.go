package tabs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type Tab struct {
	Title string
	Model tea.Model
}

type Tabs struct {
	Tabs            []Tab
	width           int
	height          int
	activeTab       int
	Keys            keyMap
	terminalFocused bool
	// styles
	windowStyle       lipgloss.Style
	highlightColor    lipgloss.AdaptiveColor
	inactiveTabBorder lipgloss.Border
	activeTabBorder   lipgloss.Border
	inactiveTabStyle  lipgloss.Style
	activeTabStyle    lipgloss.Style
	docStyle          lipgloss.Style
}

func NewTabs(
	tabs []Tab,
	windowStyle lipgloss.Style,
	highlightColor lipgloss.AdaptiveColor,
) *Tabs {
	inactiveTabBorder := tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder := tabBorderWithBottom("┘", " ", "└")
	inactiveTabStyle := lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	return &Tabs{
		Tabs:            tabs,
		windowStyle:     windowStyle,
		activeTab:       0,
		terminalFocused: false,
		Keys: keyMap{
			Left: key.NewBinding(
				key.WithKeys("left", "p"),
				key.WithHelp("←/p", "move tab left"),
			),
			Right: key.NewBinding(
				key.WithKeys("right", "n"),
				key.WithHelp("→/n", "move tab right"),
			),
		},
		inactiveTabBorder: inactiveTabBorder,
		activeTabBorder:   activeTabBorder,
		docStyle:          lipgloss.NewStyle().Padding(1, 2, 1, 2),
		highlightColor:    lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"},
		inactiveTabStyle:  inactiveTabStyle,
		activeTabStyle:    inactiveTabStyle.Border(activeTabBorder, true),
	}
}

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keyMap struct {
	Left  key.Binding
	Right key.Binding
}

func (t Tabs) GetKeyBindings() []key.Binding {
	return []key.Binding{
		t.Keys.Left, t.Keys.Right,
	}
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func (m Tabs) Init() tea.Cmd {
	var batches []tea.Cmd
	for _, tab := range m.Tabs {
		if tab.Model == nil {
			continue
		}
		batches = append(batches, tab.Model.Init())
	}
	return tea.Batch(batches...)
}

func (m Tabs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.FocusMsg:
		m.terminalFocused = true
	case tea.BlurMsg:
		m.terminalFocused = false
	case tea.KeyMsg:
		if !m.terminalFocused {
			return m, nil
		}
		switch {
		case key.Matches(msg, m.Keys.Right):
			if m.activeTab == len(m.Tabs)-1 {
				m.activeTab = 0
			} else {
				m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
			}
		case key.Matches(msg, m.Keys.Left):
			if m.activeTab == 0 {
				m.activeTab = len(m.Tabs) - 1
			} else {
				m.activeTab = max(m.activeTab-1, 0)
			}
		}
	}

	return m, nil
}

func (m Tabs) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	for i, tab := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == m.activeTab
		if isActive {
			style = m.activeTabStyle
		} else {
			style = m.inactiveTabStyle
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
		renderedTabs = append(renderedTabs, style.Render(tab.Title))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(m.windowStyle.Width((lipgloss.Width(row) - m.windowStyle.GetHorizontalFrameSize())).
		Render(m.Tabs[m.activeTab].Model.View()),
	)
	return doc.String()
}
