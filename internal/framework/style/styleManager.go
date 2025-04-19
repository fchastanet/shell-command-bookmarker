package style

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type TabStyle struct {
	ActiveTab   lipgloss.Style
	InactiveTab lipgloss.Style
}

type Manager struct {
	DocStyle          lipgloss.Style
	WindowStyle       lipgloss.Style
	HighlightColor    lipgloss.AdaptiveColor
	TableStyle        lipgloss.Style
	TableContentStyle table.Styles
	TabStyle          TabStyle
}

func NewManager() *Manager {
	highlightColor := &lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}

	inactiveTabBorder := tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder := tabBorderWithBottom("┘", " ", "└")
	inactiveTabStyle := lipgloss.NewStyle().
		Border(inactiveTabBorder, true).
		BorderForeground(highlightColor).
		Padding(0, 1)

	return &Manager{
		HighlightColor: *highlightColor,
		DocStyle:       lipgloss.NewStyle().Padding(1, 2, 1, 2),
		WindowStyle: lipgloss.NewStyle().
			BorderForeground(highlightColor).
			Padding(0, 0).
			Align(lipgloss.Center).
			Border(lipgloss.NormalBorder()).
			UnsetBorderTop(),
		TableStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")),
		TableContentStyle: tableContentStyles(),
		TabStyle: TabStyle{
			ActiveTab:   inactiveTabStyle.Border(activeTabBorder, true),
			InactiveTab: inactiveTabStyle,
		},
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

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func (m *Manager) GetTabBorderStyle(
	isFirst bool, isLast bool, isActive bool,
	width int, tabsCount int,
) lipgloss.Style {
	var style lipgloss.Style
	if isActive {
		style = m.TabStyle.ActiveTab
	} else {
		style = m.TabStyle.InactiveTab
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
	borderStyle := style.Border(border)
	borderStyle = borderStyle.Width(
		width/tabsCount -
			(borderStyle.GetBorderLeftSize() + borderStyle.GetBorderRightSize() + width%2),
	)

	return borderStyle
}
