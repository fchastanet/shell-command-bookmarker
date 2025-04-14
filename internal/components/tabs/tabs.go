package tabs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/framework/focus"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type Tab struct {
	Title string
	Model tea.Model
}

type tabsSettings struct {
	Keys keyMap
}

type styles struct {
	windowStyle      *lipgloss.Style
	inactiveTabStyle lipgloss.Style
	activeTabStyle   lipgloss.Style
	docStyle         lipgloss.Style
}

type Tabs struct {
	Tabs         []Tab
	focusManager *focus.Manager
	settings     *tabsSettings
	styles       *styles
	width        int
	height       int
	activeTab    int
}

func defaultKeyMap() keyMap {
	return keyMap{
		Left: key.NewBinding(
			key.WithKeys("left", "p"),
			key.WithHelp("←/p", "move tab left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "n"),
			key.WithHelp("→/n", "move tab right"),
		),
	}
}

//nolint:mnd // no need to check magic numbers
func getDefaultStyles(
	windowStyle *lipgloss.Style,
	highlightColor *lipgloss.AdaptiveColor,
) styles {
	if highlightColor == nil {
		highlightColor = &lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	}
	inactiveTabBorder := tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder := tabBorderWithBottom("┘", " ", "└")
	inactiveTabStyle := lipgloss.NewStyle().
		Border(inactiveTabBorder, true).
		BorderForeground(highlightColor).
		Padding(0, 1)

	return styles{
		windowStyle:      windowStyle,
		inactiveTabStyle: inactiveTabStyle,
		activeTabStyle:   inactiveTabStyle.Border(activeTabBorder, true),
		docStyle:         lipgloss.NewStyle().Padding(1, 2, 1, 2),
	}
}

func NewTabs(
	tabs []Tab,
	focusManager *focus.Manager,
	highlightColor *lipgloss.AdaptiveColor,
	windowStyle *lipgloss.Style,
) *Tabs {
	styles := getDefaultStyles(windowStyle, highlightColor)
	return &Tabs{
		width:        0,
		height:       0,
		Tabs:         tabs,
		focusManager: focusManager,
		settings: &tabsSettings{
			Keys: defaultKeyMap(),
		},
		styles:    &styles,
		activeTab: 0,
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
		t.settings.Keys.Left, t.settings.Keys.Right,
	}
}

func (t Tabs) IsFocusable() bool {
	return true
}

func (t Tabs) GetInnerFocusableComponents() []focus.Focusable {
	return nil
}

func (t Tabs) GetFocusableUniqueID() string {
	return "tabs"
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func (t Tabs) Init() tea.Cmd {
	batches := make([]tea.Cmd, len(t.Tabs))
	for _, tab := range t.Tabs {
		if tab.Model == nil {
			continue
		}
		batches = append(batches, tab.Model.Init())
	}
	return tea.Batch(batches...)
}

func (t Tabs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
	case tea.KeyMsg:
		if !t.focusManager.IsTerminalFocused() {
			return t, nil
		}
		switch {
		case key.Matches(msg, t.settings.Keys.Right):
			if t.activeTab == len(t.Tabs)-1 {
				t.activeTab = 0
			} else {
				t.activeTab = min(t.activeTab+1, len(t.Tabs)-1)
			}
		case key.Matches(msg, t.settings.Keys.Left):
			if t.activeTab == 0 {
				t.activeTab = len(t.Tabs) - 1
			} else {
				t.activeTab = max(t.activeTab-1, 0)
			}
		}
	}

	return t, nil
}

func (tab Tab) getBorderStyle(
	styles *styles, isFirst bool, isLast bool, isActive bool,
) lipgloss.Style {
	var style lipgloss.Style
	if isActive {
		style = styles.activeTabStyle
	} else {
		style = styles.inactiveTabStyle
	}
	border := style.GetBorderStyle()
	switch {
	case isFirst && isActive:
		border.BottomLeft = "│"
	case isFirst && !isActive:
		border.BottomLeft = "├"
	case isLast && isActive:
		border.BottomRight = "│"
	case isLast && !isActive:
		border.BottomRight = "┤"
	}
	return style.Border(border)
}

func (tab Tab) View(
	styles *styles, tabsCount int, width int,
	isFirst bool, isLast bool, isActive bool,
) string {
	if tab.Model == nil {
		return ""
	}
	borderStyle := tab.getBorderStyle(
		styles, isFirst, isLast, isActive,
	)
	borderStyle = borderStyle.Width(
		width/tabsCount -
			(borderStyle.GetBorderLeftSize() + borderStyle.GetBorderRightSize() + width%2),
	)
	return borderStyle.Render(tab.Title)
}

func (t Tabs) View() string {
	doc := strings.Builder{}

	renderedTabs := make([]string, len(t.Tabs))

	for i, tab := range t.Tabs {
		if tab.Model == nil {
			continue
		}
		tabsCount := len(t.Tabs)
		isFirst, isLast, isActive := i == 0, i == len(t.Tabs)-1, i == t.activeTab
		renderedTabs[i] = tab.View(
			t.styles, tabsCount, t.width,
			isFirst, isLast, isActive,
		)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(
		t.styles.windowStyle.Width(
			lipgloss.Width(row) - t.styles.windowStyle.GetHorizontalFrameSize(),
		).Render(t.Tabs[t.activeTab].Model.View()),
	)
	return doc.String()
}
