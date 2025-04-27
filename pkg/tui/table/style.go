package table

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui/colors"
)

const (
	PaddingSmall = 1

	// Height constants
	FilterHeight = 2
	HeaderHeight = 1
)

type Style struct {
	// Border style for the table
	Border *lipgloss.Style
	// Style for the table filters header
	FiltersBlock *lipgloss.Style
	// Style for the table cell
	Cell *lipgloss.Style
	// ScrollbarStyle is the style for the scrollbar.
	ScrollbarStyle *ScrollbarStyle

	Row                   *lipgloss.Style
	CurrentRow            *lipgloss.Style
	SelectedRow           *lipgloss.Style
	CurrentAndSelectedRow *lipgloss.Style

	// Height of the table header
	HeaderHeight int
	// Height of filter widget
	FilterHeight int
}

type ScrollbarStyle struct {
	Thumb string
	Track string
	Width int
}

func GetDefaultStyle() *Style {
	regular := lipgloss.NewStyle()

	CurrentBackground := colors.Grey
	CurrentForeground := colors.White
	SelectedBackground := lipgloss.Color("110")
	SelectedForeground := colors.Black
	CurrentAndSelectedBackground := lipgloss.Color("117")
	CurrentAndSelectedForeground := colors.Black

	tableFilterBlock := regular.Margin(0, 1)
	tableCell := regular.Padding(0, PaddingSmall)
	tableBorderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	// Row styles using the color theme
	row := lipgloss.NewStyle()
	currentRow := tableCell.
		Background(CurrentBackground).
		Foreground(CurrentForeground)
	selectedRow := tableCell.
		Background(SelectedBackground).
		Foreground(SelectedForeground)
	currentAndSelectedRow := tableCell.
		Background(CurrentAndSelectedBackground).
		Foreground(CurrentAndSelectedForeground)

	return &Style{
		Border:                &tableBorderStyle,
		FiltersBlock:          &tableFilterBlock,
		Cell:                  &tableCell,
		Row:                   &row,
		CurrentRow:            &currentRow,
		SelectedRow:           &selectedRow,
		CurrentAndSelectedRow: &currentAndSelectedRow,
		HeaderHeight:          HeaderHeight,
		FilterHeight:          FilterHeight,
		ScrollbarStyle: &ScrollbarStyle{
			Thumb: "█",
			Track: "░",
			Width: 1,
		},
	}
}

func (s *Style) GetTableBorderStyle() *lipgloss.Style {
	return s.Border
}

func (s *Style) GetScrollbarThumb() string {
	return s.ScrollbarStyle.Thumb
}

func (s *Style) GetScrollbarTrack() string {
	return s.ScrollbarStyle.Track
}

func (s *Style) GetScrollbarWidth() int {
	return s.ScrollbarStyle.Width
}
