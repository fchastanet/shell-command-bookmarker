package models

import (
	"errors"
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/command"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/keys"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/structure"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
	"github.com/fchastanet/shell-command-bookmarker/internal/services/models"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui/table"
	"golang.org/x/exp/maps"
)

type ErrNoMaker struct {
	Kind resource.Kind
}

func (e *ErrNoMaker) Error() string {
	return fmt.Sprintf("no maker for page of kind %s", e.Kind)
}

type ErrMakePage struct {
	Err error
	Msg structure.NavigationMsg
}

func (e *ErrMakePage) Error() string {
	return fmt.Sprintf("making page of kind %s with id %v: %v", e.Msg.Page.Kind, e.Msg.Page.ID, e.Err)
}

type ErrMakePageEmptyModel struct {
	Msg structure.NavigationMsg
}

func (e *ErrMakePageEmptyModel) Error() string {
	return fmt.Sprintf("making page of kind %s with id %v: model is nil", e.Msg.Page.Kind, e.Msg.Page.ID)
}

type FocusedPaneChangedMsg struct {
	From structure.Position
	To   structure.Position
}

func GetFocusedPaneChangedCmd(from, to structure.Position) tea.Cmd {
	return func() tea.Msg {
		return FocusedPaneChangedMsg{From: from, To: to}
	}
}

var (
	ErrAlreadyAtFirstPage  = errors.New("already at first page")
	ErrCannotCloseLastPane = errors.New("cannot close last pane")
)

type CacheInterface interface {
	// Get retrieves a model from the cache.
	Get(page structure.Page) structure.ChildModel
	// Put stores a model in the cache.
	Put(page structure.Page, model structure.ChildModel)
	// UpdateAll updates all models in the cache with the given message.
	UpdateAll(msg tea.Msg) []tea.Cmd
	// Update updates a specific model in the cache with the given message.
	Update(key structure.Page, msg tea.Msg) tea.Cmd
}

// Maker makes new models
type Maker interface {
	Make(id resource.ID, width, height int) (structure.ChildModel, error)
}

// PaneManager manages the layout of the three panes that compose the Pug full screen terminal app.
type PaneManager struct {
	// cache of previously made models
	cache        CacheInterface
	styles       *styles.Styles
	globalKeyMap *keys.GlobalKeyMap
	paneKeyMap   *keys.PaneNavigationKeyMap

	// makerFactory for making models for panes
	makerFactory func(kind resource.Kind) Maker
	// panes tracks currently visible panes
	panes map[structure.Position]pane
	// the position of the currently focused pane
	focused structure.Position
	// total width and height of the terminal space available to panes.
	width, height int
	// leftPaneWidth is the width of the left pane when sharing the terminal
	// with other panes.
	leftPaneWidth int
	// topPaneHeight is the height of the top pane.
	topPaneHeight int
}

type pane struct {
	model structure.ChildModel
	page  structure.Page
}

// NewPaneManager constructs the pane manager with at least the explorer, which
// occupies the left pane.
func NewPaneManager(
	makerFactory func(kind resource.Kind) Maker,
	myStyles *styles.Styles,
	globalKeyMap *keys.GlobalKeyMap,
	paneKeyMap *keys.PaneNavigationKeyMap,
) *PaneManager {
	p := &PaneManager{
		makerFactory:  makerFactory,
		styles:        myStyles,
		cache:         structure.NewCache(),
		panes:         make(map[structure.Position]pane),
		leftPaneWidth: myStyles.PaneStyle.DefaultLeftPaneWidth,
		topPaneHeight: myStyles.PaneStyle.DefaultTopPaneHeight,
		globalKeyMap:  globalKeyMap,
		paneKeyMap:    paneKeyMap,
		// The left pane is the default focused pane.
		focused: structure.TopPane,
		width:   0,
		height:  0,
	}
	return p
}

func (p *PaneManager) Init() tea.Cmd {
	return p.setPane(structure.NavigationMsg{
		Position:     structure.TopPane,
		Page:         structure.Page{Kind: structure.CommandListKind, ID: 0},
		DisableFocus: false,
	})
}

func (p *PaneManager) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	updatePanes := true

	// Handle special messages
	if cmd, handled := p.handleSpecialMessages(msg); handled {
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		updatePanes = false
	}

	if updatePanes {
		// Send keys to focused pane
		cmd := p.updateModel(p.focused, msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
			return tea.Batch(cmds...)
		}
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Handle key events for the pane manager
		cmd := p.handleKeyEvent(keyMsg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	} else if updatePanes {
		// Send remaining message types to cached panes except focused one.
		cmds = append(cmds, p.updateUnfocusedPanes(msg)...)
	}

	return tea.Batch(cmds...)
}

// handleSpecialMessages processes special message types like window size, navigation, etc.
func (p *PaneManager) handleSpecialMessages(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.updateLeftWidth(0)
		p.updateTopHeight(0)
		p.updateChildSizes()
		return nil, true
	case structure.NavigationMsg:
		return p.setPane(msg), true
	case table.RowDefaultActionMsg[*models.Command]:
		return p.setBottomPane(msg.RowID, true), true
	case table.RowSelectedActionMsg[*models.Command]:
		if _, ok := p.panes[structure.BottomPane]; ok {
			return p.setBottomPane(msg.RowID, false), true
		}
	case command.EditorCancelledMsg:
		// The command editor was cancelled, so we need to close the bottom pane
		// and focus the top pane.
		return p.closeFocusedPane(), true
	}

	return nil, false
}

// updateUnfocusedPanes sends messages to all panes except the focused one
func (p *PaneManager) updateUnfocusedPanes(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd

	for position := range p.panes {
		if position == p.focused {
			continue
		}
		if cmd := p.updateModel(position, msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return cmds
}

func (p *PaneManager) handleKeyEvent(keyMsg tea.KeyMsg) tea.Cmd {
	// Handle close pane action separately
	if key.Matches(keyMsg, *p.paneKeyMap.ClosePane) {
		return p.closeFocusedPane()
	}

	// Handle pane resize operations
	if cmd := p.handleResizeKeys(keyMsg); cmd != nil {
		return cmd
	}

	// Handle pane navigation operations
	return p.handleNavigationKeys(keyMsg)
}

// handleNavigationKeys handles key bindings for navigating between panes
func (p *PaneManager) handleNavigationKeys(keyMsg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(keyMsg, *p.paneKeyMap.SwitchPane):
		return p.handleSwitchPane(false)
	case key.Matches(keyMsg, *p.paneKeyMap.SwitchPaneBack):
		return p.handleSwitchPane(true)
	case key.Matches(keyMsg, *p.paneKeyMap.LeftPane):
		return p.handleFocusPane(structure.LeftPane)
	case key.Matches(keyMsg, *p.paneKeyMap.TopPane):
		return p.handleFocusPane(structure.TopPane)
	case key.Matches(keyMsg, *p.paneKeyMap.BottomPane):
		return p.handleFocusPane(structure.BottomPane)
	default:
		return nil
	}
}

// handleResizeKeys handles key bindings for resizing panes
func (p *PaneManager) handleResizeKeys(keyMsg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(keyMsg, *p.paneKeyMap.ShrinkPaneWidth):
		return p.handleShrinkPaneWidth()
	case key.Matches(keyMsg, *p.paneKeyMap.GrowPaneWidth):
		return p.handleGrowPaneWidth()
	case key.Matches(keyMsg, *p.paneKeyMap.ShrinkPaneHeight):
		return p.handleShrinkPaneHeight()
	case key.Matches(keyMsg, *p.paneKeyMap.GrowPaneHeight):
		return p.handleGrowPaneHeight()
	default:
		return nil
	}
}

func (p *PaneManager) handleShrinkPaneWidth() tea.Cmd {
	p.updateLeftWidth(-1)
	p.updateChildSizes()
	return tui.GetDummyCmd()
}

func (p *PaneManager) handleGrowPaneWidth() tea.Cmd {
	p.updateLeftWidth(1)
	p.updateChildSizes()
	return tui.GetDummyCmd()
}

func (p *PaneManager) handleShrinkPaneHeight() tea.Cmd {
	p.updateTopHeight(-1)
	p.updateChildSizes()
	return tui.GetDummyCmd()
}

func (p *PaneManager) handleGrowPaneHeight() tea.Cmd {
	p.updateTopHeight(1)
	p.updateChildSizes()
	return tui.GetDummyCmd()
}

func (p *PaneManager) handleSwitchPane(back bool) tea.Cmd {
	fromPos := p.focused
	p.cycleFocusedPane(back)
	return GetFocusedPaneChangedCmd(fromPos, p.focused)
}

func (p *PaneManager) handleFocusPane(position structure.Position) tea.Cmd {
	fromPos := p.focused
	p.focusPane(position)
	return GetFocusedPaneChangedCmd(fromPos, p.focused)
}

func (p *PaneManager) setBottomPane(rowID resource.ID, focusIfSameRowID bool) tea.Cmd {
	// Handle row default action by opening the editor in the bottom right pane
	bottomPane := p.panes[structure.BottomPane]
	if bottomPane.page.ID == rowID {
		// The bottom right pane is already showing the editor for this command
		// so just bring it into focus.
		if focusIfSameRowID {
			p.focusPane(structure.BottomPane)
		}
		return nil
	}
	return p.setPane(
		structure.NavigationMsg{
			Page: structure.Page{
				Kind: structure.CommandEditorKind,
				ID:   rowID,
			},
			Position:     structure.BottomPane,
			DisableFocus: true,
		},
	)
}

// FocusedModel retrieves the model of the focused pane.
func (p *PaneManager) FocusedModel() structure.ChildModel {
	return p.panes[p.focused].model
}

func (p *PaneManager) FocusedPosition() structure.Position {
	return p.focused
}

// cycleFocusedPane makes the next pane the focused pane. If last is true then the
// previous pane is made the focused pane.
func (p *PaneManager) cycleFocusedPane(last bool) {
	positions := maps.Keys(p.panes)
	slices.Sort(positions)
	var focusedIndex int
	for i, pos := range positions {
		if pos == p.focused {
			focusedIndex = i
		}
	}
	var newFocusedIndex int
	if last {
		newFocusedIndex = focusedIndex - 1
		if newFocusedIndex < 0 {
			newFocusedIndex = len(positions) + newFocusedIndex
		}
	} else {
		newFocusedIndex = (focusedIndex + 1) % len(positions)
	}
	p.focusPane(positions[newFocusedIndex])
}

func (p *PaneManager) closeFocusedPane() tea.Cmd {
	if len(p.panes) == 1 {
		return tui.ReportError(ErrCannotCloseLastPane)
	}
	fromPos := p.focused
	delete(p.panes, p.focused)
	p.updateChildSizes()
	p.cycleFocusedPane(false)
	return GetFocusedPaneChangedCmd(fromPos, p.focused)
}

func (p *PaneManager) updateLeftWidth(delta int) {
	if _, ok := p.panes[structure.LeftPane]; !ok {
		// There is no vertical split to adjust
		return
	}
	paneStyle := p.styles.PaneStyle
	p.leftPaneWidth = clamp(
		p.leftPaneWidth+delta,
		paneStyle.MinPaneWidth,
		p.width-paneStyle.MinPaneWidth,
	)
}

func (p *PaneManager) updateTopHeight(delta int) {
	if _, ok := p.panes[structure.TopPane]; !ok {
		// There is no horizontal split to adjust
		return
	} else if _, ok := p.panes[structure.BottomPane]; !ok {
		// There is no horizontal split to adjust
		return
	}
	if p.focused == structure.BottomPane {
		delta = -delta
	}
	paneStyle := p.styles.PaneStyle
	p.topPaneHeight = clamp(
		p.topPaneHeight+delta,
		paneStyle.MinPaneHeight,
		p.height-paneStyle.MinPaneHeight,
	)
}

func (p *PaneManager) updateChildSizes() {
	for position := range p.panes {
		p.updateModel(position, tea.WindowSizeMsg{
			Width:  p.paneWidth(position) - p.styles.PaneStyle.BordersWidth,
			Height: p.paneHeight(position) - p.styles.PaneStyle.BordersWidth,
		})
	}
}

func (p *PaneManager) updateModel(position structure.Position, msg tea.Msg) tea.Cmd {
	return p.panes[position].model.Update(msg)
}

func (p *PaneManager) setPane(msg structure.NavigationMsg) (cmd tea.Cmd) {
	var cmds []tea.Cmd
	if pane, ok := p.panes[msg.Position]; ok && pane.page == msg.Page {
		// Pane is already showing requested page, so just bring it into focus.
		if !msg.DisableFocus {
			p.focusPane(msg.Position)
		}
		return nil
	}
	model := p.cache.Get(msg.Page)
	if model == nil {
		maker := p.makerFactory(msg.Page.Kind)
		if maker == nil {
			return tui.ReportError(&ErrNoMaker{Kind: msg.Page.Kind})
		}
		var err error
		model, err = maker.Make(msg.Page.ID, 0, 0)
		if err != nil {
			return tui.ReportError(&ErrMakePage{Msg: msg, Err: err})
		}
		if model == nil {
			return tui.ReportError(&ErrMakePageEmptyModel{Msg: msg})
		}
		p.cache.Put(msg.Page, model)
		cmds = append(cmds, model.Init())
	}
	p.panes[msg.Position] = pane{
		model: model,
		page:  msg.Page,
	}
	if msg.Position == structure.TopPane {
		// A new top right pane replaces any bottom right pane as well.
		delete(p.panes, structure.BottomPane)
	}
	p.updateChildSizes()
	if !msg.DisableFocus {
		p.focusPane(msg.Position)
	}
	return tea.Batch(cmds...)
}

func (p *PaneManager) focusPane(position structure.Position) {
	if _, ok := p.panes[position]; !ok {
		// There is no pane to focus at requested position
		return
	}
	p.focused = position
}

func (p *PaneManager) paneWidth(position structure.Position) int {
	switch position {
	case structure.LeftPane:
		if len(p.panes) > 1 {
			return p.leftPaneWidth
		}
	case structure.TopPane, structure.BottomPane:
		paneStyle := p.styles.PaneStyle
		if _, ok := p.panes[structure.LeftPane]; ok {
			return max(
				paneStyle.MinPaneWidth,
				p.width-p.leftPaneWidth,
			)
		}
	}
	return p.width
}

func (p *PaneManager) paneHeight(position structure.Position) int {
	switch position {
	case structure.TopPane:
		if _, ok := p.panes[structure.BottomPane]; ok {
			return p.topPaneHeight
		}
	case structure.BottomPane:
		if _, ok := p.panes[structure.TopPane]; ok {
			return p.height - p.topPaneHeight
		}
	case structure.LeftPane:
		return p.height
	}
	return p.height
}

func (p *PaneManager) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		removeEmptyStrings(
			p.renderPane(structure.LeftPane),
			lipgloss.JoinVertical(lipgloss.Top,
				removeEmptyStrings(
					p.renderPane(structure.TopPane),
					p.renderPane(structure.BottomPane),
				)...,
			),
		)...,
	)
}

func (p *PaneManager) renderPane(position structure.Position) string {
	if _, ok := p.panes[position]; !ok {
		return ""
	}
	model := p.panes[position].model
	isFocused := position == p.focused
	renderedPane := lipgloss.NewStyle().
		Width(p.paneWidth(position) - p.styles.PaneStyle.BordersWidth).
		Height(p.paneHeight(position) - p.styles.PaneStyle.BordersWidth).
		MaxWidth(p.paneWidth(position) - p.styles.PaneStyle.BordersWidth).
		Render(model.View())
	// Optionally, the pane model can embed text in its borders.
	borderTexts := make(map[styles.BorderPosition]string)
	if textInBorder, ok := model.(interface {
		BorderText() map[styles.BorderPosition]string
	}); ok {
		borderTexts = textInBorder.BorderText()
	}
	if !isFocused {
		switch position {
		case structure.LeftPane:
			borderTexts[styles.TopRightBorder] = p.paneKeyMap.LeftPane.Keys()[0]
		case structure.TopPane:
			borderTexts[styles.TopRightBorder] = p.paneKeyMap.TopPane.Keys()[0]
		case structure.BottomPane:
			borderTexts[styles.TopRightBorder] = p.paneKeyMap.BottomPane.Keys()[0]
		}
	}
	return styles.Borderize(
		renderedPane, isFocused, borderTexts, p.styles.ColorTheme,
	)
}

func (p *PaneManager) HelpBindings() (bindings []*key.Binding) {
	paneKeyMap := *p.paneKeyMap
	panesCount := len(p.panes)

	// Get pane presence status
	bottomPanePresent := p.isPanePresent(structure.BottomPane)
	topPanePresent := p.isPanePresent(structure.TopPane)
	leftPanePresent := p.isPanePresent(structure.LeftPane)

	// Add switch pane bindings if we have multiple panes
	if panesCount > 1 {
		bindings = append(
			bindings,
			paneKeyMap.SwitchPane,
			paneKeyMap.SwitchPaneBack,
		)

		// Add pane-specific bindings
		p.addPaneSpecificBindings(&bindings, bottomPanePresent, topPanePresent, leftPanePresent)
	}

	// Add height resize bindings if we have both top and bottom panes
	if bottomPanePresent && topPanePresent {
		bindings = append(bindings, p.paneKeyMap.ShrinkPaneHeight, p.paneKeyMap.GrowPaneHeight)
	}

	// Add width resize bindings if we have left pane
	if leftPanePresent {
		bindings = append(bindings, p.paneKeyMap.ShrinkPaneWidth, p.paneKeyMap.GrowPaneWidth)
	}

	// Add close pane binding if focus is not on top pane and we have other panes
	if p.focused != structure.TopPane && (bottomPanePresent || leftPanePresent) {
		bindings = append(bindings, p.paneKeyMap.ClosePane)
	}

	// Add model-specific bindings if the focused model has them
	if model, ok := p.FocusedModel().(structure.ModelHelpBindings); ok {
		bindings = append(bindings, model.HelpBindings()...)
	}

	return bindings
}

// isPanePresent checks if a pane is present at the given position
func (p *PaneManager) isPanePresent(position structure.Position) bool {
	_, present := p.panes[position]
	return present
}

// addPaneSpecificBindings adds bindings for specific panes if they are present
func (p *PaneManager) addPaneSpecificBindings(bindings *[]*key.Binding, bottomPanePresent, topPanePresent, leftPanePresent bool) {
	if bottomPanePresent {
		*bindings = append(*bindings, p.paneKeyMap.BottomPane)
	}
	if topPanePresent {
		*bindings = append(*bindings, p.paneKeyMap.TopPane)
	}
	if leftPanePresent {
		*bindings = append(*bindings, p.paneKeyMap.LeftPane)
	}
}

func removeEmptyStrings(strings ...string) []string {
	n := 0
	for _, s := range strings {
		if s != "" {
			strings[n] = s
			n++
		}
	}
	return strings[:n]
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
