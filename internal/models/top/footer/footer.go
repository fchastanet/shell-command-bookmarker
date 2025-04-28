package footer

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
)

// Model represents the footer component
type Model struct {
	errorMsg      error
	styles        *styles.Styles
	helpWidget    string
	versionWidget string
	infoMsg       string
	width         int
}

// New creates a new footer component
func New(myStyles *styles.Styles, helpWidget, versionWidget string) Model {
	return Model{
		width:         0,
		styles:        myStyles,
		helpWidget:    helpWidget,
		versionWidget: versionWidget,
		errorMsg:      nil,
		infoMsg:       "",
	}
}

// Height returns the height of the footer component when rendered
func (m *Model) Height() int {
	return m.styles.FooterStyle.Height
}

// Width returns the width of the footer component
func (m *Model) Width() int {
	return m.width
}

// SetWidth updates the width of the footer component
func (m *Model) SetWidth(width int) {
	m.width = width
}

// SetError updates the error message displayed in the footer
func (m *Model) SetError(err error) {
	m.errorMsg = err
	m.infoMsg = ""
}

// SetInfo updates the info message displayed in the footer
func (m *Model) SetInfo(info string) {
	m.infoMsg = info
	m.errorMsg = nil
}

// ClearMessages clears both error and info messages
func (m *Model) ClearMessages() {
	m.errorMsg = nil
	m.infoMsg = ""
}

// availableMessageWidth returns the width available for messages
func (m *Model) availableMessageWidth() int {
	// -2 to accommodate padding
	return max(0, m.width-lipgloss.Width(m.helpWidget)-lipgloss.Width(m.versionWidget))
}

// View renders the footer component
func (m *Model) View() string {
	footer := m.helpWidget

	switch {
	case m.errorMsg != nil:
		footer += m.styles.FooterStyle.ErrorStyle.
			Width(m.availableMessageWidth()).
			Render(m.errorMsg.Error())
	case m.infoMsg != "":
		footer += m.styles.FooterStyle.InfoStyle.
			Width(m.availableMessageWidth()).
			Render(m.infoMsg)
	default:
		footer += m.styles.FooterStyle.DefaultStyle.
			Width(m.availableMessageWidth()).
			Render(m.infoMsg)
	}

	footer += m.versionWidget

	return m.styles.FooterStyle.Main.
		MaxWidth(m.width).
		Width(m.width).
		Render(footer)
}
