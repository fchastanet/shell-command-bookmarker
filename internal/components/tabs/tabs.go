package tabs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/framework/focus"
	"github.com/fchastanet/shell-command-bookmarker/internal/framework/style"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type Tab struct {
	Title     string
	Model     tea.Model
	tabsModel *Tabs
}

type tabsSettings struct {
	Keys keyMap
}

type Tabs struct {
	Tabs         []Tab
	focusManager *focus.Manager
	styleManager *style.Manager
	settings     *tabsSettings
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

func NewTabs(
	tabs []Tab,
	focusManager *focus.Manager,
	styleManager *style.Manager,
) *Tabs {
	tabsModel := &Tabs{
		width:        0,
		height:       0,
		Tabs:         tabs,
		focusManager: focusManager,
		settings: &tabsSettings{
			Keys: defaultKeyMap(),
		},
		styleManager: styleManager,
		activeTab:    0,
	}

	// set tabs parent
	for i := range tabs {
		tabs[i].tabsModel = tabsModel
	}
	return tabsModel
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
		t.Tabs[t.activeTab].Model.Update(msg)
	}

	return t, nil
}

func (tab Tab) View(
	tabsCount int, width int,
	isFirst bool, isLast bool, isActive bool,
) string {
	if tab.Model == nil {
		return ""
	}
	borderStyle := tab.tabsModel.styleManager.GetTabBorderStyle(
		isFirst, isLast, isActive,
		width, tabsCount,
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
			tabsCount, t.width,
			isFirst, isLast, isActive,
		)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(
		t.styleManager.WindowStyle.Width(
			lipgloss.Width(row) - t.styleManager.WindowStyle.GetHorizontalFrameSize(),
		).Render(t.Tabs[t.activeTab].Model.View()),
	)
	return doc.String()
}
