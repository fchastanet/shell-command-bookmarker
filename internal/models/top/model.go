package top

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fchastanet/shell-command-bookmarker/internal/models"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/keys"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/structure"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/top/footer"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/top/header"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/top/help"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	"github.com/fchastanet/shell-command-bookmarker/internal/version"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui/table"
)

// alter how all messages are handled.
type mode int

const (
	normalMode mode = iota // default
	promptMode             // confirm prompt is visible and taking input
	filterMode             // filter is visible and taking input
)

// indicate parent components that filter has been closed.
type FilterClosedMsg struct{}

func FilterClosed() tea.Msg {
	return FilterClosedMsg{}
}

// Define constants for magic numbers
const (
	performanceMonitorInterval = 5 * time.Second
	bytesInMegabyte            = 1024 * 1024
)

type Model struct {
	*models.PaneManager
	appService            *services.AppService
	styles                *styles.Styles
	filterKeyMap          *keys.FilterKeyMap
	globalKeyMap          *keys.GlobalKeyMap
	paneKeyMap            *keys.PaneNavigationKeyMap
	tableNavigationKeyMap *table.Navigation
	tableActionKeyMap     *table.Action

	prompt      *models.Prompt
	spinner     *spinner.Model
	footerModel footer.Model
	headerModel header.Model

	helpModel help.Model

	width    int
	height   int
	mode     mode
	spinning bool

	// Performance monitoring state
	perfMonitorActive bool
}

func NewModel(
	appService *services.AppService,
	myStyles *styles.Styles,
) Model {
	// Work-around for
	// https://github.com/charmbracelet/bubbletea/issues/1036#issuecomment-2158563056
	_ = lipgloss.HasDarkBackground()

	spinnerObj := spinner.New(spinner.WithSpinner(spinner.Line))
	makers := makeMakers(appService, myStyles, &spinnerObj)

	helpWidget := myStyles.HelpStyle.Main.Render("? help")
	versionWidget := myStyles.FooterStyle.Version.Render(version.Get())

	// Create help and footer components
	helpModel := help.New(myStyles)
	footerModel := footer.New(myStyles, helpWidget, versionWidget)

	// Create header component with application name
	headerModel := header.New(myStyles, "Shell Command Bookmarker")

	m := Model{
		PaneManager:           models.NewPaneManager(makers, myStyles),
		filterKeyMap:          keys.GetFilterKeyMap(),
		globalKeyMap:          keys.GetGlobalKeyMap(),
		paneKeyMap:            keys.GetPaneNavigationKeyMap(),
		tableNavigationKeyMap: keys.GetTableNavigationKeyMap(),
		tableActionKeyMap:     keys.GetTableActionKeyMap(),
		spinner:               &spinnerObj,
		appService:            appService,
		styles:                myStyles,
		helpModel:             helpModel,
		footerModel:           footerModel,
		headerModel:           headerModel,
		width:                 0,
		height:                0,
		mode:                  normalMode,
		spinning:              false,
		prompt:                nil,
		perfMonitorActive:     false,
	}
	return m
}

func (m *Model) Init() tea.Cmd {
	return models.SafeCmd(tea.Batch(
		m.PaneManager.Init(),
	))
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Use enhanced logging for better debugging
	m.appService.LoggerService.EnhancedLogTeaMsg(msg)

	return m.dispatchMessage(msg)
}

// dispatchMessage routes messages to their appropriate handlers
func (m *Model) dispatchMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	// First handle performance monitoring messages
	if cmd := m.handlePerformanceMessages(msg); cmd != nil {
		return m, cmd
	}

	// Then handle other specific message types
	switch msg := msg.(type) {
	case spinner.TickMsg:
		return m.handleSpinnerTick(msg)
	case models.PromptMsg:
		return m.handlePrompt(&msg)
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case tui.ErrorMsg, tui.InfoMsg:
		return m.handleStatusMsg(msg)
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case cursor.BlinkMsg:
		return m.handleBlink(msg)
	case tui.MemoryStatsMsg:
		m.handleMemoryStats(msg)
		return m, tui.PerformanceMonitorTick(performanceMonitorInterval)
	default:
		return m.handleGenericMessage(msg)
	}
}

// handlePerformanceMessages handles performance monitoring related messages
func (m *Model) handlePerformanceMessages(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case tui.PerformanceMonitorStartMsg:
		if !m.perfMonitorActive {
			m.perfMonitorActive = true
			m.footerModel.SetInfo("Performance monitoring started")
			return tui.PerformanceMonitorTick(performanceMonitorInterval)
		}
	case tui.PerformanceMonitorStopMsg:
		if m.perfMonitorActive {
			m.perfMonitorActive = false
			m.footerModel.SetInfo("Performance monitoring stopped")
			return nil
		}
	}
	return nil
}

// handleGenericMessage handles any message types not explicitly handled in dispatchMessage
func (m *Model) handleGenericMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Send remaining msg types to pane manager to route accordingly.
	return m, m.PaneManager.Update(msg)
}

// handleSpinnerTick processes spinner tick messages
func (m *Model) handleSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	*m.spinner, cmd = m.spinner.Update(msg)
	if m.spinning {
		// Continue spinning spinner.
		return m, cmd
	}
	return m, nil
}

// handlePrompt processes prompt messages
func (m *Model) handlePrompt(msg *models.PromptMsg) (tea.Model, tea.Cmd) {
	// Enable prompt widget
	m.mode = promptMode
	var blink tea.Cmd
	m.prompt, blink = models.NewPrompt(msg, m.styles.PromptStyle)
	// Send out message to panes to resize themselves to make room for the prompt above it.
	_ = m.PaneManager.Update(tea.WindowSizeMsg{
		Height: m.viewHeight(),
		Width:  m.viewWidth(),
	})
	return m, blink
}

// handleKeyMsg processes key messages
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Pressing any key makes any info/error message in the footer disappear
	m.footerModel.ClearMessages()
	_, teaCmd := m.manageKeyInMode(msg)
	if teaCmd != nil {
		return m, teaCmd
	}
	return m.manageKey(msg)
}

// handleStatusMsg processes status messages
func (m *Model) handleStatusMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tui.ErrorMsg:
		m.footerModel.SetError(error(msg))
	case tui.InfoMsg:
		m.footerModel.SetInfo(string(msg))
	}
	return m, nil
}

// handleWindowSize processes window size messages
func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.helpModel.SetWidth(m.width)
	m.footerModel.SetWidth(m.width)
	m.headerModel.SetWidth(m.width) // Set the header width
	return m, m.PaneManager.Update(tea.WindowSizeMsg{
		Height: m.viewHeight(),
		Width:  m.viewWidth(),
	})
}

// handleBlink processes cursor blink messages
func (m *Model) handleBlink(msg cursor.BlinkMsg) (tea.Model, tea.Cmd) {
	// Send blink message to prompt if in prompt mode otherwise forward it
	// to the active pane to handle.
	if m.mode == promptMode {
		return m, m.prompt.HandleBlink(msg)
	}
	return m, m.FocusedModel().Update(msg)
}

// handleMemoryStats processes and displays memory statistics
func (m *Model) handleMemoryStats(msg tui.MemoryStatsMsg) {
	statsInfo := fmt.Sprintf("Memory: %d MB in use | %d MB total | %d MB sys | GC runs: %d",
		msg.Alloc/bytesInMegabyte, msg.TotalAlloc/bytesInMegabyte, msg.Sys/bytesInMegabyte, msg.NumGC)

	m.footerModel.SetInfo(statsInfo)
}

func (m *Model) manageKeyInMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.mode {
	case promptMode:
		closePrompt, cmd := m.prompt.HandleKey(msg)
		if closePrompt {
			// Send message to panes to resize themselves to expand back
			// into space occupied by prompt.
			m.mode = normalMode
			_ = m.PaneManager.Update(tea.WindowSizeMsg{
				Height: m.viewHeight(),
				Width:  m.viewWidth(),
			})
		}
		return m, cmd
	case filterMode:
		switch {
		case key.Matches(msg, *m.filterKeyMap.Blur):
			// Switch back to normal mode, and send message to current model
			// to blur the filter widget
			m.mode = normalMode
			cmd = m.FocusedModel().Update(tui.FilterBlurMsg{})
			return m, cmd
		case key.Matches(msg, *m.filterKeyMap.Close):
			// Switch back to normal mode, and send message to current model
			// to close the filter widget
			m.mode = normalMode
			closeMsg := tui.FilterCloseMsg{}
			cmd = m.FocusedModel().Update(closeMsg)
			if cmd == nil {
				return m, FilterClosed
			}
			return m, cmd
		default:
			// Wrap key message in a filter key message and send to current
			// model.
			cmd = m.FocusedModel().Update(tui.FilterKeyMsg(msg))
			return m, cmd
		}
	case normalMode:
		// In normal mode, we let manageKey handle the key message.
	}

	return m, nil
}

func (m *Model) manageKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case key.Matches(msg, *m.globalKeyMap.Quit):
		// ctrl-c quits the app, but not before prompting the user for
		// confirmation.
		return m, models.YesNoPrompt("Quit Shell Command Bookmarker?", tea.Quit)
	case key.Matches(msg, *m.globalKeyMap.Help):
		// '?' toggles help widget
		m.helpModel.Toggle()
		// Help widget takes up space so update panes' dimensions
		m.PaneManager.Update(tea.WindowSizeMsg{
			Height: m.viewHeight(),
			Width:  m.viewWidth(),
		})
	case key.Matches(msg, *m.tableActionKeyMap.Filter):
		// '/' enables filter mode if the current model indicates it
		// supports it, which it does so by sending back a non-nil command.
		if cmd = m.FocusedModel().Update(tui.FilterFocusReqMsg{}); cmd != nil {
			m.mode = filterMode
		}
		return m, cmd
	case key.Matches(msg, *m.globalKeyMap.Debug):
		// ctrl+d shows memory stats for debugging performance
		if m.perfMonitorActive {
			return m, tui.StopPerformanceMonitor()
		}
		return m, tui.StartPerformanceMonitor(performanceMonitorInterval)
	case key.Matches(msg, *m.globalKeyMap.Search):
		return m, models.NavigateTo(structure.SearchKind, structure.WithPosition(structure.LeftPane))
	default:
	}
	// Send all other keys to panes.
	if cmd := m.PaneManager.Update(msg); cmd != nil {
		return m, cmd
	}
	return m, nil
}

func (m *Model) View() string {
	// Start composing vertical stack of components that fill entire terminal.
	var components []string

	// Add header with application name
	components = append(components, m.headerModel.View())

	// Add prompt if in prompt mode.
	if m.mode == promptMode {
		components = append(components, m.prompt.View(m.width))
	}
	// Add panes
	components = append(components, lipgloss.NewStyle().
		Height(m.viewHeight()).
		Width(m.viewWidth()).
		Render(m.PaneManager.View()),
	)

	// Add help if enabled
	if m.helpModel.IsVisible() {
		// Update help bindings before rendering
		m.updateHelpBindings()
		components = append(components, m.helpModel.View())
	}

	// Add footer
	components = append(components, m.footerModel.View())

	return strings.Join(components, "\n")
}

// updateHelpBindings updates the key bindings displayed in help based on current mode
func (m *Model) updateHelpBindings() {
	// Clear previous binding sets
	m.helpModel.ClearBindingSets()

	switch m.mode {
	case promptMode:
		// For prompt mode, just use a single set
		m.helpModel.AddBindingSet("Prompt Controls", m.prompt.HelpBindings())
	case filterMode:
		// For filter mode, just use a single set
		m.helpModel.AddBindingSet("Filter Controls", keys.KeyMapToSlice(*m.filterKeyMap))
	case normalMode:
		// For normal mode, organize bindings into logical groups
		m.helpModel.AddBindingSet("Global", keys.KeyMapToSlice(*m.globalKeyMap))
		if len(m.HelpBindings()) > 0 {
			m.helpModel.AddBindingSet("Pane Actions", m.HelpBindings())
		}
		m.helpModel.AddBindingSet("Table Nav", keys.KeyMapToSlice(*m.tableNavigationKeyMap))
		m.helpModel.AddBindingSet("Table Actions", keys.KeyMapToSlice(*m.tableActionKeyMap))
	}
}

// contentHeight returns the height available to the panes
func (m *Model) viewHeight() int {
	// Start with full height
	vh := m.height

	// Subtract header height
	vh -= m.headerModel.Height()

	// Subtract footer height
	vh -= m.footerModel.Height()

	// Subtract prompt height if in prompt mode
	if m.mode == promptMode {
		vh -= lipgloss.Height(m.prompt.View(m.width))
	}

	// Subtract help height if visible
	if m.helpModel.IsVisible() {
		vh -= m.helpModel.Height()
	}

	// Ensure we don't go below minimum height
	return max(m.styles.PaneStyle.MinContentHeight, vh)
}

// contentWidth returns the width available within the main view
func (m *Model) viewWidth() int {
	// Use full width but ensure we don't go below minimum width
	return max(m.styles.PaneStyle.MinContentWidth, m.width)
}
