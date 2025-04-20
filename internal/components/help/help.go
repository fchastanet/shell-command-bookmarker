package help

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/framework/style"
)

type Model struct {
	width            int
	height           int
	help             *help.Model
	shortHelpHandler func() []key.Binding
	fullHelpHandler  func() [][]key.Binding
	styleManager     *style.Manager
	settings         *settings
}

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keyMap struct {
	Help key.Binding
	Quit key.Binding
}

type settings struct {
	keys keyMap
}

type KeyBindingsHelpInterface interface {
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
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

func (m Model) GetKeyBindings() []key.Binding {
	return []key.Binding{
		m.settings.keys.Help,
		m.settings.keys.Quit,
	}
}

func NewAppHelpModel(
	styleManager *style.Manager,
) *Model {
	helpModel := help.New()
	m := Model{
		width:        0,
		height:       0,
		help:         &helpModel,
		styleManager: styleManager,
		shortHelpHandler: func() []key.Binding {
			return []key.Binding{}
		},
		fullHelpHandler: func() [][]key.Binding {
			return [][]key.Binding{}
		},
		settings: &settings{
			keys: defaultKeyMap(),
		},
	}
	return &m
}

func (m *Model) SetShortHelpHandler(
	shortHelpHandler func() []key.Binding,
) {
	m.shortHelpHandler = shortHelpHandler
}

func (m *Model) SetFullHelpHandler(
	fullHelpHandler func() [][]key.Binding,
) {
	m.fullHelpHandler = fullHelpHandler
}

func (m *Model) Init() tea.Cmd {
	// Initialize sub-models
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// If we set a width on the help menu it can gracefully truncate
		// its view as needed.
		m.help.Width = msg.Width
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.settings.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.settings.keys.Quit):
			cmds = append(cmds, tea.Quit)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) ShortHelp() []key.Binding {
	return m.shortHelpHandler()
}

func (m *Model) FullHelp() [][]key.Binding {
	return m.fullHelpHandler()
}

func (m *Model) View() string {
	doc := strings.Builder{}

	helpView := m.help.View(m)

	doc.WriteString(helpView)
	return m.styleManager.DocStyle.Render(doc.String())
}
