package top

import (
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
)

func (m *Model) handlePerformanceStart() tea.Cmd {
	if !m.perfMonitorActive {
		m.perfMonitorActive = true
		m.footerModel.SetInfo("Performance monitoring started")
		return tui.PerformanceMonitorTick(performanceMonitorInterval)
	}
	return nil
}

func (m *Model) handlePerformanceStop() tea.Cmd {
	if m.perfMonitorActive {
		m.perfMonitorActive = false
		m.footerModel.SetInfo("Performance monitoring stopped")
		return nil
	}
	return nil
}

// handleSpinnerTick processes spinner tick messages
func (m *Model) handleSpinnerTick(msg spinner.TickMsg) tea.Cmd {
	// If the spinner is not active, return immediately
	var cmd tea.Cmd
	*m.spinner, cmd = m.spinner.Update(msg)
	if m.spinning {
		// Continue spinning spinner.
		return cmd
	}
	return nil
}

// handlePrompt processes prompt messages
func (m *Model) handlePrompt(msg *tui.PromptMsg) tea.Cmd {
	// Enable prompt widget
	m.mode = promptMode
	var blink tea.Cmd
	m.prompt, blink = tui.NewPrompt(msg, m.styles.PromptStyle)
	// Send out message to panes to resize themselves to make room for the prompt above it.
	slog.Debug("handlePrompt", "viewHeight", m.viewHeight())
	return blink
}

func (m *Model) handleError(msg tui.ErrorMsg) tea.Cmd {
	m.footerModel.SetError(error(msg))
	m.messageClearTime = m.messageClearTime.Add(messageDisplayDuration)
	return scheduleClearMessage(messageDisplayDuration)
}

func (m *Model) handleInfo(msg tui.InfoMsg) tea.Cmd {
	m.footerModel.SetInfo(string(msg))
	m.messageClearTime = time.Now().Add(messageDisplayDuration)
	return scheduleClearMessage(messageDisplayDuration)
}

func (m *Model) handleMessageClearTick() tea.Cmd {
	// Only clear the message if the current time is after the scheduled clear time
	// This ensures a new message that resets the clear time won't be cleared prematurely
	if !m.messageClearTime.IsZero() && time.Now().After(m.messageClearTime) {
		m.footerModel.ClearMessages()
		m.messageClearTime = time.Time{}
	}
	return nil
}

// handleWindowSize processes window size messages
func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) tea.Cmd {
	m.width = msg.Width
	m.height = msg.Height
	m.helpModel.SetWidth(m.width)
	m.footerModel.SetWidth(m.width)
	m.headerModel.SetWidth(m.width) // Set the header width
	slog.Debug("handleWindowSize", "viewHeight", m.viewHeight())
	return nil
}

// handleFocusPaneChangedMsg handles the pane focus change message
func (m *Model) handleFocusPaneChangedMsg(_ tea.Msg) tea.Cmd {
	// Update the help bindings when the focused pane changes
	m.updateHelpBindings()
	// Send message to panes to resize themselves to make room for the prompt above it.
	slog.Debug("handleWindowSize", "viewHeight", m.viewHeight())
	return m.PaneManager.Update(tea.WindowSizeMsg{
		Height: m.viewHeight(),
		Width:  m.viewWidth(),
	})
}

// handleBlink processes cursor blink messages
func (m *Model) handleBlink(msg cursor.BlinkMsg) tea.Cmd {
	// Send blink message to prompt if in prompt mode
	if m.mode == promptMode {
		return m.prompt.HandleBlink(msg)
	}
	// otherwise forward it to the active pane to handle.
	return m.FocusedModel().Update(msg)
}
