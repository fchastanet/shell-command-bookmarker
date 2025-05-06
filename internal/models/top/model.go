package top

import (
	"fmt"
	"log/slog"
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

// QuitClearScreenMsg is a message type for quitting with screen clearing
type QuitClearScreenMsg struct{}

// QuitWithClearScreen returns a command that will quit the application with a clear screen
func QuitWithClearScreen() tea.Cmd {
	return tea.Sequence(
		tea.ExitAltScreen, // Exit alternate screen if it was used
		func() tea.Msg {
			return tea.WindowSizeMsg{Width: 0, Height: 0}
		},
		func() tea.Msg {
			return QuitClearScreenMsg{}
		},
	)
}

// Define constants for magic numbers
const (
	performanceMonitorInterval = 5 * time.Second
	bytesInMegabyte            = 1024 * 1024

	// How long messages remain displayed before auto-clearing
	messageDisplayDuration = 2 * time.Second
)

// MessageClearTickMsg represents a tick to check if messages should be cleared
type MessageClearTickMsg struct{}

type KeyMaps struct {
	filter          *keys.FilterKeyMap
	global          *keys.GlobalKeyMap
	pane            *keys.PaneNavigationKeyMap
	tableNavigation *table.Navigation
	tableAction     *table.Action
	editor          *keys.EditorKeyMap
}

type Model struct {
	// Time when the current message should be cleared
	messageClearTime time.Time

	*models.PaneManager
	appService *services.AppService
	styles     *styles.Styles
	keyMaps    *KeyMaps

	prompt      *tui.Prompt
	spinner     *spinner.Model
	footerModel footer.Model
	headerModel header.Model

	helpModel help.Model

	width  int
	height int
	mode   mode

	// Whether the spinner is currently active
	spinning bool

	// Performance monitoring state
	perfMonitorActive bool

	// Flag to indicate we're quitting and should clear the screen
	quitting bool
}

func NewModel(
	appService *services.AppService,
	myStyles *styles.Styles,
) Model {
	// Work-around for
	// https://github.com/charmbracelet/bubbletea/issues/1036#issuecomment-2158563056
	_ = lipgloss.HasDarkBackground()

	keyMaps := &KeyMaps{
		editor:          keys.GetDefaultEditorKeyMap(),
		filter:          keys.GetFilterKeyMap(),
		global:          keys.GetGlobalKeyMap(),
		pane:            keys.GetPaneNavigationKeyMap(),
		tableNavigation: keys.GetTableNavigationKeyMap(),
		tableAction:     keys.GetTableActionKeyMap(),
	}

	spinnerObj := spinner.New(spinner.WithSpinner(spinner.Line))
	makers := makeMakers(appService, myStyles, &spinnerObj, keyMaps)

	helpWidget := myStyles.HelpStyle.Main.Render("alt+? help")
	versionWidget := myStyles.FooterStyle.Version.Render(version.Get())

	// Create help and footer components
	helpModel := help.New(myStyles)
	footerModel := footer.New(myStyles, helpWidget, versionWidget)

	// Create header component with application name
	headerModel := header.New(myStyles, "Shell Command Bookmarker")

	m := Model{
		PaneManager: models.NewPaneManager(
			makers,
			myStyles,
			keyMaps.global,
			keyMaps.pane,
		),
		spinner:           &spinnerObj,
		appService:        appService,
		styles:            myStyles,
		keyMaps:           keyMaps,
		helpModel:         helpModel,
		footerModel:       footerModel,
		headerModel:       headerModel,
		width:             0,
		height:            0,
		mode:              normalMode,
		spinning:          false,
		prompt:            nil,
		messageClearTime:  time.Time{},
		perfMonitorActive: false,
		quitting:          false,
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
//
//nolint:cyclop // not really complex
func (m *Model) dispatchMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	// First handle performance monitoring messages
	if cmd := m.handlePerformanceMessages(msg); cmd != nil {
		return m, cmd
	}

	// Then handle other specific message types
	switch msg := msg.(type) {
	case QuitClearScreenMsg:
		m.quitting = true
		return m, tea.Quit
	case spinner.TickMsg:
		return m.handleSpinnerTick(msg)
	case tui.PromptMsg:
		return m.handlePrompt(&msg)
	case tui.ErrorMsg, tui.InfoMsg, MessageClearTickMsg:
		return m.handleStatusMsg(msg)
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case structure.FocusedPaneChangedMsg:
		return m.handleFocusPaneChangedMsg(msg)
	case cursor.BlinkMsg:
		return m.handleBlink(msg)
	case tui.MemoryStatsMsg:
		m.handleMemoryStats(msg)
		return m, tui.PerformanceMonitorTick(performanceMonitorInterval)
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		return m.handleKeyMsg(keyMsg)
	}
	cmd := m.PaneManager.Update(msg)
	if cmd != nil {
		return m, cmd
	}
	return m, nil
}

// handleFocusPaneChangedMsg handles the pane focus change message
func (m *Model) handleFocusPaneChangedMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Update the help bindings when the focused pane changes
	m.updateHelpBindings()
	// Send message to panes to resize themselves to make room for the prompt above it.
	slog.Debug("handleWindowSize", "viewHeight", m.viewHeight())
	_ = m.PaneManager.Update(tea.WindowSizeMsg{
		Height: m.viewHeight(),
		Width:  m.viewWidth(),
	})
	m.PaneManager.Update(msg)
	return m, nil
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
func (m *Model) handlePrompt(msg *tui.PromptMsg) (tea.Model, tea.Cmd) {
	// Enable prompt widget
	m.mode = promptMode
	var blink tea.Cmd
	m.prompt, blink = tui.NewPrompt(msg, m.styles.PromptStyle)
	// Send out message to panes to resize themselves to make room for the prompt above it.
	slog.Debug("handlePrompt", "viewHeight", m.viewHeight())
	_ = m.PaneManager.Update(tea.WindowSizeMsg{
		Height: m.viewHeight(),
		Width:  m.viewWidth(),
	})
	return m, blink
}

// handleStatusMsg processes status messages
func (m *Model) handleStatusMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tui.ErrorMsg:
		m.footerModel.SetError(error(msg))
		m.messageClearTime = time.Now().Add(messageDisplayDuration)
		cmds = append(cmds, scheduleClearMessage(messageDisplayDuration))
	case tui.InfoMsg:
		m.footerModel.SetInfo(string(msg))
		m.messageClearTime = time.Now().Add(messageDisplayDuration)
		cmds = append(cmds, scheduleClearMessage(messageDisplayDuration))
	case MessageClearTickMsg:
		// Only clear the message if the current time is after the scheduled clear time
		// This ensures a new message that resets the clear time won't be cleared prematurely
		if !m.messageClearTime.IsZero() && time.Now().After(m.messageClearTime) {
			m.footerModel.ClearMessages()
			m.messageClearTime = time.Time{}
		}
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// scheduleClearMessage creates a tick command that triggers after the specified duration
func scheduleClearMessage(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return MessageClearTickMsg{}
	})
}

// handleWindowSize processes window size messages
func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.helpModel.SetWidth(m.width)
	m.footerModel.SetWidth(m.width)
	m.headerModel.SetWidth(m.width) // Set the header width
	slog.Debug("handleWindowSize", "viewHeight", m.viewHeight())
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

// handleKeyMsg processes key messages
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		case key.Matches(msg, *m.keyMaps.filter.Blur):
			// Switch back to normal mode, and send message to current model
			// to blur the filter widget
			m.mode = normalMode
			cmd = m.FocusedModel().Update(tui.FilterBlurMsg{})
			return m, cmd
		case key.Matches(msg, *m.keyMaps.filter.Close):
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
		cmd := m.PaneManager.Update(msg)
		if cmd != nil {
			return m, cmd
		}
		// In normal mode, we let manageKey handle the key message.
		return m.manageKey(msg)
	}

	return m, nil
}

func (m *Model) manageKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case key.Matches(msg, *m.keyMaps.global.Quit):
		// ctrl-c quits the app, but not before prompting the user for
		// confirmation.
		return m, tui.YesNoPrompt("Quit Shell Command Bookmarker?", true, QuitWithClearScreen())
	case key.Matches(msg, *m.keyMaps.global.Help):
		// '?' toggles help widget
		m.helpModel.Toggle()
		m.updateHelpBindings()
		// Help widget takes up space so update panes' dimensions
		slog.Debug("handleHelpToggle", "viewHeight", m.viewHeight())
		m.PaneManager.Update(tea.WindowSizeMsg{
			Height: m.viewHeight(),
			Width:  m.viewWidth(),
		})
	case key.Matches(msg, *m.keyMaps.tableAction.Filter):
		// '/' enables filter mode if the current model indicates it
		// supports it, which it does so by sending back a non-nil command.
		if cmd = m.FocusedModel().Update(tui.FilterFocusReqMsg{}); cmd != nil {
			m.mode = filterMode
		}
		return m, cmd
	case key.Matches(msg, *m.keyMaps.global.Debug):
		// ctrl+d shows memory stats for debugging performance
		if m.perfMonitorActive {
			return m, tui.StopPerformanceMonitor()
		}
		return m, tui.StartPerformanceMonitor(performanceMonitorInterval)
	case key.Matches(msg, *m.keyMaps.global.Search):
		return m, models.NavigateTo(structure.SearchKind, structure.WithPosition(structure.LeftPane))
	default:
	}
	return m, nil
}

func (m *Model) View() string {
	// When quitting, return an empty string with height 0 to clear the screen
	if m.quitting {
		return lipgloss.NewStyle().Height(0).Render("")
	}

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
	if m.helpModel.IsVisible() {
		switch m.mode {
		case promptMode:
			// For prompt mode, just use a single set
			m.helpModel.AddBindingSet("Prompt Controls", m.prompt.HelpBindings())
		case filterMode:
			// For filter mode, just use a single set
			m.helpModel.AddBindingSet("Filter Controls", keys.KeyMapToSlice(*m.keyMaps.filter))
		case normalMode:
			// For normal mode, organize bindings into logical groups
			m.helpModel.AddBindingSet("Global", keys.KeyMapToSlice(*m.keyMaps.global))
			m.helpModel.AddBindingSet("Pane Navigation", m.HelpBindings())
			if m.FocusedPosition() == structure.TopPane {
				m.helpModel.AddBindingSet("Table Nav", keys.KeyMapToSlice(*m.keyMaps.tableNavigation))
				m.helpModel.AddBindingSet("Table Actions", keys.KeyMapToSlice(*m.keyMaps.tableAction))
			}
		}
	}
}

// contentHeight returns the height available to the panes
func (m *Model) viewHeight() int {
	slog.Debug("viewHeight", "height", m.height)
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
	helpHeight := m.helpModel.Height()
	slog.Debug("viewHeight", "helpHeightIncludingBorders", helpHeight)
	vh -= helpHeight

	slog.Debug("viewHeight", "vh", vh)

	// Ensure we don't go below minimum height
	return max(m.styles.PaneStyle.MinContentHeight, vh)
}

// contentWidth returns the width available within the main view
func (m *Model) viewWidth() int {
	// Use full width but ensure we don't go below minimum width
	return max(m.styles.PaneStyle.MinContentWidth, m.width)
}
