package customTable

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TableModel struct {
	table      table.Model
	Keys       keyMap
	tableStyle lipgloss.Style
}

type TableOption func(*TableModel)

func WithTableStyle(tableStyle lipgloss.Style) TableOption {
	return func(t *TableModel) {
		t.tableStyle = tableStyle
	}
}

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keyMap struct {
	Esc   key.Binding
	Enter key.Binding
}

func (t TableModel) GetKeyBindings() []key.Binding {
	return []key.Binding{
		t.Keys.Esc, t.Keys.Enter,
	}
}

var defaultBaseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func NewTableModel(t table.Model, opts ...TableOption) *TableModel {
	tableModel := &TableModel{
		table: t,
		Keys: keyMap{
			Esc: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("Escape", "quit table edition"),
			),
			Enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("Enter", "edit cell"),
			),
		},
		tableStyle: defaultBaseStyle,
	}
	for _, opt := range opts {
		opt(tableModel)
	}
	return tableModel
}

func (m TableModel) Init() tea.Cmd {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	m.table.SetStyles(s)
	return nil
}

func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m TableModel) View() string {
	return m.tableStyle.Render(m.table.View()) + "\n"
}
