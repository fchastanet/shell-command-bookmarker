// Package editor provides command editing functionality
package command

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

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
	// Convert the resource ID to a command ID (uint)
	commandID := uint(id)

	// Number of input fields
	const numInputFields = 3 // Title, Description, Script

	// Load the command from the database - Safely convert uint to int
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

	// Add help text at the bottom
	helpText := "\n" + m.styles.EditorStyle.HelpText.Render("↑/↓: Navigate • Enter: Save • Esc: Cancel")

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
		// Update command in database
		err := m.DBService.UpdateCommand(m.command)
		if err != nil {
			slog.Error("Failed to save command", "id", m.command.ID, "error", err)
			return tui.ReportError(err)
		}

		return tui.ReportInfo("Command #%d saved successfully", m.command.ID)
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
