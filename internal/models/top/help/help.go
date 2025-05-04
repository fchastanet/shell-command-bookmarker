package help

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
)

// BindingSet represents a group of key bindings with a title
type BindingSet struct {
	Title    string
	Bindings []*key.Binding
}

// Model represents the help component
type Model struct {
	helpStyle   *lipgloss.Style
	styles      *styles.Styles
	bindingSets []BindingSet
	width       int
	showHelp    bool
}

// New creates a new help component
func New(myStyles *styles.Styles) Model {
	return Model{
		width:       0,
		styles:      myStyles,
		showHelp:    false,
		bindingSets: []BindingSet{},
		helpStyle:   myStyles.HelpStyle.Main,
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
		return m.maxBindingSetHeight() + m.styles.HelpStyle.BordersWidth
	}
	return 0
}

func (m *Model) maxBindingSetHeight() int {
	maxHeight := 0
	for _, set := range m.bindingSets {
		maxHeight = max(maxHeight, len(set.Bindings)+1) // +1 for the headers
	}
	return maxHeight
}

// Width returns the width of the help component
func (m *Model) Width() int {
	return m.width
}

// SetWidth updates the width of the help component
func (m *Model) SetWidth(width int) {
	m.width = width
}

// AddBindingSet adds a set of bindings with a title
func (m *Model) AddBindingSet(title string, bindings []*key.Binding) {
	m.bindingSets = append(m.bindingSets, BindingSet{
		Title:    title,
		Bindings: bindings,
	})
}

// ClearBindingSets removes all binding sets
func (m *Model) ClearBindingSets() {
	m.bindingSets = []BindingSet{}
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

	return m.viewBindingSets(m.maxBindingSetHeight())
}

// viewBindingSets renders the help component with binding sets
func (m *Model) viewBindingSets(rows int) string {
	columns := make([]string, 0, len(m.bindingSets))
	totalWidth := 0

	// Process each binding set
	for _, set := range m.bindingSets {
		bindings := removeDuplicateBindings(set.Bindings)
		if len(bindings) == 0 {
			continue
		}

		// Render the title with explicit left alignment
		title := m.styles.HelpStyle.TitleStyle.
			Render(set.Title)

		var helpKeys []string
		var descriptions []string

		// Add the title as the first row
		helpKeys = append(helpKeys, title)
		descriptions = append(descriptions, "")

		// Add the bindings
		for j := 0; j < min(rows-1, len(bindings)); j++ {
			helpKeys = append(helpKeys, m.styles.HelpStyle.KeyStyle.Render(bindings[j].Help().Key))
			descriptions = append(descriptions, m.styles.HelpStyle.DescStyle.Render(bindings[j].Help().Desc))
		}

		// Create columns for this set
		cols := []string{
			strings.Join(helpKeys, "\n"),
			strings.Join(descriptions, "\n"),
		}

		// Add spacing between sets
		if len(columns) > 0 {
			columns = append(columns, "   ")
		}

		// Join the columns horizontally
		setContent := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

		// Check if adding this set would exceed available width
		newWidth := totalWidth + lipgloss.Width(setContent) + m.styles.HelpStyle.ColumnMargin
		if len(columns) > 0 && newWidth > m.width-m.styles.HelpStyle.BordersWidth {
			break
		}

		totalWidth = newWidth
		columns = append(columns, setContent)
	}

	// Join all sets horizontally and enclose in a border
	content := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
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
