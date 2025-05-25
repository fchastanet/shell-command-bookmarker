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

//nolint:cyclop // not really complex
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Use enhanced logging for better debugging
	m.appService.LoggerService.EnhancedLogTeaMsg(msg)

	switch typedMsg := msg.(type) {
	case tui.PerformanceMonitorStartMsg:
		return m, m.handlePerformanceStart()
	case tui.PerformanceMonitorStopMsg:
		return m, m.handlePerformanceStop()
	case QuitClearScreenMsg:
		m.quitting = true
		return m, tea.Quit
	case spinner.TickMsg:
		return m, m.handleSpinnerTick(typedMsg)
	case tui.PromptMsg:
		cmd := m.handlePrompt(&typedMsg)
		m.PaneManager.Update(tea.WindowSizeMsg{
			Height: m.viewHeight(),
			Width:  m.viewWidth(),
		})
		return m, cmd
	case tui.ErrorMsg:
		return m, m.handleError(typedMsg)
	case tui.InfoMsg:
		return m, m.handleInfo(typedMsg)
	case MessageClearTickMsg:
		return m, m.handleMessageClearTick()
	case tea.WindowSizeMsg:
		cmd := m.handleWindowSize(typedMsg)
		m.PaneManager.Update(tea.WindowSizeMsg{
			Height: m.viewHeight(),
			Width:  m.viewWidth(),
		})
		return m, cmd
	case tea.KeyMsg:
		return m, m.handleKeyMsg(typedMsg)
	case structure.FocusedPaneChangedMsg:
		return m, m.handleFocusPaneChangedMsg(typedMsg)
	case cursor.BlinkMsg:
		return m, m.handleBlink(typedMsg)
	case tui.MemoryStatsMsg:
		return m, m.handleMemoryStats(typedMsg)
	}

	cmd := m.PaneManager.Update(msg)
	return m, cmd
}

// scheduleClearMessage creates a tick command that triggers after the specified duration
func scheduleClearMessage(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return MessageClearTickMsg{}
	})
}

// handleMemoryStats processes and displays memory statistics
func (m *Model) handleMemoryStats(msg tui.MemoryStatsMsg) tea.Cmd {
	statsInfo := fmt.Sprintf("Memory: %d MB in use | %d MB total | %d MB sys | GC runs: %d",
		msg.Alloc/bytesInMegabyte, msg.TotalAlloc/bytesInMegabyte, msg.Sys/bytesInMegabyte, msg.NumGC)

	m.footerModel.SetInfo(statsInfo)
	return tui.PerformanceMonitorTick(performanceMonitorInterval)
}

//nolint:cyclop // not really complex
func (m *Model) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch m.mode {
	case promptMode:
		closePrompt, cmd := m.prompt.HandleKey(msg)
		var cmd2 tea.Cmd
		if closePrompt {
			// Send message to panes to resize themselves to expand back
			// into space occupied by prompt.
			m.mode = normalMode
			cmd2 = m.PaneManager.Update(tea.WindowSizeMsg{
				Height: m.viewHeight(),
				Width:  m.viewWidth(),
			})
		}
		return tea.Batch(cmd, cmd2)
	case filterMode:
		switch {
		case key.Matches(msg, *m.keyMaps.filter.Blur):
			// Switch back to normal mode, and send message to current model
			// to blur the filter widget
			m.mode = normalMode
			return m.PaneManager.Update(tui.FilterBlurMsg{})
		case key.Matches(msg, *m.keyMaps.filter.Close):
			// Switch back to normal mode, and send message to current model
			// to close the filter widget
			m.mode = normalMode
			closeMsg := tui.FilterCloseMsg{}
			return m.PaneManager.Update(closeMsg)
		default:
			// Wrap key message in a filter key message and send to current
			// model.
			return m.FocusedModel().Update(tui.FilterKeyMsg(msg))
		}
	case normalMode:
		switch {
		case key.Matches(msg, *m.keyMaps.global.Quit):
			// ctrl-c quits the app, but not before prompting the user for
			// confirmation.
			return tui.YesNoPrompt(
				"Quit Shell Command Bookmarker?",
				true,
				QuitWithClearScreen(),
			)
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
			return nil
		case key.Matches(msg, *m.keyMaps.tableAction.Filter):
			// '/' enables filter mode if the current model indicates it
			// supports it, which it does so by sending back a non-nil command.
			cmd := m.FocusedModel().Update(tui.FilterFocusReqMsg{})
			if cmd != nil {
				m.mode = filterMode
			}
			return cmd
		case key.Matches(msg, *m.keyMaps.global.Debug):
			// ctrl+d shows memory stats for debugging performance
			if m.perfMonitorActive {
				return tui.StopPerformanceMonitor()
			}
			return tui.StartPerformanceMonitor(performanceMonitorInterval)
		case key.Matches(msg, *m.keyMaps.global.Search):
			return models.NavigateTo(structure.SearchKind, structure.WithPosition(structure.LeftPane))
		}
	}

	// no key matches, check on children models
	return m.PaneManager.FocusedModel().Update(msg)
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
