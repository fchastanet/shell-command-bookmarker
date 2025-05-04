package table

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
	"github.com/fchastanet/shell-command-bookmarker/pkg/utils"
	"golang.org/x/exp/maps"
)

const HalfPageMultiplier = 2

// Model defines a state for the table widget.
type Model[V resource.Identifiable] struct {
	borderColor lipgloss.TerminalColor

	previewKind      resource.Kind
	styles           *Style
	navigationKeyMap *Navigation
	actionKeyMap     *Action
	rowRenderer      RowRenderer[V]
	rendered         map[resource.ID]RenderedRow

	// items are the unfiltered set of items available to the table.
	items    map[resource.ID]V
	sortFunc SortFunc[V]

	selected map[resource.ID]V

	border lipgloss.Border
	cols   []Column
	rows   []V

	filter textinput.Model

	currentRowIndex int
	currentRowID    resource.ID

	// index of first visible row
	start int

	// width of table without borders
	width int
	// height of table without borders
	height int

	focused    bool
	selectable bool
}

// Column defines the table structure.
type Column struct {
	TruncationFunc func(s string, w int, tail string) string
	Key            ColumnKey
	// TODO: Default to upper case of key
	Title      string
	Width      int
	FlexFactor int
	// RightAlign aligns content to the right. If false, content is aligned to
	// the left.
	RightAlign bool
}

type ColumnKey string

type RowRenderer[V any] func(V) RenderedRow

// RenderedRow provides the rendered string for each column in a row.
type RenderedRow map[ColumnKey]string

type SortFunc[V any] func(V, V) int

// BulkInsertMsg performs a bulk insertion of entities into a table
type BulkInsertMsg[T any] []T

// New creates a new model for the table widget.
func New[V resource.Identifiable](
	tableStyles *Style,
	cols []Column, fn RowRenderer[V],
	width, height int, opts ...Option[V],
) Model[V] {
	filter := textinput.New()
	filter.Prompt = "Filter: "

	m := Model[V]{
		focused:          false,
		styles:           tableStyles,
		navigationKeyMap: nil,
		actionKeyMap:     nil,
		cols:             make([]Column, len(cols)),
		rows:             []V{},
		rowRenderer:      fn,
		items:            make(map[resource.ID]V),
		rendered:         make(map[resource.ID]RenderedRow),
		selected:         make(map[resource.ID]V),
		selectable:       true,
		filter:           filter,
		border:           lipgloss.NormalBorder(),
		borderColor:      lipgloss.NoColor{},
		currentRowIndex:  -1,
		currentRowID:     resource.ID(0),
		sortFunc:         nil,
		start:            0,
		width:            width,
		height:           height,
		previewKind:      resource.DefaultKind{},
	}
	for _, fn := range opts {
		fn(&m)
	}
	if m.navigationKeyMap == nil {
		m.navigationKeyMap = GetDefaultNavigation()
	}
	if m.actionKeyMap == nil {
		m.actionKeyMap = GetDefaultAction()
	}

	// Copy column structs onto receiver, because the caller may modify columns.
	copy(m.cols, cols)
	// For each column, set default truncation function if unset.
	for i, col := range m.cols {
		if col.TruncationFunc == nil {
			m.cols[i].TruncationFunc = GetDefaultTruncationFunc()
		}
	}

	m.setDimensions(width, height)

	return m
}

type Option[V resource.Identifiable] func(m *Model[V])

// WithSortFunc configures the table to sort rows using the given func.
func WithSortFunc[V resource.Identifiable](sortFunc func(V, V) int) Option[V] {
	return func(m *Model[V]) {
		m.sortFunc = sortFunc
	}
}

// WithNavigation configures the table to use the given navigation keys.
func WithNavigation[V resource.Identifiable](nav *Navigation) Option[V] {
	return func(m *Model[V]) {
		m.navigationKeyMap = nav
	}
}

func WithAction[V resource.Identifiable](action *Action) Option[V] {
	return func(m *Model[V]) {
		m.actionKeyMap = action
	}
}

// WithSelectable sets whether rows are selectable.
func WithSelectable[V resource.Identifiable](s bool) Option[V] {
	return func(m *Model[V]) {
		m.selectable = s
	}
}

// WithPreview configures the table to automatically populate the bottom right
// pane with a model corresponding to the current row.
func WithPreview[V resource.Identifiable](kind resource.Kind) Option[V] {
	return func(m *Model[V]) {
		m.previewKind = kind
	}
}

func (m *Model[V]) IsFocused() bool {
	return m.focused
}

func (m *Model[V]) Focus() {
	m.focused = true
}

func (m *Model[V]) Blur() {
	m.focused = false
	m.filter.Blur()
}

func (m *Model[V]) SetRows(rows []V) {
	m.rows = rows
}

func (m *Model[V]) SetColumns(columns []Column) {
	m.cols = columns
}

func (m *Model[V]) SetWidth(width int) {
	m.width = width
}

func (m *Model[V]) SetHeight(height int) {
	m.height = height
}

func (m *Model[V]) filterVisible() bool {
	// Filter is visible if it's either in focus, or it has a non-empty value.
	return m.filter.Focused() || m.filter.Value() != ""
}

// setDimensions sets the dimensions of the table.
func (m *Model[V]) setDimensions(width, height int) {
	m.height = height
	m.width = width
	m.setColumnWidths()

	m.setStart()
}

// rowAreaHeight returns the height of the terminal allocated to rows.
func (m *Model[V]) rowAreaHeight() int {
	height := max(0, m.height-m.styles.HeaderHeight)

	if m.filterVisible() {
		// Accommodate height of filter widget
		return max(0, height-m.styles.FilterHeight)
	}
	return height
}

// visibleRows returns the number of visible rows that can be
// rendered in the available space.
func (m *Model[V]) visibleRows() int {
	// The number of visible rows cannot exceed the row area height.
	return min(m.rowAreaHeight(), len(m.rows)-m.start)
}

// Update is the Bubble Tea update loop.
func (m *Model[V]) Update(msg tea.Msg) (*Model[V], tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case tea.WindowSizeMsg:
		m.setDimensions(msg.Width, msg.Height)
		return m, nil
	case tui.FilterFocusReqMsg, tui.FilterBlurMsg, tui.FilterCloseMsg, tui.FilterKeyMsg:
		return m.handleFilterMsg(msg)
	case resource.Event[V]:
		return m.handleResourceEvent(msg)
	case BulkInsertMsg[V]:
		m.AddItems(msg...)
		return m, nil
	default:
		if m.filter.Focused() {
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

// handleResourceEvent handles resource events such as create, update, delete
func (m *Model[V]) handleResourceEvent(msg resource.Event[V]) (*Model[V], tea.Cmd) {
	switch msg.Type {
	case resource.CreatedEvent, resource.UpdatedEvent:
		m.AddItems(msg.Payload)
	case resource.DeletedEvent:
		m.removeItem(msg.Payload)
	}
	return m, nil
}

// handleFilterMsg handles filter-related messages
func (m *Model[V]) handleFilterMsg(msg tea.Msg) (*Model[V], tea.Cmd) {
	switch msg := msg.(type) {
	case tui.FilterFocusReqMsg:
		// Focus the filter widget
		blink := m.filter.Focus()
		// Start blinking the cursor
		return m, blink

	case tui.FilterBlurMsg:
		// Blur the filter widget
		m.filter.Blur()
		return m, nil

	case tui.FilterCloseMsg:
		// Close the filter widget
		m.filter.Blur()
		m.filter.SetValue("")
		// Un-filter table items
		m.setRows(maps.Values(m.items)...)
		return m, nil

	case tui.FilterKeyMsg:
		// unwrap key and send to filter widget
		msgKey := tea.KeyMsg(msg)
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msgKey)
		// Filter table items
		m.setRows(maps.Values(m.items)...)
		return m, cmd
	}
	return m, nil
}

// handleKeyMsg processes key press messages
func (m *Model[V]) handleKeyMsg(msg tea.KeyMsg) (*Model[V], tea.Cmd) {
	// Group navigation keys
	if cmd := m.handleNavigationKey(msg); cmd != nil {
		return m, cmd
	}

	// Group selection keys
	if cmd := m.handleSelectionKey(msg); cmd {
		return m, nil
	}

	// Handle action keys
	if cmd := m.handleActionKey(msg); cmd != nil {
		return m, cmd
	}

	return m, nil
}

func (m *Model[V]) handleActionKey(msg tea.KeyMsg) tea.Cmd {
	actions := m.actionKeyMap
	switch {
	case key.Matches(msg, *actions.Enter):
		row, ok := m.CurrentRow()
		if !ok {
			return nil
		}
		return tui.CmdHandler(RowDefaultActionMsg[V]{
			Row:   row,
			RowID: row.GetID(),
			Kind:  m.previewKind,
		})
	case key.Matches(msg, *actions.Reload):
		return tui.CmdHandler(ReloadMsg[V]{
			RowID: -1, InfoMsg: nil,
		})
	case key.Matches(msg, *actions.Delete):
		row, ok := m.CurrentRow()
		if !ok {
			return nil
		}
		return tui.CmdHandler(RowDeleteActionMsg[V]{
			Row:   row,
			RowID: row.GetID(),
		})
	}
	return nil
}

// handleNavigationKey handles all navigation key presses
func (m *Model[V]) handleNavigationKey(msg tea.KeyMsg) tea.Cmd {
	nav := m.navigationKeyMap
	switch {
	case key.Matches(msg, *nav.LineUp):
		m.MoveUp(1)
	case key.Matches(msg, *nav.LineDown):
		m.MoveDown(1)
	case key.Matches(msg, *nav.PageUp):
		m.MoveUp(m.rowAreaHeight())
	case key.Matches(msg, *nav.PageDown):
		m.MoveDown(m.rowAreaHeight())
	case key.Matches(msg, *nav.HalfPageUp):
		m.MoveUp(m.rowAreaHeight() / HalfPageMultiplier)
	case key.Matches(msg, *nav.HalfPageDown):
		m.MoveDown(m.rowAreaHeight() / HalfPageMultiplier)
	case key.Matches(msg, *nav.GotoTop):
		m.GotoTop()
	case key.Matches(msg, *nav.GotoBottom):
		m.GotoBottom()
	default:
		return nil
	}

	return tui.CmdHandler(RowSelectedActionMsg[V]{
		Row:   m.rows[m.currentRowIndex],
		RowID: m.currentRowID,
	})
}

// handleSelectionKey handles all selection key presses
func (m *Model[V]) handleSelectionKey(msg tea.KeyMsg) bool {
	nav := m.actionKeyMap
	switch {
	case key.Matches(msg, *nav.Select):
		m.ToggleSelection()
	case key.Matches(msg, *nav.SelectAll):
		m.SelectAll()
	case key.Matches(msg, *nav.SelectClear):
		m.DeselectAll()
	case key.Matches(msg, *nav.SelectRange):
		m.SelectRange()
	default:
		return false
	}
	return true
}

// PreviewCurrentRow returns information for previewing the current row
func (m *Model[V]) PreviewCurrentRow() (
	resourceKind resource.Kind,
	resourceID resource.ID,
	previewAvailable bool,
) {
	resourceKind = m.previewKind
	previewAvailable = false
	if _, ok := m.CurrentRow(); ok {
		resourceID = m.currentRowID
		previewAvailable = true
	}
	return resourceKind, resourceID, previewAvailable
}

// View renders the table.
func (m *Model[V]) View() string {
	// Table is composed of a vertical stack of components:
	// (a) optional filter widget
	// (b) header
	// (c) rows + scrollbar
	//
	// TODO: this allocation logic is wrong
	components := make([]string, 0, 1+1+m.visibleRows())
	if m.filterVisible() {
		components = append(components,
			m.styles.FiltersBlock.Render(m.filter.View()),
			// Add horizontal rule between filter widget and table
			strings.Repeat("─", m.width))
	}
	components = append(components, m.headersView())
	// Generate scrollbar
	scrollbar := tui.Scrollbar(
		m.styles.ScrollbarStyle,
		m.rowAreaHeight(), len(m.rows), m.visibleRows(), m.start)
	// Get all the visible rows
	rows := make([]string, 0, m.visibleRows())
	for i := range m.visibleRows() {
		rows = append(rows, m.renderRow(m.start+i))
	}
	rowArea := lipgloss.NewStyle().
		Width(m.width - m.styles.ScrollbarStyle.Width).
		Render(strings.Join(rows, "\n"))
	// Put rows alongside the scrollbar to the right.
	components = append(components, lipgloss.JoinHorizontal(lipgloss.Top, rowArea, scrollbar))
	// Render table components, ensuring it is at least a min height
	content := lipgloss.NewStyle().
		Height(m.height).
		MaxHeight(m.height).
		Render(lipgloss.JoinVertical(lipgloss.Top, components...))
	return content
}

// Metadata renders a short string summarizing table row metadata.
func (m *Model[V]) Metadata() string {
	var metadata string
	// Calculate the top and bottom visible row ordinal numbers
	top := m.start + 1
	bottom := m.start + m.visibleRows()
	prefix := fmt.Sprintf("%d-%d of ", top, bottom)
	if m.filterVisible() {
		metadata = prefix + fmt.Sprintf("%d/%d", len(m.rows), len(m.items))
	} else {
		metadata = prefix + strconv.Itoa(len(m.rows))
	}
	return metadata
}

// CurrentRow returns the current row the user has highlighted.  If the table is
// empty then false is returned.
func (m *Model[V]) CurrentRow() (V, bool) {
	if m.currentRowIndex < 0 || m.currentRowIndex >= len(m.rows) {
		return *new(V), false
	}
	return m.rows[m.currentRowIndex], true
}

// SelectedOrCurrent returns either the selected rows, or if there are no
// selections, the current row
func (m *Model[V]) SelectedOrCurrent() []V {
	if len(m.selected) > 0 {
		rows := make([]V, len(m.selected))
		var i int
		for _, v := range m.selected {
			rows[i] = v
			i++
		}
		return rows
	}
	if row, ok := m.CurrentRow(); ok {
		return []V{row}
	}
	return nil
}

// ToggleSelection toggles the selection of the current row.
func (m *Model[V]) ToggleSelection() {
	if !m.selectable {
		return
	}
	current, ok := m.CurrentRow()
	if !ok {
		return
	}
	if _, isSelected := m.selected[current.GetID()]; isSelected {
		delete(m.selected, current.GetID())
	} else {
		m.selected[current.GetID()] = current
	}
}

// ToggleSelectionByID toggles the selection of the row with the given ID. If
// the ID does not exist no action is taken.
func (m *Model[V]) ToggleSelectionByID(id resource.ID) {
	if !m.selectable {
		return
	}
	v, ok := m.items[id]
	if !ok {
		return
	}
	if _, isSelected := m.selected[id]; isSelected {
		delete(m.selected, id)
	} else {
		m.selected[id] = v
	}
}

// SelectAll selects all rows. Any rows not currently selected are selected.
func (m *Model[V]) SelectAll() {
	if !m.selectable {
		return
	}
	for _, row := range m.rows {
		m.selected[row.GetID()] = row
	}
}

// DeselectAll de-selects any rows that are currently selected
func (m *Model[V]) DeselectAll() {
	if !m.selectable {
		return
	}
	m.selected = make(map[resource.ID]V)
}

// SelectRange selects a range of rows. If the current row is *below* a selected
// row then rows between them are selected, including the current row.
// Otherwise, if the current row is *above* a selected row then rows between
// them are selected, including the current row. If there are no selected rows
// then no action is taken.
func (m *Model[V]) SelectRange() {
	if !m.selectable {
		return
	}
	if len(m.selected) == 0 {
		return
	}
	// Determine the first row to select, and the number of rows to select.
	first := -1
	n := 0
	for i, row := range m.rows {
		if i == m.currentRowIndex && first > -1 && first < m.currentRowIndex {
			// Select rows before and including current row
			n = m.currentRowIndex - first + 1
			break
		}
		if _, ok := m.selected[row.GetID()]; !ok {
			// Ignore unselected rows
			continue
		}
		if i > m.currentRowIndex {
			// Select rows including current row and all rows up to but not
			// including next selected row
			first = m.currentRowIndex
			n = i - m.currentRowIndex
			break
		}
		// Start selecting rows after this currently selected row.
		first = i + 1
	}
	for _, row := range m.rows[first : first+n] {
		m.selected[row.GetID()] = row
	}
}

// SetItems overwrites all existing items in the table with items.
func (m *Model[V]) SetItems(items ...V) {
	m.items = make(map[resource.ID]V)
	m.rendered = make(map[resource.ID]RenderedRow)
	m.AddItems(items...)
}

// AddItems idem potently adds items to the table,
// updating any items that exist on the table already.
func (m *Model[V]) AddItems(items ...V) {
	for _, item := range items {
		// Add/update item
		m.items[item.GetID()] = item
		// (Re-)render item's row.
		m.rendered[item.GetID()] = m.rowRenderer(item)
	}
	m.setRows(maps.Values(m.items)...)
}

func (m *Model[V]) removeItem(item V) {
	delete(m.rendered, item.GetID())
	delete(m.items, item.GetID())
	delete(m.selected, item.GetID())
	for i, row := range m.rows {
		if row.GetID() == item.GetID() {
			// TODO: this might well produce a memory leak. See note:
			// https://go.dev/wiki/SliceTricks#delete-without-preserving-order
			m.rows = append(m.rows[:i], m.rows[i+1:]...)
			break
		}
	}
	if item.GetID() == m.currentRowID {
		// If item being removed is the current row the make the row above it
		// the new current row. (MoveUp also calls setStart, see below).
		m.MoveUp(1)
	} else {
		// Removing item may well affect index of first visible row, so
		// re-calculate just in case.
		m.setStart()
	}
}

// setRows processes and filters items for display
func (m *Model[V]) setRows(items ...V) {
	// Process items with filtering
	m.processFilteredItems(items)

	// Sort rows and locate current row
	m.sortAndLocateCurrentRow()

	// Set start index
	m.setStart()
}

// processFilteredItems handles filtering of items for display
func (m *Model[V]) processFilteredItems(items []V) {
	selected := make(map[resource.ID]V)
	m.rows = make([]V, 0, len(items))

	for _, item := range items {
		if m.filterVisible() && !m.matchFilter(item) {
			// Skip item that doesn't match filter
			continue
		}
		m.rows = append(m.rows, item)
		if m.selectable {
			if _, ok := m.selected[item.GetID()]; ok {
				selected[item.GetID()] = item
			}
		}
	}
	m.selected = selected
}

// sortAndLocateCurrentRow sorts the rows and tracks the current row
func (m *Model[V]) sortAndLocateCurrentRow() {
	// Sort rows in-place
	if m.sortFunc != nil {
		slices.SortFunc(m.rows, func(i, j V) int {
			return m.sortFunc(i, j)
		})
	}

	// Track current row index
	m.currentRowIndex = -1
	for i, row := range m.rows {
		if row.GetID() == m.currentRowID {
			m.currentRowIndex = i
			break
		}
	}

	// Set default current row if needed
	if len(m.rows) > 0 && m.currentRowIndex == -1 {
		m.currentRowIndex = 0
		m.currentRowID = m.rows[m.currentRowIndex].GetID()
	}
}

// matchFilter returns true if the item with the given ID matches the filter
// value.
func (m *Model[V]) matchFilter(item V) bool {
	for _, col := range m.rendered[item.GetID()] {
		if strings.Contains(col, m.filter.Value()) {
			return true
		}
	}
	return false
}

// MoveUp moves the current row up by any number of rows.
// It can not go above the first row.
func (m *Model[V]) MoveUp(n int) {
	m.moveCurrentRow(-n)
}

// MoveDown moves the current row down by any number of rows.
// It can not go below the last row.
func (m *Model[V]) MoveDown(n int) {
	m.moveCurrentRow(n)
}

func (m *Model[V]) moveCurrentRow(n int) {
	if len(m.rows) > 0 {
		m.currentRowIndex = clamp(m.currentRowIndex+n, 0, len(m.rows)-1)
		m.currentRowID = m.rows[m.currentRowIndex].GetID()
		m.setStart()
	}
}

func (m *Model[V]) setStart() {
	// Start index must be at least the current row index minus the max number
	// of visible rows.
	minimum := max(0, m.currentRowIndex-m.rowAreaHeight()+1)
	// Start index must be at most the lesser of:
	// (a) the current row index, or
	// (b) the number of rows minus the maximum number of visible rows (as many
	// rows as possible are rendered)
	maximum := max(0, min(m.currentRowIndex, len(m.rows)-m.rowAreaHeight()))
	m.start = clamp(m.start, minimum, maximum)
}

// GotoTop makes the top row the current row.
func (m *Model[V]) GotoID(id resource.ID) {
	if id == m.currentRowID {
		return
	}
	if _, ok := m.items[id]; ok {
		m.currentRowID = id
		m.currentRowIndex = 0
		for i, r := range m.rows {
			if r.GetID() == id {
				m.currentRowIndex = i
				break
			}
		}
	}
	m.setStart()
}

// GotoTop makes the top row the current row.
func (m *Model[V]) GotoTop() {
	m.MoveUp(m.currentRowIndex)
}

// GotoBottom makes the bottom row the current row.
func (m *Model[V]) GotoBottom() {
	m.MoveDown(len(m.rows))
}

// GetNextRowIDRelativeToCurrentRow returns the ID of the row after the current row
// If the current row is the last row, it returns previous row ID
// If the current row is the last row, it returns 0
func (m *Model[V]) GetNextRowIDRelativeToCurrentRow() resource.ID {
	if m.currentRowIndex+1 >= len(m.rows) {
		// If the current row is the last row, return previous row ID
		if m.currentRowIndex-1 >= 0 {
			return m.rows[m.currentRowIndex-1].GetID()
		}
		// If the current row is the first row, return 0
		return resource.ID(0)
	}
	return m.rows[m.currentRowIndex+1].GetID()
}

func (m *Model[V]) headersView() string {
	s := make([]string, 0, len(m.cols))
	// TODO: use index and don't use append below
	for _, col := range m.cols {
		style := lipgloss.NewStyle().Width(col.Width).MaxWidth(col.Width).Inline(true)
		if col.RightAlign {
			style = style.AlignHorizontal(lipgloss.Right)
		}
		renderedCell := style.Render(TruncateRight(col.Title, col.Width, "…"))
		s = append(s, m.styles.Cell.Render(renderedCell))
	}
	return lipgloss.NewStyle().
		MaxWidth(m.width).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, s...))
}

func (m *Model[V]) renderRow(rowIdx int) string {
	row := m.rows[rowIdx]

	var (
		current  bool
		selected bool
	)
	if _, ok := m.selected[row.GetID()]; ok {
		selected = true
	}
	current = rowIdx == m.currentRowIndex
	rowStyle := m.styles.Cell

	switch {
	case current && selected:
		rowStyle = m.styles.CurrentAndSelectedRow
	case current:
		rowStyle = m.styles.CurrentRow
	case selected:
		rowStyle = m.styles.SelectedRow
	}

	cells := m.rendered[row.GetID()]
	styledCells := make([]string, len(m.cols))
	for i, col := range m.cols {
		content := cells[col.Key]
		// Truncate content if it is wider than column
		truncated := col.TruncationFunc(content, col.Width, "…")
		// Ensure content is all on one line.
		style := lipgloss.NewStyle().
			Width(col.Width).
			MaxWidth(col.Width).
			Inline(true)
		if col.RightAlign {
			style = style.AlignHorizontal(lipgloss.Right)
		}

		if current || selected {
			// For highlighted rows, ignore any styling in the cell content
			// and just apply the row's background/foreground colors
			style.
				Background(rowStyle.GetBackground()).
				Foreground(rowStyle.GetForeground())

			truncated = utils.RemoveAnsiCodes(truncated)
		}
		// For normal rows, just apply the regular styling
		inlined := style.Render(truncated)
		// Apply block-styling to content
		boxed := lipgloss.NewStyle().
			PaddingRight(1 + m.styles.Cell.GetPaddingLeft()).
			Render(inlined)
		styledCells[i] = boxed
	}

	// Join cells together to form a row, ensuring it doesn't exceed maximum
	// table width
	renderedRow := lipgloss.JoinHorizontal(lipgloss.Left, styledCells...)
	// Apply row style
	renderedRow = rowStyle.
		MaxWidth(m.width).
		Render(renderedRow)

	return renderedRow
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
