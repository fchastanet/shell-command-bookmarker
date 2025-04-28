package top

import (
	"strings"

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
	"github.com/fchastanet/shell-command-bookmarker/internal/models/top/help"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	"github.com/fchastanet/shell-command-bookmarker/internal/version"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
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

type Model struct {
	*models.PaneManager
	appService   *services.AppService
	styles       *styles.Styles
	filterKeyMap *keys.FilterKeyMap
	globalKeyMap *keys.GlobalKeyMap
	paneKeyMap   *keys.PaneNavigationKeyMap

	prompt      *models.Prompt
	spinner     *spinner.Model
	footerModel footer.Model

	helpModel help.Model

	width    int
	height   int
	mode     mode
	spinning bool
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

	m := Model{
		PaneManager:  models.NewPaneManager(makers, myStyles),
		filterKeyMap: keys.GetFilterKeyMap(),
		globalKeyMap: keys.GetGlobalKeyMap(),
		paneKeyMap:   keys.GetPaneNavigationKeyMap(),
		spinner:      &spinnerObj,
		appService:   appService,
		styles:       myStyles,
		helpModel:    helpModel,
		footerModel:  footerModel,
		width:        0,
		height:       0,
		mode:         normalMode,
		spinning:     false,
		prompt:       nil,
	}
	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.PaneManager.Init(),
	)
}

//nolint:cyclop // don't see how to simplify right now
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.appService.LoggerService.LogTeaMsg(msg)
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// Keep shared spinner spinning as long as there are tasks running.
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		*m.spinner, cmd = m.spinner.Update(msg)
		if m.spinning {
			// Continue spinning spinner.
			return m, cmd
		}
	case models.PromptMsg:
		// Enable prompt widget
		m.mode = promptMode
		var blink tea.Cmd
		m.prompt, blink = models.NewPrompt(&msg, m.styles.PromptStyle)
		// Send out message to panes to resize themselves to make room for the prompt above it.
		_ = m.PaneManager.Update(tea.WindowSizeMsg{
			Height: m.viewHeight(),
			Width:  m.viewWidth(),
		})
		return m, tea.Batch(cmd, blink)
	case tea.KeyMsg:
		// Pressing any key makes any info/error message in the footer disappear
		m.footerModel.ClearMessages()
		_, teaCmd := m.manageKeyInMode(msg)
		if teaCmd != nil {
			cmds = append(cmds, teaCmd)
			return m, tea.Batch(cmds...)
		}
		return m.manageKey(msg)
	case tui.ErrorMsg:
		m.footerModel.SetError(error(msg))
	case tui.InfoMsg:
		m.footerModel.SetInfo(string(msg))
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.helpModel.SetWidth(m.width)
		m.footerModel.SetWidth(m.width)
		m.PaneManager.Update(tea.WindowSizeMsg{
			Height: m.viewHeight(),
			Width:  m.viewWidth(),
		})
	case cursor.BlinkMsg:
		// Send blink message to prompt if in prompt mode otherwise forward it
		// to the active pane to handle.
		if m.mode == promptMode {
			cmd = m.prompt.HandleBlink(msg)
		} else {
			cmd = m.FocusedModel().Update(msg)
		}
		return m, cmd
	default:
		// Send remaining msg types to pane manager to route accordingly.
		cmds = append(cmds, m.PaneManager.Update(msg))
	}
	return m, tea.Batch(cmds...)
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
	case key.Matches(msg, *m.globalKeyMap.Filter):
		// '/' enables filter mode if the current model indicates it
		// supports it, which it does so by sending back a non-nil command.
		if cmd = m.FocusedModel().Update(tui.FilterFocusReqMsg{}); cmd != nil {
			m.mode = filterMode
		}
		return m, cmd
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
	// Compile list of bindings to render
	bindings := []*key.Binding{}

	switch m.mode {
	case promptMode:
		bindings = append(bindings, m.prompt.HelpBindings()...)
	case filterMode:
		bindings = append(bindings, keys.KeyMapToSlice(*m.filterKeyMap)...)
	case normalMode:
		bindings = append(bindings, m.HelpBindings()...)
		bindings = append(bindings, keys.KeyMapToSlice(*m.globalKeyMap)...)
		bindings = append(bindings, keys.KeyMapToSlice(*m.paneKeyMap)...)
	}

	m.helpModel.SetBindings(bindings)
}

// viewHeight returns the height available to the panes
//
// TODO: rename contentHeight
func (m *Model) viewHeight() int {
	vh := m.height - m.footerModel.Height()
	if m.mode == promptMode {
		vh -= m.styles.PromptStyle.Height
	}
	if m.helpModel.IsVisible() {
		vh -= m.helpModel.Height()
	}
	return max(m.styles.PaneStyle.MinContentHeight, vh)
}

// viewWidth retrieves the width available within the main view
//
// TODO: rename contentWidth
func (m *Model) viewWidth() int {
	return max(m.styles.PaneStyle.MinContentWidth, m.width)
}
