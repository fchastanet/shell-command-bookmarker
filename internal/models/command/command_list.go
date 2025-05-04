package command

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/structure"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	dbmodels "github.com/fchastanet/shell-command-bookmarker/internal/services/models"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui/table"
)

type ListMaker struct {
	App              *services.AppService
	NavigationKeyMap *table.Navigation
	ActionKeyMap     *table.Action
	Styles           *styles.Styles
	Spinner          *spinner.Model
}

const (
	idColumnPercentWidth         = 3
	titleColumnPercentWidth      = 19
	scriptColumnPercentWidth     = 65
	statusColumnPercentWidth     = 7
	lintStatusColumnPercentWidth = 6

	percent    = 100
	sidesCount = 2
)

func (mm *ListMaker) Make(_ resource.ID, width, height int) (structure.ChildModel, error) {
	idColumn := table.Column{
		Key:            "id",
		Title:          "Id",
		FlexFactor:     0,
		Width:          0,
		TruncationFunc: table.NoTruncate,
		RightAlign:     false,
	}
	titleColumn := table.Column{
		Key:            "title",
		Title:          "Title",
		FlexFactor:     0,
		Width:          0,
		TruncationFunc: table.GetDefaultTruncationFunc(),
		RightAlign:     false,
	}
	scriptColumn := table.Column{
		Key:            "script",
		Title:          "Script",
		FlexFactor:     0,
		Width:          0,
		TruncationFunc: table.GetDefaultTruncationFunc(),
		RightAlign:     false,
	}
	statusColumn := table.Column{
		Key:            "status",
		Title:          "Status",
		FlexFactor:     0,
		Width:          0,
		TruncationFunc: table.GetDefaultTruncationFunc(),
		RightAlign:     false,
	}
	lintStatusColumn := table.Column{
		Key:            "lintStatus",
		Title:          "Lint",
		FlexFactor:     0,
		Width:          0,
		TruncationFunc: table.GetDefaultTruncationFunc(),
		RightAlign:     false,
	}

	m := &commandsList{
		AppService:       mm.App,
		Model:            nil,
		reloading:        false,
		spinner:          mm.Spinner,
		width:            width,
		height:           height,
		styles:           mm.Styles,
		idColumn:         &idColumn,
		titleColumn:      &titleColumn,
		scriptColumn:     &scriptColumn,
		statusColumn:     &statusColumn,
		lintStatusColumn: &lintStatusColumn,
	}
	renderer := func(cmd *dbmodels.Command) table.RenderedRow {
		return mm.renderRow(cmd, m)
	}
	tbl := table.New(
		mm.Styles.TableStyle,
		m.getColumns(0),
		renderer,
		width,
		height,
		table.WithSortFunc(dbmodels.CommandSorter),
		table.WithPreview[*dbmodels.Command](structure.CommandKind),
		table.WithNavigation[*dbmodels.Command](mm.NavigationKeyMap),
		table.WithAction[*dbmodels.Command](mm.ActionKeyMap),
	)
	m.Model = &tbl
	return m, nil
}

func (mm *ListMaker) renderRow(
	cmd *dbmodels.Command,
	commandsListModel *commandsList,
) table.RenderedRow {
	return table.RenderedRow{
		commandsListModel.idColumn.Key:         fmt.Sprintf("%d", cmd.GetID()),
		commandsListModel.titleColumn.Key:      cmd.Title,
		commandsListModel.scriptColumn.Key:     cmd.Script,
		commandsListModel.statusColumn.Key:     formatStatus(cmd, commandsListModel.styles.EditorStyle),
		commandsListModel.lintStatusColumn.Key: formatLintStatus(cmd, commandsListModel.styles.EditorStyle),
	}
}

func formatStatus(
	cmd *dbmodels.Command,
	editorStyle *styles.EditorStyle,
) string {
	switch cmd.Status {
	case dbmodels.CommandStatusBookmarked:
		return editorStyle.StatusOK.Render(string(cmd.Status))
	case dbmodels.CommandStatusSaved:
		return editorStyle.StatusOK.Render(string(cmd.Status))
	case dbmodels.CommandStatusImported:
		return editorStyle.ReadonlyValue.Render(string(cmd.Status))
	case dbmodels.CommandStatusObsolete:
		return editorStyle.StatusDisabled.Render(string(cmd.Status))
	case dbmodels.CommandStatusArchived:
		return editorStyle.StatusDisabled.Render(string(cmd.Status))
	case dbmodels.CommandStatusDeleted:
		return editorStyle.StatusWarning.Render(string(cmd.Status))
	default:
		return string(cmd.Status)
	}
}

func formatLintStatus(
	cmd *dbmodels.Command,
	editorStyle *styles.EditorStyle,
) string {
	switch cmd.LintStatus {
	case dbmodels.LintStatusOK:
		return editorStyle.StatusOK.Render("OK")
	case dbmodels.LintStatusWarning:
		return editorStyle.StatusWarning.Render("Warning")
	case dbmodels.LintStatusError:
		return editorStyle.StatusError.Render("Error")
	case dbmodels.LintStatusShellcheckFailed:
		return editorStyle.StatusError.Render("Shellcheck Failed")
	case dbmodels.LintStatusNotAvailable:
		return editorStyle.StatusDisabled.Render("Not Available")
	default:
		return editorStyle.StatusDisabled.Render("Not Available")
	}
}

type commandsList struct {
	Model *table.Model[*dbmodels.Command]
	*services.AppService
	styles  *styles.Styles
	spinner *spinner.Model

	idColumn         *table.Column
	titleColumn      *table.Column
	scriptColumn     *table.Column
	statusColumn     *table.Column
	lintStatusColumn *table.Column

	reloading bool
	height    int
	width     int
}

func (m *commandsList) getColumns(width int) []table.Column {
	slog.Debug("getColumns", "width", width)
	const columnsCount = 5
	w := width -
		columnsCount*m.styles.TableStyle.Cell.GetHorizontalPadding()*sidesCount
	m.idColumn.Width = idColumnPercentWidth * w / percent
	m.titleColumn.Width = titleColumnPercentWidth * w / percent
	m.scriptColumn.Width = scriptColumnPercentWidth * w / percent
	m.statusColumn.Width = statusColumnPercentWidth * w / percent
	m.lintStatusColumn.Width = lintStatusColumnPercentWidth * w / percent
	return []table.Column{
		*m.idColumn,
		*m.titleColumn,
		*m.scriptColumn,
		*m.statusColumn,
		*m.lintStatusColumn,
	}
}

func (m *commandsList) Init() tea.Cmd {
	return func() tea.Msg {
		return tea.FocusMsg{}
	}
}

func (m *commandsList) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case table.ReloadMsg[*dbmodels.Command]:
		return m.handleReload(msg)
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.BlurMsg:
		m.Model.Blur()
	case tea.FocusMsg:
		return m.handleFocus()
	case table.RowDeleteActionMsg[*dbmodels.Command]:
		return m.handleDeleteRow(msg)
	}

	// Handle keyboard and mouse events in the table widget
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m *commandsList) handleWindowSize(msg tea.WindowSizeMsg) tea.Cmd {
	m.width = msg.Width
	m.height = msg.Height
	m.Model.SetWidth(m.width)
	m.Model.SetHeight(m.height)
	m.Model.SetColumns(m.getColumns(m.width))
	return nil
}

func (m *commandsList) handleFocus() tea.Cmd {
	m.Model.SetColumns(m.getColumns(m.width))
	return func() tea.Msg {
		rows, err := m.HistoryService.GetHistoryRows()
		if err != nil {
			slog.Error("Error getting history rows", "error", err)
			return nil
		}
		m.Model.Focus()

		return table.BulkInsertMsg[*dbmodels.Command](rows)
	}
}

func (m *commandsList) handleReload(msg table.ReloadMsg[*dbmodels.Command]) tea.Cmd {
	if m.reloading {
		return nil
	}
	m.reloading = true
	return tea.Batch(
		tui.ReportInfo("reloading started"),
		func() tea.Msg {
			defer func() {
				m.reloading = false
			}()
			rows, err := m.HistoryService.GetHistoryRows()
			if err != nil {
				return tui.ErrorMsg(fmt.Errorf("reloading state failed: %w", err))
			}
			m.Model.SetItems(rows...)

			if msg.RowID != -1 {
				m.Model.GotoID(msg.RowID)
				if msg.InfoMsg != nil {
					return *msg.InfoMsg
				}
				return tui.InfoMsg(fmt.Sprintf(
					"reloading finished, selected new item %d", msg.RowID,
				))
			}

			if msg.InfoMsg != nil {
				return *msg.InfoMsg
			}
			return tui.InfoMsg("reloading finished")
		},
	)
}

// handleDeleteRow handles a request to delete a command row
func (m *commandsList) handleDeleteRow(msg table.RowDeleteActionMsg[*dbmodels.Command]) tea.Cmd {
	cmd := msg.Row
	if cmd == nil {
		return nil
	}

	confirmMessage := fmt.Sprintf("Delete command #%d: %s?", cmd.GetID(), cmd.Title)

	// Pass our wrapper function as the action to YesNoPrompt
	return tui.YesNoPrompt(confirmMessage, false, func() tea.Msg {
		nextRowID := m.Model.GetNextRowIDRelativeToCurrentRow()
		// Mark the command as deleted in the database
		originalStatus := cmd.Status
		cmd.Status = dbmodels.CommandStatusDeleted
		err := m.DBService.UpdateCommand(cmd)
		if err != nil {
			slog.Error("Error marking command as deleted", "error", err, "id", cmd.GetID())
			// Revert status change if update fails
			cmd.Status = originalStatus
			return tui.ErrorMsg(fmt.Errorf("failed to mark command as deleted: %w", err))
		}

		// Return a message that will trigger the reload
		infoMsg := tui.InfoMsg(fmt.Sprintf(
			"Command #%d marked as deleted", cmd.GetID(),
		))
		return table.ReloadMsg[*dbmodels.Command]{
			RowID:   nextRowID,
			InfoMsg: &infoMsg,
		}
	})
}

func (m *commandsList) View() string {
	if m.reloading {
		return "Pulling state " + m.spinner.View()
	}
	return m.Model.View()
}
