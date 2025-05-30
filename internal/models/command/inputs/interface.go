package inputs

import tea "github.com/charmbracelet/bubbletea"

type Input interface {
	Update(msg tea.Msg) (Input, tea.Cmd)
	Blur()
	Focus() tea.Cmd
	Value() string
	View() string
	SetWidth(width int)
	SetHeight(height int)
	SetReadOnly(readOnly bool)
	SetValue(value string)
	SetCharLimit(charLimit int)
}
