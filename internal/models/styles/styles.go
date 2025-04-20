package styles

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	docStyle          lipgloss.Style
	windowStyle       lipgloss.Style
	highlightColor    lipgloss.AdaptiveColor
	tableStyle        lipgloss.Style
	tableContentStyle table.Styles
}

func NewStyles() *Styles {
	checkDimension()
	highlightColor := &lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}

	return &Styles{
		highlightColor: *highlightColor,
		docStyle:       lipgloss.NewStyle().Padding(1, 2, 1, 2),
		windowStyle: lipgloss.NewStyle().
			BorderForeground(highlightColor).
			Padding(0, 0).
			Align(lipgloss.Center).
			Border(lipgloss.NormalBorder()).
			UnsetBorderTop(),
		tableStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")),
		tableContentStyle: tableContentStyles(),
	}
}

func tableContentStyles() table.Styles {
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
	return s
}

func (s *Styles) GetTableStyle() lipgloss.Style {
	return s.tableStyle
}
func (s *Styles) GetDocStyle() lipgloss.Style {
	return s.docStyle
}
func (s *Styles) GetWindowStyle() lipgloss.Style {
	return s.windowStyle
}
func (s *Styles) GetHighlightColor() lipgloss.AdaptiveColor {
	return s.highlightColor
}
func (s *Styles) GetTableContentStyle() table.Styles {
	return s.tableContentStyle
}
