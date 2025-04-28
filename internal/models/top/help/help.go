package help

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
)

// Model represents the help component
type Model struct {
	helpStyle *lipgloss.Style
	styles    *styles.Styles
	bindings  []*key.Binding
	width     int
	showHelp  bool
}

// New creates a new help component
func New(myStyles *styles.Styles) Model {
	return Model{
		width:     0,
		styles:    myStyles,
		showHelp:  false,
		bindings:  []*key.Binding{},
		helpStyle: myStyles.HelpStyle.Main,
	}
}

// Toggle switches the visibility state of the help component
func (m *Model) Toggle() {
	m.showHelp = !m.showHelp
}

// IsVisible returns whether the help is currently visible
func (m *Model) IsVisible() bool {
	return m.showHelp
}

// Height returns the height of the help component when rendered
func (m *Model) Height() int {
	if m.showHelp {
		return m.styles.HelpStyle.Height
	}
	return 0
}

// Width returns the width of the help component
func (m *Model) Width() int {
	return m.width
}

// SetWidth updates the width of the help component
func (m *Model) SetWidth(width int) {
	m.width = width
}

// SetBindings updates the key bindings displayed in help
func (m *Model) SetBindings(bindings []*key.Binding) {
	m.bindings = bindings
}

// GetHelpWidget returns the help widget text
func (m *Model) GetHelpWidget() string {
	return m.helpStyle.Render("? help")
}

// View renders the help component
func (m *Model) View() string {
	if !m.showHelp {
		return ""
	}

	bindings := removeDuplicateBindings(m.bindings)

	// Enumerate through each group of bindings, populating a series of
	// pairs of columns, one for keys, one for descriptions
	var (
		pairs []string
		width int
		// Subtract 2 to accommodate borders
		rows = m.styles.HelpStyle.Height - 2
	)
	for i := 0; i < len(bindings); i += rows {
		var (
			helpKeys     []string
			descriptions []string
		)
		for j := i; j < min(i+rows, len(bindings)); j++ {
			helpKeys = append(helpKeys, m.styles.HelpStyle.KeyStyle.Render(bindings[j].Help().Key))
			descriptions = append(descriptions, m.styles.HelpStyle.DescStyle.Render(bindings[j].Help().Desc))
		}
		// Render pair of columns; beyond the first pair, render a three space
		// left margin, in order to visually separate the pairs.
		var cols []string
		if len(pairs) > 0 {
			cols = []string{"   "}
		}
		cols = append(cols,
			strings.Join(helpKeys, "\n"),
			strings.Join(descriptions, "\n"),
		)

		pair := lipgloss.JoinHorizontal(lipgloss.Top, cols...)
		// check whether it exceeds the maximum width avail (the width of the
		// terminal, subtracting 2 for the borders).
		width += lipgloss.Width(pair)
		if width > m.width-2 {
			break
		}
		pairs = append(pairs, pair)
	}
	// Join pairs of columns and enclose in a border
	content := lipgloss.JoinHorizontal(lipgloss.Top, pairs...)
	return m.styles.PaneStyle.TopBorder.
		Height(rows).
		Width(m.width - m.styles.PaneStyle.BordersWidth).
		Render(content)
}

// removeDuplicateBindings removes duplicate bindings from a list of bindings. A
// binding is deemed a duplicate if another binding has the same list of keys.
func removeDuplicateBindings(bindings []*key.Binding) []*key.Binding {
	seen := make(map[string]struct{})
	var i int
	for _, b := range bindings {
		bKey := strings.Join(b.Keys(), " ")
		if _, ok := seen[bKey]; ok {
			// duplicate, skip
			continue
		}
		seen[bKey] = struct{}{}
		bindings[i] = b
		i++
	}
	return bindings[:i]
}
