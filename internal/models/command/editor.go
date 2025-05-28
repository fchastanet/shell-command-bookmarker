// Package editor provides command editing functionality
package command

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/command/inputs"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/keys"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/structure"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	dbmodels "github.com/fchastanet/shell-command-bookmarker/internal/services/models"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui/table"
	"github.com/fchastanet/shell-command-bookmarker/pkg/utils"
)

// Error definitions
var (
	// ErrCommandNotFound is returned when a command with the given ID is not found
	ErrCommandNotFound = errors.New("command not found with ID")
)

type EditorMaker struct {
	App          *services.AppService
	Styles       *styles.Styles
	EditorKeyMap *keys.EditorKeyMap
}

// Number of input fields
const (
	numInputFields           = 3    // Title, Description, Script
	titleInputMaxSize        = 50   // Max size for title input
	descriptionInputMaxSize  = 1000 // Max size for description input
	descriptionInputHeight   = 5    // Height for description input
	scriptInputHeight        = 5    // Height for script input
	descriptionWordwrapWidth = 80   // Word wrap width for description input
	inputFieldPadding        = 2
)

// Make creates a new command editor model based on the command ID
func (mm *EditorMaker) Make(id resource.ID, width, height int) (structure.ChildModel, error) {
	// Convert the resource ID to a command ID
	commandID := id

	// Load the command from the database
	command, err := mm.App.DBService.GetCommandByID(commandID)
	if err != nil {
		return nil, fmt.Errorf("failed to load command %d: %w", commandID, err)
	}

	if command == nil {
		return nil, fmt.Errorf("%w %d", ErrCommandNotFound, commandID)
	}

	// Create the editor model
	return &commandEditor{
		AppService:    mm.App,
		styles:        mm.Styles,
		command:       command,
		width:         width,
		height:        height,
		inputs:        make([]inputs.Input, numInputFields),
		focused:       -1,
		EditorKeyMap:  mm.EditorKeyMap,
		pagePosition:  0,
		contentHeight: 0,
	}, nil
}

// commandEditor is the model for editing a command
type commandEditor struct {
	*services.AppService
	styles        *styles.Styles
	command       *dbmodels.Command
	EditorKeyMap  *keys.EditorKeyMap
	inputs        []inputs.Input
	width         int
	height        int
	focused       int
	pagePosition  int
	contentHeight int
}

// Init initializes the command editor
func (m *commandEditor) Init() tea.Cmd {
	// Initialize the text inputs
	titleInput := inputs.NewInputWrapper("Enter title")
	titleInput.SetCharLimit(titleInputMaxSize)
	titleInput.SetValue(m.command.Title)
	titleInput.Focus()

	descriptionInput := inputs.NewTextAreaWrapper(
		descriptionInputHeight,
		"Enter description (markdown)",
		inputs.WithMarkdown(descriptionWordwrapWidth))
	descriptionInput.SetCharLimit(descriptionInputMaxSize)
	descriptionInput.SetValue(m.command.Description)

	scriptInput := inputs.NewTextAreaWrapper(scriptInputHeight, "Enter script")
	scriptInput.SetValue(m.command.Script)

	m.inputs = []inputs.Input{titleInput, descriptionInput, scriptInput}

	return nil
}

// formatLintStatus returns a styled string representing the lint status
func (m *commandEditor) formatLintStatus() string {
	switch m.command.LintStatus {
	case dbmodels.LintStatusOK:
		return m.styles.EditorStyle.StatusOK.Render("OK")
	case dbmodels.LintStatusWarning:
		return m.styles.EditorStyle.StatusWarning.Render("Warning")
	case dbmodels.LintStatusError:
		return m.styles.EditorStyle.StatusError.Render("Error")
	case dbmodels.LintStatusShellcheckFailed:
		return m.styles.EditorStyle.StatusError.Render("Shellcheck Failed")
	case dbmodels.LintStatusNotAvailable:
		return m.styles.EditorStyle.StatusDisabled.Render("Not Available")
	default:
		return m.styles.EditorStyle.StatusDisabled.Render("Not Available")
	}
}

// Update handles updates to the command editor
func (m *commandEditor) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	slog.Debug("commandEditor.Update", "msgType", fmt.Sprintf("%T", msg), "focused", m.focused)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleWindowSizeMsg(msg)
	case structure.FocusedPaneChangedMsg:
		cmds = append(cmds, m.handleFocusedPaneChangedMsg(msg))
	case table.RowSelectedActionMsg[*dbmodels.Command]:
		// This message is sent when a row is selected in the command table
		m.command = msg.Row
		m.inputs[0].SetValue(m.command.Title)
		m.inputs[1].SetValue(m.command.Description)
		m.inputs[2].SetValue(m.command.Script)
	case tea.KeyMsg:
		cmd := m.handleKeyMsg(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
			return tea.Batch(cmds...)
		}
		if msg.Alt {
			// avoid alt key combinations to be eaten by inputs
			return nil
		}
	}

	// Update the active input field
	if m.focused >= 0 {
		var cmd tea.Cmd
		m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (m *commandEditor) handleWindowSizeMsg(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	for _, input := range m.inputs {
		input.SetWidth(m.width - inputFieldPadding)
	}
}

func (m *commandEditor) handleFocusedPaneChangedMsg(msg structure.FocusedPaneChangedMsg) tea.Cmd {
	if m.focused >= 0 && m.inputs[m.focused] != nil {
		m.inputs[m.focused].Blur()
	}
	if msg.To == structure.BottomPane {
		m.focused = -1
	}
	for _, input := range m.inputs {
		input.SetReadOnly(msg.To != structure.BottomPane)
	}
	return tui.GetDummyCmd()
}

func (m *commandEditor) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	var cmds []tea.Cmd

	switch {
	case key.Matches(msg, *m.EditorKeyMap.PreviousField):
		cmds = append(cmds, m.prevField())
	case key.Matches(msg, *m.EditorKeyMap.NextField):
		cmds = append(cmds, m.nextField())
	case key.Matches(msg, *m.EditorKeyMap.PreviousPage):
		m.prevPage()
	case key.Matches(msg, *m.EditorKeyMap.NextPage):
		m.nextPage()
	case key.Matches(msg, *m.EditorKeyMap.Save):
		return m.save()
	case key.Matches(msg, *m.EditorKeyMap.Cancel):
		return m.cancel()
	}

	return tea.Batch(cmds...)
}

// View renders the command editor
func (m *commandEditor) View() string {
	slog.Debug("commandEditor.View", "focused", m.focused)
	var content strings.Builder

	// Add common elements first
	m.addCommonElements(&content)

	// Add readonly information section
	m.addReadonlySection(&content)

	// strip the first lines of content to fit the screen
	contentStr := content.String()
	visibleContent := utils.RemoveFirstLines(contentStr, m.pagePosition)

	// Create the content area with scrollbar
	contentArea := lipgloss.NewStyle().
		Width(m.width - m.styles.ScrollbarStyle.Width).
		Render(visibleContent + "\n")

	// Generate scrollbar
	scrollbar := m.generateScrollbar(contentStr, visibleContent)

	contentScroll := lipgloss.NewStyle().
		Height(m.height).
		MaxHeight(m.height).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, contentArea, scrollbar))

	return contentScroll
}

// addCommonElements adds the title, help text, and input fields to the content
func (m *commandEditor) addCommonElements(content *strings.Builder) {
	title := m.styles.EditorStyle.Title.Render("Command Editor")

	// Add help text at the top
	helpTextStyle := *m.styles.EditorStyle.HelpText
	if m.focused == -1 {
		helpTextStyle = helpTextStyle.Bold(true).Foreground(lipgloss.Color("255"))
	}
	helpText := helpTextStyle.Render("⭾/Shift-⭾: Fields • ⇞/⇟: Scroll • Ctrl+S: Save • Esc: Cancel")
	content.WriteString(helpText)

	// Add the title
	content.WriteString(title + "\n\n")

	// Labels for our fields
	labels := []string{"Title:", "Description:", "Script:"}

	// Render each field with its label
	for i, label := range labels {
		labelStyle := m.styles.EditorStyle.Label
		if m.focused == i {
			labelStyle = m.styles.EditorStyle.LabelFocused
		}
		// Style the label
		styledLabel := labelStyle.Render(label)

		// Render the input field
		fmt.Fprintf(content, "%s\n%s\n\n", styledLabel, m.inputs[i].View())
	}
}

// addReadonlySection adds the readonly information section to the content
func (m *commandEditor) addReadonlySection(content *strings.Builder) {
	// Add readonly information section
	readonlyTitle := m.styles.EditorStyle.Label.Render("Readonly Information:")
	content.WriteString(readonlyTitle + "\n\n")

	// Format and add each readonly field with label and value
	createLabel := m.styles.EditorStyle.ReadonlyLabel.Render("Created:")
	createValue := m.styles.EditorStyle.ReadonlyValue.Render(
		m.command.CreationDatetime.Format(time.DateTime))

	modifyLabel := m.styles.EditorStyle.ReadonlyLabel.Render("Modified:")
	modifyValue := m.styles.EditorStyle.ReadonlyValue.Render(
		m.command.ModificationDatetime.Format(time.DateTime))

	lintStatusLabel := m.styles.EditorStyle.ReadonlyLabel.Render("Lint Status:")
	lintIssuesLabel := m.styles.EditorStyle.ReadonlyLabel.Render("Lint Issues:")

	// Add the formatted readonly information
	fmt.Fprintf(content, "%s %s\n", createLabel, createValue)
	fmt.Fprintf(content, "%s %s\n", modifyLabel, modifyValue)
	fmt.Fprintf(content, "%s %s\n", lintStatusLabel, m.formatLintStatus())

	m.addLintIssues(content, lintIssuesLabel)
}

// addLintIssues adds the lint issues section to the content
func (m *commandEditor) addLintIssues(content *strings.Builder, lintIssuesLabel string) {
	// Parse and display lint issues
	issues := m.command.GetLintIssues()
	if len(issues) == 0 {
		fmt.Fprintf(content, "%s %s\n\n", lintIssuesLabel,
			m.styles.EditorStyle.ReadonlyValue.Render("None"))
		return
	}

	fmt.Fprintf(content, "%s %s\n", lintIssuesLabel,
		m.styles.EditorStyle.ReadonlyValue.Render(fmt.Sprintf("%d issues found:", len(issues))))

	// Display each lint issue
	for i, issue := range issues {
		// Format the issue number
		num := m.styles.EditorStyle.ReadonlyLabel.Render(fmt.Sprintf("%d.", i+1))

		// Extract and format the message
		message := "Unknown issue"
		if msg, ok := issue["message"].(string); ok {
			message = msg
		}

		// Extract and format the level
		level := "unknown"
		if lvl, ok := issue["level"].(string); ok {
			level = lvl
		}

		// Style based on level
		styledMessage := m.getStyledMessage(level, message)
		fmt.Fprintf(content, "   %s %s %s\n", num, level, styledMessage)
	}
	content.WriteString("\n")
}

// getStyledMessage returns styled message based on issue level
func (m *commandEditor) getStyledMessage(level, message string) string {
	switch level {
	case "error":
		return m.styles.EditorStyle.StatusError.Render(message)
	case "warning":
		return m.styles.EditorStyle.StatusWarning.Render(message)
	case "info":
		return m.styles.EditorStyle.StatusOK.Render(message)
	default:
		return m.styles.EditorStyle.ReadonlyValue.Render(message)
	}
}

// generateScrollbar creates the scrollbar for the editor
func (m *commandEditor) generateScrollbar(contentStr, visibleContent string) string {
	// Generate scrollbar
	const minEditorHeight = 15 // Minimum height for the editor to ensure usability
	availableHeight := max(minEditorHeight, m.height)
	m.contentHeight = lipgloss.Height(contentStr)
	visibleContentHeight := lipgloss.Height(visibleContent)
	return tui.Scrollbar(
		m.styles.EditorStyle.ScrollbarStyle,
		availableHeight,
		m.contentHeight,
		min(availableHeight, visibleContentHeight),
		m.pagePosition,
	)
}

// nextField focuses the next field
func (m *commandEditor) nextField() tea.Cmd {
	if m.focused >= 0 {
		m.inputs[m.focused].Blur()
	}
	if m.focused == len(m.inputs)-1 {
		m.focused = -1
		return nil
	}

	m.focused = (m.focused + 1) % len(m.inputs)

	return m.inputs[m.focused].Focus()
}

// prevField focuses the previous field
func (m *commandEditor) prevField() tea.Cmd {
	if m.focused >= 0 {
		m.inputs[m.focused].Blur()
	}
	switch m.focused {
	case -1:
		m.focused = len(m.inputs) - 1
		return nil
	case 0:
		m.focused = -1
	default:
		m.focused = (m.focused - 1) % len(m.inputs)
	}
	if m.focused >= 0 {
		return m.inputs[m.focused].Focus()
	}
	return tui.GetDummyCmd()
}

// nextPage scrolls to the next page
func (m *commandEditor) nextPage() {
	slog.Debug("nextPage", "pagePosition", m.pagePosition, "height", m.height, "contentHeight", m.contentHeight)
	m.pagePosition = min(m.pagePosition+m.height, m.contentHeight-m.height)
	slog.Debug("nextPage", "newPagePosition", m.pagePosition)
}

// prevPage scrolls to the previous page
func (m *commandEditor) prevPage() {
	slog.Debug("prevPage", "pagePosition", m.pagePosition, "height", m.height, "contentHeight", m.contentHeight)
	m.pagePosition = max(m.pagePosition-m.height, 0)
	slog.Debug("prevPage", "newPagePosition", m.pagePosition)
}

func (m *commandEditor) EditionInProgress() bool {
	return m.command.Title != m.inputs[0].Value() ||
		m.command.Description != m.inputs[1].Value() ||
		m.command.Script != m.inputs[2].Value()
}

// save saves the current command
func (m *commandEditor) save() tea.Cmd {
	// Update the command with values from the input fields
	oldTitle := m.command.Title
	oldDescription := m.command.Description
	oldScript := m.command.Script

	m.command.Title = m.inputs[0].Value()
	m.command.Description = m.inputs[1].Value()
	m.command.Script = m.inputs[2].Value()

	// Only update if there are actual changes
	if oldTitle != m.command.Title ||
		oldDescription != m.command.Description ||
		oldScript != m.command.Script {
		// Update command in database using HistoryService
		newCommand, err := m.HistoryService.UpdateCommand(m.command)
		if err != nil {
			slog.Error("Failed to save command", "id", m.command.ID, "error", err)
			return tui.ReportError(err)
		}
		m.command = newCommand

		// Trigger table reload to reflect changes
		infoMsg := tui.InfoMsg(fmt.Sprintf(
			"Command #%d saved successfully", m.command.ID,
		))

		return tui.CmdHandler(table.ReloadMsg[*dbmodels.Command]{
			RowID:   m.command.ID,
			InfoMsg: &infoMsg,
		})
	}

	return tui.ReportInfo("No changes to save for command #%d", m.command.ID)
}

type EditorCancelledMsg struct{}

// cancel returns from the editor without saving
func (m *commandEditor) cancel() tea.Cmd {
	return tea.Batch(
		tui.ReportInfo("Edit cancelled for command #%d", m.command.ID),
		tui.CmdHandler(EditorCancelledMsg{}),
	)
}

// BorderText returns text to display in the border
func (m *commandEditor) BorderText() map[styles.BorderPosition]string {
	return map[styles.BorderPosition]string{
		styles.TopMiddleBorder: fmt.Sprintf("Command #%d", m.command.ID),
	}
}
