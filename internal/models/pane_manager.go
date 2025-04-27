package models

import (
	"errors"
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/keys"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/structure"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
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

var (
	ErrAlreadyAtFirstPage  = errors.New("already at first page")
	ErrCannotCloseLastPane = errors.New("cannot close last pane")
	ErrNotFound            = errors.New("resource not found")
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
	commonKeyMap *keys.CommonKeyMap
	globalKeyMap *keys.GlobalKeyMap
	paneKeyMap   *keys.PaneNavigationKeyMap

	// makerFactory for making models for panes
	makerFactory func(kind resource.Kind) Maker
	// panes tracks currently visible panes
	panes map[structure.Position]pane
	// history tracks previously visited models for the top right pane.
	history []pane
	// the position of the currently focused pane
	focused structure.Position
	// total width and height of the terminal space available to panes.
	width, height int
	// leftPaneWidth is the width of the left pane when sharing the terminal
	// with other panes.
	leftPaneWidth int
	// topRightPaneHeight is the height of the top right pane.
	topRightHeight int
}

type pane struct {
	model structure.ChildModel
	page  structure.Page
}

type tablePane interface {
	PreviewCurrentRow() (resource.Kind, resource.ID, bool)
}

// NewPaneManager constructs the pane manager with at least the explorer, which
// occupies the left pane.
func NewPaneManager(
	makerFactory func(kind resource.Kind) Maker,
	myStyles *styles.Styles,
) *PaneManager {
	p := &PaneManager{
		makerFactory:   makerFactory,
		styles:         myStyles,
		cache:          structure.NewCache(),
		panes:          make(map[structure.Position]pane),
		leftPaneWidth:  myStyles.PaneStyle.DefaultLeftPaneWidth,
		topRightHeight: myStyles.PaneStyle.DefaultTopRightPaneHeight,
		commonKeyMap:   keys.GetCommonKeyMap(),
		globalKeyMap:   keys.GetGlobalKeyMap(),
		paneKeyMap:     keys.GetPaneNavigationKeyMap(),
		// The left pane is the default focused pane.
		focused: structure.LeftPane,
		width:   0,
		height:  0,
		history: make([]pane, 0),
	}
	return p
}

func (p *PaneManager) Init() tea.Cmd {
	return p.setPane(structure.NavigationMsg{
		Position:     structure.LeftPane,
		Page:         structure.Page{Kind: structure.CommandListKind, ID: 0},
		DisableFocus: false,
	})
}

//nolint:funlen,cyclop // I don't know how to simplify right now
func (p *PaneManager) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, p.commonKeyMap.Back):
			if p.focused != structure.TopRightPane {
				// History is only maintained for the top right pane.
				break
			}
			if len(p.history) == 1 {
				// At dawn of history; can't go further back.
				return tui.ReportError(ErrAlreadyAtFirstPage)
			}
			// Pop current model from history
			p.history = p.history[:len(p.history)-1]
			// Set pane to last model
			p.panes[structure.TopRightPane] = p.history[len(p.history)-1]
			// A new top right pane replaces any bottom right pane as well.
			delete(p.panes, structure.BottomRightPane)
			p.updateChildSizes()
		case key.Matches(msg, p.globalKeyMap.ShrinkPaneWidth):
			p.updateLeftWidth(-1)
			p.updateChildSizes()
		case key.Matches(msg, p.globalKeyMap.GrowPaneWidth):
			p.updateLeftWidth(1)
			p.updateChildSizes()
		case key.Matches(msg, p.globalKeyMap.ShrinkPaneHeight):
			p.updateTopRightHeight(-1)
			p.updateChildSizes()
		case key.Matches(msg, p.globalKeyMap.GrowPaneHeight):
			p.updateTopRightHeight(1)
			p.updateChildSizes()
		case key.Matches(msg, p.paneKeyMap.SwitchPane):
			p.cycleFocusedPane(false)
		case key.Matches(msg, p.paneKeyMap.SwitchPaneBack):
			p.cycleFocusedPane(true)
		case key.Matches(msg, p.globalKeyMap.ClosePane):
			cmds = append(cmds, p.closeFocusedPane())
		case key.Matches(msg, p.paneKeyMap.LeftPane):
			p.focusPane(structure.LeftPane)
		case key.Matches(msg, p.paneKeyMap.TopRightPane):
			p.focusPane(structure.TopRightPane)
		case key.Matches(msg, p.paneKeyMap.BottomRightPane):
			p.focusPane(structure.BottomRightPane)
		default:
			// Send remaining keys to focused pane
			cmds = append(cmds, p.updateModel(p.focused, msg))
		}
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.updateLeftWidth(0)
		p.updateTopRightHeight(0)
		p.updateChildSizes()
	case structure.NavigationMsg:
		cmds = append(cmds, p.setPane(msg))
	default:
		// Send remaining message types to cached panes.
		cmds = p.cache.UpdateAll(msg)
	}

	// Check that if the top right pane is a table with a current row, then
	// ensure the bottom left pane corresponds to that current row, e.g. if the
	// top right pane is a tasks table, then the bottom right pane shows the
	// output for the current task row.
	if pane, ok := p.panes[structure.TopRightPane]; ok {
		if table, ok := pane.model.(tablePane); ok {
			if kind, id, ok := table.PreviewCurrentRow(); ok {
				cmd := p.setPane(structure.NavigationMsg{
					Page:         structure.Page{Kind: kind, ID: id},
					Position:     structure.BottomRightPane,
					DisableFocus: true,
				})
				cmds = append(cmds, cmd)
			}
		}
	}
	return tea.Batch(cmds...)
}

// FocusedModel retrieves the model of the focused pane.
func (p *PaneManager) FocusedModel() structure.ChildModel {
	return p.panes[p.focused].model
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
	delete(p.panes, p.focused)
	p.updateChildSizes()
	p.cycleFocusedPane(false)
	return nil
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

func (p *PaneManager) updateTopRightHeight(delta int) {
	if _, ok := p.panes[structure.TopRightPane]; !ok {
		// There is no horizontal split to adjust
		return
	} else if _, ok := p.panes[structure.BottomRightPane]; !ok {
		// There is no horizontal split to adjust
		return
	}
	if p.focused == structure.BottomRightPane {
		delta = -delta
	}
	paneStyle := p.styles.PaneStyle
	p.topRightHeight = clamp(
		p.topRightHeight+delta,
		paneStyle.MinPaneHeight,
		p.height-paneStyle.MinPaneWidth,
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
	cmds := []tea.Cmd{}
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
	if msg.Position == structure.TopRightPane {
		// A new top right pane replaces any bottom right pane as well.
		delete(p.panes, structure.BottomRightPane)
		// Track the models for the top right pane, so that the user can go back
		// to previous
		p.history = append(p.history, p.panes[structure.TopRightPane])
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
	case structure.TopRightPane, structure.BottomRightPane:
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
	case structure.TopRightPane:
		if _, ok := p.panes[structure.BottomRightPane]; ok {
			return p.topRightHeight
		}
	case structure.BottomRightPane:
		if _, ok := p.panes[structure.TopRightPane]; ok {
			return p.height - p.topRightHeight
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
					p.renderPane(structure.TopRightPane),
					p.renderPane(structure.BottomRightPane),
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
		case structure.TopRightPane:
			borderTexts[styles.TopRightBorder] = p.paneKeyMap.TopRightPane.Keys()[0]
		case structure.BottomRightPane:
			borderTexts[styles.TopRightBorder] = p.paneKeyMap.BottomRightPane.Keys()[0]
		}
	}
	return styles.Borderize(
		renderedPane, isFocused, borderTexts, p.styles.ColorTheme,
	)
}

func (p *PaneManager) HelpBindings() (bindings []key.Binding) {
	if p.focused == structure.TopRightPane {
		// Only the top right pane has the ability to "go back"
		bindings = append(bindings, p.commonKeyMap.Back)
	}
	if model, ok := p.FocusedModel().(structure.ModelHelpBindings); ok {
		bindings = append(bindings, model.HelpBindings()...)
	}
	return bindings
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
