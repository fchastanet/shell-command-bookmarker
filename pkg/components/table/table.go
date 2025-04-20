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

type Model struct {
	table       *table.Model
	settings    *settings
	tableStyles TableStyles
	Width       int
	Height      int
}

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keyMap struct {
	Esc   key.Binding
	Enter key.Binding
}

type TableStyles interface {
	GetTableStyle() lipgloss.Style
}

func (model *Model) GetKeyBindings() []key.Binding {
	return []key.Binding{
		model.settings.keys.Esc, model.settings.keys.Enter,
	}
}

func NewModel(
	tableModel *table.Model,
	tableStyles TableStyles,
) *Model {
	model := &Model{
		table:       tableModel,
		tableStyles: tableStyles,
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
		Width:  0,
		Height: 0,
	}
	return model
}

func (model *Model) IsFocused() bool {
	return model.table.Focused()
}

func (model *Model) Focus() {
	model.table.Focus()
}

func (model *Model) Blur() {
	model.table.Blur()
}

func (model *Model) SetRows(rows []table.Row) {
	model.table.SetRows(rows)
}

func (model *Model) SetColumns(columns []table.Column) {
	model.table.SetColumns(columns)
}

func (model *Model) SetWidth(width int) {
	model.table.SetWidth(width)
}

func (model *Model) SetHeight(height int) {
	model.table.SetHeight(height)
}

func (model *Model) Init() tea.Cmd {
	return nil
}

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.BlurMsg:
		model.table.Blur()
	case tea.FocusMsg:
		model.table.Focus()
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if model.table.Focused() {
				model.table.Blur()
			} else {
				model.table.Focus()
			}
		case "enter":
			if model.table.Focused() {
				return model, tea.Batch(
					tea.Printf("Let's go to %s!", model.table.SelectedRow()),
				)
			}
		}
	}
	_, cmd = model.table.Update(msg)
	return model, cmd
}

func (model *Model) View() string {
	return model.tableStyles.GetTableStyle().Render(model.table.View())
}
