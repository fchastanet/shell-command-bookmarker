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

type Model struct {
	*models.PaneManager
	appService   *services.AppService
	styles       *styles.Styles
	filterKeyMap *keys.FilterKeyMap
	globalKeyMap *keys.GlobalKeyMap
	paneKeyMap   *keys.PaneNavigationKeyMap

	width         int
	height        int
	mode          mode
	showHelp      bool
	prompt        *models.Prompt
	spinner       *spinner.Model
	spinning      bool
	err           error
	info          string
	versionWidget string
	helpWidget    string
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

	m := Model{
		PaneManager:   models.NewPaneManager(makers, myStyles),
		filterKeyMap:  keys.GetFilterKeyMap(),
		globalKeyMap:  keys.GetGlobalKeyMap(),
		paneKeyMap:    keys.GetPaneNavigationKeyMap(),
		spinner:       &spinnerObj,
		appService:    appService,
		styles:        myStyles,
		helpWidget:    helpWidget,
		versionWidget: versionWidget,
		width:         0,
		height:        0,
		mode:          normalMode,
		showHelp:      false,
		spinning:      false,
		err:           nil,
		info:          "",
		prompt:        nil,
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
		m.info = ""
		m.err = nil
		_, teaCmd := m.manageKeyInMode(msg)
		if teaCmd != nil {
			cmds = append(cmds, teaCmd)
			return m, tea.Batch(cmds...)
		}
		return m.manageKey(msg)
	case tui.ErrorMsg:
		m.err = error(msg)
	case tui.InfoMsg:
		m.info = string(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
		case key.Matches(msg, m.globalKeyMap.Quit):
			// Allow user to quit app whilst in filter mode. In this case,
			// switch back to normal mode, blur the filter widget, and let
			// the key handler below handle the quit action.
			m.mode = normalMode
			cmd = m.FocusedModel().Update(tui.FilterBlurMsg{})
			return m, cmd
		case key.Matches(msg, m.filterKeyMap.Blur):
			// Switch back to normal mode, and send message to current model
			// to blur the filter widget
			m.mode = normalMode
			cmd = m.FocusedModel().Update(tui.FilterBlurMsg{})
			return m, cmd
		case key.Matches(msg, m.filterKeyMap.Close):
			// Switch back to normal mode, and send message to current model
			// to close the filter widget
			m.mode = normalMode
			cmd = m.FocusedModel().Update(tui.FilterCloseMsg{})
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
	case key.Matches(msg, m.globalKeyMap.Quit):
		// ctrl-c quits the app, but not before prompting the user for
		// confirmation.
		return m, models.YesNoPrompt("Quit Shell Command Bookmarker?", tea.Quit)
	case key.Matches(msg, m.globalKeyMap.Help):
		// '?' toggles help widget
		m.showHelp = !m.showHelp
		// Help widget takes up space so update panes' dimensions
		m.PaneManager.Update(tea.WindowSizeMsg{
			Height: m.viewHeight(),
			Width:  m.viewWidth(),
		})
	case key.Matches(msg, m.globalKeyMap.Filter):
		// '/' enables filter mode if the current model indicates it
		// supports it, which it does so by sending back a non-nil command.
		if cmd = m.FocusedModel().Update(tui.FilterFocusReqMsg{}); cmd != nil {
			m.mode = filterMode
		}
		return m, cmd
	case key.Matches(msg, m.globalKeyMap.Search):
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
	if m.showHelp {
		components = append(components, m.help())
	}
	// Compose footer
	footer := m.helpWidget
	switch {
	case m.err != nil:
		footer += m.styles.FooterStyle.ErrorStyle.
			Width(m.availableFooterMsgWidth()).
			Render(m.err.Error())
	case m.info != "":
		footer += m.styles.FooterStyle.InfoStyle.
			Width(m.availableFooterMsgWidth()).
			Render(m.info)
	default:
		footer += m.styles.FooterStyle.DefaultStyle.
			Width(m.availableFooterMsgWidth()).
			Render(m.info)
	}
	footer += m.versionWidget
	// Add footer
	components = append(components, m.styles.FooterStyle.Main.
		MaxWidth(m.width).
		Width(m.width).
		Render(footer),
	)
	return strings.Join(components, "\n")
}

func (m *Model) availableFooterMsgWidth() int {
	// -2 to accommodate padding
	return max(0, m.width-lipgloss.Width(m.helpWidget)-lipgloss.Width(m.versionWidget))
}

// type taskCompletionMsg struct {
//	error

// viewHeight returns the height available to the panes
//
// TODO: rename contentHeight
func (m *Model) viewHeight() int {
	vh := m.height - m.styles.FooterStyle.Height
	if m.mode == promptMode {
		vh -= m.styles.PromptStyle.Height
	}
	if m.showHelp {
		vh -= m.styles.HelpStyle.Height
	}
	return max(m.styles.PaneStyle.MinContentHeight, vh)
}

// viewWidth retrieves the width available within the main view
//
// TODO: rename contentWidth
func (m *Model) viewWidth() int {
	return max(m.styles.PaneStyle.MinContentWidth, m.width)
}

// help renders key bindings
func (m *Model) help() string {
	// Compile list of bindings to render
	bindings := []key.Binding{m.globalKeyMap.Help, m.globalKeyMap.Quit}
	addDefaultBindings := true
	switch m.mode {
	case promptMode:
		bindings = append(bindings, m.prompt.HelpBindings()...)
		addDefaultBindings = false
	case filterMode:
		bindings = append(bindings, keys.KeyMapToSlice(*m.filterKeyMap)...)
	case normalMode:
		addDefaultBindings = true
		bindings = append(bindings, m.PaneManager.HelpBindings()...)
	}
	if addDefaultBindings {
		bindings = append(bindings, keys.KeyMapToSlice(*m.globalKeyMap)...)
		bindings = append(bindings, keys.KeyMapToSlice(*m.paneKeyMap)...)
	}
	bindings = removeDuplicateBindings(bindings)

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
func removeDuplicateBindings(bindings []key.Binding) []key.Binding {
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
