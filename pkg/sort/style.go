package sort

import "github.com/charmbracelet/lipgloss"

type Style struct {
	ActiveStyle   *lipgloss.Style
	InactiveStyle *lipgloss.Style
}

type EditorSortStyles interface {
	GetActiveStyle() *lipgloss.Style
	GetInactiveStyle() *lipgloss.Style
}

func GetDefaultEditorSortStyles() EditorSortStyles {
	activeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	return Style{
		ActiveStyle:   &activeStyle,
		InactiveStyle: &inactiveStyle,
	}
}

func (s Style) GetActiveStyle() *lipgloss.Style {
	return s.ActiveStyle
}
func (s Style) GetInactiveStyle() *lipgloss.Style {
	return s.InactiveStyle
}
