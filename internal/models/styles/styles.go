package styles

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	TableStyle  *TableStyle
	PaneStyle   *PaneStyle
	HelpStyle   *HelpStyle
	FooterStyle *FooterStyle
	WindowStyle *WindowStyle
	PromptStyle *PromptStyle
	// ColorTheme is the color theme used in the application.
	ColorTheme *ColorTheme
}

type ScrollbarStyle struct {
	Thumb string
	Track string
	Width int
}

type FooterStyle struct {
	Height       int
	ErrorStyle   *lipgloss.Style
	InfoStyle    *lipgloss.Style
	DefaultStyle *lipgloss.Style
	Main         *lipgloss.Style
	Version      *lipgloss.Style
}

type PaneStyle struct {
	FooterHeight int
	HeaderHeight int
	// defaultLeftPaneWidth is the default width of the left pane.
	DefaultLeftPaneWidth  int
	DefaultRightPaneWidth int
	// defaultTopRightPaneHeight is the default height of the top right pane.
	DefaultTopRightPaneHeight int
	// minimum width of each pane
	MinPaneWidth int
	// minimum height of each pane
	MinPaneHeight int
	// MinContentHeight is the minimum height of content above the footer.
	MinContentHeight int
	// MinContentWidth is the minimum width of the content.
	MinContentWidth int
	BordersWidth    int
	TopBorder       *lipgloss.Style
}

type TableStyle struct {
	// Style for the table content
	Content *table.Styles
	// Border style for the table
	Border *lipgloss.Style
	// Style for the table filters header
	FiltersBlock *lipgloss.Style
	// Style for the table cell
	Cell *lipgloss.Style
	// Height of the table header
	HeaderHeight int
	// Height of filter widget
	FilterHeight int
	// Minimum recommended height for the table widget. Respecting this minimum
	// ensures the header and the borders and the filter widget are visible.
	MinHeight int
	// ScrollbarStyle is the style for the scrollbar.
	ScrollbarStyle *ScrollbarStyle

	Row                   *lipgloss.Style
	CurrentRow            *lipgloss.Style
	SelectedRow           *lipgloss.Style
	CurrentAndSelectedRow *lipgloss.Style
}

type WindowStyle struct {
	// MinHeight is the minimum height of the TUI.
	MinHeight int
	// Height of prompt including borders
	PromptHeight   int
	BorderStyle    *lipgloss.Style
	Background     lipgloss.Color
	Foreground     lipgloss.Color
	DocStyle       *lipgloss.Style
	HighlightColor *lipgloss.AdaptiveColor
}

type HelpStyle struct {
	// Height of help widget, including borders
	Height    int
	Main      *lipgloss.Style
	KeyStyle  *lipgloss.Style
	DescStyle *lipgloss.Style
}

type PromptStyle struct {
	ThickBorder *lipgloss.Style
	Regular     *lipgloss.Style
	PlaceHolder *lipgloss.Style
	Height      int
}

func NewStyles() *Styles {
	s := &Styles{
		TableStyle:  nil,
		PaneStyle:   nil,
		HelpStyle:   nil,
		FooterStyle: nil,
		WindowStyle: nil,
		PromptStyle: nil,
		ColorTheme:  nil,
	}

	// Initialize color theme
	colorTheme := NewDefaultColorTheme()

	// Initialize styles using the color theme
	s.ColorTheme = colorTheme
	s.initBaseStyles(colorTheme)
	s.initComponentStyles(colorTheme)

	return s
}

func (s *Styles) initBaseStyles(colorTheme *ColorTheme) {
	// Create highlight color
	highlightColor := &lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}

	// Setup base styles
	regular := lipgloss.NewStyle()
	padded := regular.Padding(0, PaddingSmall)

	thickBorder := regular.
		Border(lipgloss.ThickBorder()).
		BorderForeground(Violet).
		BorderForeground(Red).
		Padding(0, PaddingSmall)

	footerInline := regular.Inline(true)
	topBorder := regular.Border(lipgloss.NormalBorder())

	docStyle := lipgloss.NewStyle().Padding(
		PaddingSmall, PaddingMedium, PaddingSmall, PaddingMedium,
	)
	windowBorder := lipgloss.NewStyle().
		BorderForeground(highlightColor).
		Padding(0, 0).
		Align(lipgloss.Center).
		Border(lipgloss.NormalBorder()).
		UnsetBorderTop()

	// Initialize window style
	s.WindowStyle = &WindowStyle{
		PromptHeight:   HeightPrompt,
		MinHeight:      HeightMinimum,
		BorderStyle:    &windowBorder,
		Background:     EvenLighterGrey,
		Foreground:     Black,
		DocStyle:       &docStyle,
		HighlightColor: highlightColor,
	}

	// Initialize prompt style
	placeholder := lipgloss.NewStyle().Faint(true)
	s.PromptStyle = &PromptStyle{
		ThickBorder: &thickBorder,
		Regular:     &regular,
		PlaceHolder: &placeholder,
		Height:      HeightPrompt,
	}

	// Initialize pane style
	s.PaneStyle = &PaneStyle{
		DefaultLeftPaneWidth:      WidthLeftPane,
		DefaultRightPaneWidth:     WidthRightPane,
		DefaultTopRightPaneHeight: TopRightPaneHeight,
		MinPaneWidth:              WidthMinPane,
		MinPaneHeight:             HeightMinPane,
		MinContentHeight:          HeightMinimum - HeightFooter,
		MinContentWidth:           WidthMinContent,
		FooterHeight:              HeightFooter,
		HeaderHeight:              HeightFooter,
		TopBorder:                 &topBorder,
		BordersWidth:              BordersWidth,
	}

	// Initialize footer style
	footerDefaultStyle := padded.Foreground(Black).Background(EvenLighterGrey)
	footerErrorStyle := regular.Padding(0, PaddingSmall).
		Background(colorTheme.ErrorLogLevel).
		Foreground(White)
	footerInfoStyle := padded.Foreground(Black).Background(LightGreen)
	versionStyle := padded.Background(DarkGrey).Foreground(White)
	s.FooterStyle = &FooterStyle{
		Height:       HeightFooter,
		DefaultStyle: &footerDefaultStyle,
		ErrorStyle:   &footerErrorStyle,
		InfoStyle:    &footerInfoStyle,
		Main:         &footerInline,
		Version:      &versionStyle,
	}
}

func (s *Styles) initComponentStyles(colorTheme *ColorTheme) {
	// Initialize help style
	padded := lipgloss.NewStyle().Padding(0, PaddingSmall)
	bold := lipgloss.NewStyle().Bold(true)
	regular := lipgloss.NewStyle()

	// Create help style
	helpMainStyle := padded.Background(Grey).Foreground(White)
	helpKeyStyle := bold.Foreground(colorTheme.HelpKey).Margin(0, 1, 0, 0)
	helpDescStyle := regular.Foreground(colorTheme.HelpDesc)
	s.HelpStyle = &HelpStyle{
		Height:    HeightHelp,
		Main:      &helpMainStyle,
		KeyStyle:  &helpKeyStyle,
		DescStyle: &helpDescStyle,
	}

	// Create table styles
	tableFilterBlock := regular.Margin(0, 1)
	tableCell := regular.Padding(0, PaddingSmall)
	tableBorderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	// Row styles using the color theme
	row := lipgloss.NewStyle()
	currentRow := lipgloss.NewStyle().
		Background(colorTheme.CurrentBackground).
		Foreground(colorTheme.CurrentForeground)
	selectedRow := lipgloss.NewStyle().
		Background(colorTheme.SelectedBackground).
		Foreground(colorTheme.SelectedForeground)
	currentAndSelectedRow := lipgloss.NewStyle().
		Background(colorTheme.CurrentAndSelectedBackground).
		Foreground(colorTheme.CurrentAndSelectedForeground)

	// Initialize table style
	s.TableStyle = &TableStyle{
		Content:               tableContentStyles(),
		Border:                &tableBorderStyle,
		FiltersBlock:          &tableFilterBlock,
		Cell:                  &tableCell,
		HeaderHeight:          HeightFooter,
		FilterHeight:          HeightFilter,
		MinHeight:             HeightMinimum,
		Row:                   &row,
		CurrentRow:            &currentRow,
		SelectedRow:           &selectedRow,
		CurrentAndSelectedRow: &currentAndSelectedRow,
		ScrollbarStyle: &ScrollbarStyle{
			Thumb: "█",
			Track: "░",
			Width: 1,
		},
	}
}

func (s *Styles) Init() {
	s.checkDimension()
}

func (s *Styles) checkDimension() {
	slog.Debug("checking dimensions of styles",
		"minContentHeight", s.PaneStyle.MinContentHeight,
		"minPaneHeight", s.PaneStyle.MinPaneHeight,
		"minHelpHeight", MinHelpHeight,
	)
	if s.PaneStyle.MinPaneHeight+MinHelpHeight > s.PaneStyle.MinContentHeight {
		panic("minimum heights of panes, prompt, footer, and help cannot exceed overall minimum height")
	}
	slog.Debug("checking dimensions of styles",
		"minPaneWidth", s.PaneStyle.MinPaneWidth,
		"minContentWidth", s.PaneStyle.MinContentWidth,
	)
	if s.PaneStyle.MinPaneWidth*2 > s.PaneStyle.MinContentWidth {
		panic("minimum width of panes must be no more than half of the minimum content width")
	}
	slog.Debug("checking dimensions of styles",
		"minPaneHeight", s.PaneStyle.MinPaneHeight,
		"minTopRightPaneHeight", s.PaneStyle.DefaultTopRightPaneHeight,
	)
	if s.PaneStyle.MinPaneHeight > s.PaneStyle.DefaultTopRightPaneHeight {
		panic("default top right pane height must not be lower than the overall minimum height")
	}
	slog.Debug("checking dimensions of styles",
		"minPaneWidth", s.PaneStyle.MinPaneWidth,
		"minRightPaneWidth", s.PaneStyle.DefaultLeftPaneWidth,
	)
	if s.PaneStyle.MinPaneWidth > s.PaneStyle.DefaultLeftPaneWidth {
		panic("default left pane width must not be lower than the overall minimum width")
	}
}

func tableContentStyles() *table.Styles {
	s := table.DefaultStyles()
	header := s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Header = &header
	selected := s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	s.Selected = &selected
	return &s
}

func (s *TableStyle) GetTableBorderStyle() *lipgloss.Style {
	return s.Border
}

func (s *TableStyle) GetTableContentStyle() *table.Styles {
	return s.Content
}

func (s *TableStyle) GetScrollbarThumb() string {
	return s.ScrollbarStyle.Thumb
}

func (s *TableStyle) GetScrollbarTrack() string {
	return s.ScrollbarStyle.Track
}

func (s *TableStyle) GetScrollbarWidth() int {
	return s.ScrollbarStyle.Width
}
