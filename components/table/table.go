package table

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type settings struct {
	keys keyMap
}

type styles struct {
	tableStyle lipgloss.Style
}

type Model struct {
	table    *table.Model
	settings *settings
	styles   *styles
}

type Option func(*Model)

func WithTableStyle(tableStyle *lipgloss.Style) Option {
	return func(model *Model) {
		model.styles.tableStyle = *tableStyle
	}
}

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keyMap struct {
	Esc   key.Binding
	Enter key.Binding
}

func (model Model) GetKeyBindings() []key.Binding {
	return []key.Binding{
		model.settings.keys.Esc, model.settings.keys.Enter,
	}
}

func getDefaultStyles() *styles {
	return &styles{
		tableStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")),
	}
}

func NewModel(tableModel *table.Model, opts ...Option) *Model {
	model := &Model{
		table: tableModel,
		settings: &settings{
			keys: keyMap{
				Esc: key.NewBinding(
					key.WithKeys("esc"),
					key.WithHelp("Escape", "quit table edition"),
				),
				Enter: key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("Enter", "edit cell"),
				),
			},
		},
		styles: getDefaultStyles(),
	}
	for _, opt := range opts {
		opt(model)
	}
	return model
}

func (model Model) Init() tea.Cmd {
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
	model.table.SetStyles(s)
	return nil
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "esc":
			if model.table.Focused() {
				model.table.Blur()
			} else {
				model.table.Focus()
			}
		case "enter":
			return model, tea.Batch(
				tea.Printf("Let's go to %s!", model.table.SelectedRow()[1]),
			)
		}
	}
	tableModel, cmd := model.table.Update(msg)
	model.table = &tableModel
	return model, cmd
}

func (model Model) View() string {
	return model.styles.tableStyle.Render(model.table.View()) + "\n"
}
