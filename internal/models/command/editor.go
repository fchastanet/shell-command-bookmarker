// Package editor provides command editing functionality
package command

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/keys"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/structure"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	dbmodels "github.com/fchastanet/shell-command-bookmarker/internal/services/models"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui/table"
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

// Make creates a new command editor model based on the command ID
func (mm *EditorMaker) Make(id resource.ID, width, height int) (structure.ChildModel, error) {
	// Convert the resource ID to a command ID
	commandID := id

	// Number of input fields
	const numInputFields = 3 // Title, Description, Script

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
		AppService:   mm.App,
		styles:       mm.Styles,
		command:      command,
		width:        width,
		height:       height,
		inputs:       make([]textinput.Model, numInputFields),
		focused:      0,
		EditorKeyMap: mm.EditorKeyMap,
	}, nil
}

// commandEditor is the model for editing a command
type commandEditor struct {
	*services.AppService
	styles       *styles.Styles
	command      *dbmodels.Command
	EditorKeyMap *keys.EditorKeyMap
	inputs       []textinput.Model
	width        int
	height       int
	focused      int
}

// Init initializes the command editor
func (m *commandEditor) Init() tea.Cmd {
	// Initialize the text inputs
	titleInput := textinput.New()
	titleInput.Placeholder = "Enter title"
	titleInput.SetValue(m.command.Title)
	titleInput.Focus()

	descriptionInput := textinput.New()
	descriptionInput.Placeholder = "Enter description"
	descriptionInput.SetValue(m.command.Description)
	descriptionInput.CharLimit = 200

	scriptInput := textinput.New()
	scriptInput.Placeholder = "Enter script"
	scriptInput.SetValue(m.command.Script)

	m.inputs = []textinput.Model{titleInput, descriptionInput, scriptInput}

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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, *m.EditorKeyMap.PreviousField):
			m.prevField()
		case key.Matches(msg, *m.EditorKeyMap.NextField):
			m.nextField()
		case key.Matches(msg, *m.EditorKeyMap.Save):
			return m.save()
		case key.Matches(msg, *m.EditorKeyMap.Cancel):
			// Return without saving
			return m.cancel()
		}
	}

	// Update the active input
	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// View renders the command editor
func (m *commandEditor) View() string {
	title := m.styles.EditorStyle.Title.Render("Command Editor")

	// Labels for our fields
	labels := []string{"Title:", "Description:", "Script:"}

	var content strings.Builder
	content.WriteString(title + "\n\n")

	// Render each field with its label
	for i, label := range labels {
		// Style the label
		styledLabel := m.styles.EditorStyle.Label.Render(label)

		// Render the input field
		content.WriteString(fmt.Sprintf("%s\n%s\n\n", styledLabel, m.inputs[i].View()))
	}

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
	content.WriteString(fmt.Sprintf("%s %s\n", createLabel, createValue))
	content.WriteString(fmt.Sprintf("%s %s\n", modifyLabel, modifyValue))
	content.WriteString(fmt.Sprintf("%s %s\n", lintStatusLabel, m.formatLintStatus()))

	// Parse and display lint issues
	issues := m.command.GetLintIssues()
	if len(issues) == 0 {
		content.WriteString(fmt.Sprintf("%s %s\n\n", lintIssuesLabel,
			m.styles.EditorStyle.ReadonlyValue.Render("None")))
	} else {
		content.WriteString(fmt.Sprintf("%s %s\n", lintIssuesLabel,
			m.styles.EditorStyle.ReadonlyValue.Render(fmt.Sprintf("%d issues found:", len(issues)))))

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
			var styledMessage string
			switch level {
			case "error":
				styledMessage = m.styles.EditorStyle.StatusError.Render(message)
			case "warning":
				styledMessage = m.styles.EditorStyle.StatusWarning.Render(message)
			case "info":
				styledMessage = m.styles.EditorStyle.StatusOK.Render(message)
			default:
				styledMessage = m.styles.EditorStyle.ReadonlyValue.Render(message)
			}

			content.WriteString(fmt.Sprintf("   %s %s\n", num, styledMessage))
		}
		content.WriteString("\n")
	}

	// Add help text at the bottom
	helpText := m.styles.EditorStyle.HelpText.Render("↑/↓: Navigate • Enter: Save • Esc: Cancel")
	content.WriteString(helpText)

	return content.String()
}

// nextField focuses the next field
func (m *commandEditor) nextField() {
	m.inputs[m.focused].Blur()
	m.focused = (m.focused + 1) % len(m.inputs)
	m.inputs[m.focused].Focus()
}

// prevField focuses the previous field
func (m *commandEditor) prevField() {
	m.inputs[m.focused].Blur()
	m.focused = (m.focused - 1 + len(m.inputs)) % len(m.inputs)
	m.inputs[m.focused].Focus()
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
			RowID:   resource.ID(m.command.ID),
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
