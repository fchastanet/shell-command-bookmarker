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
	idColumnPercentWidth     = 3
	titleColumnPercentWidth  = 19
	scriptColumnPercentWidth = 71
	statusColumnPercentWidth = 7

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

	m := &commandsList{
		AppService:   mm.App,
		Model:        nil,
		reloading:    false,
		spinner:      mm.Spinner,
		width:        width,
		height:       height,
		styles:       mm.Styles,
		idColumn:     &idColumn,
		titleColumn:  &titleColumn,
		scriptColumn: &scriptColumn,
		statusColumn: &statusColumn,
	}

	renderer := func(cmd *dbmodels.Command) table.RenderedRow {
		return table.RenderedRow{
			idColumn.Key:     fmt.Sprintf("%d", cmd.GetID()),
			titleColumn.Key:  cmd.Title,
			scriptColumn.Key: cmd.Script,
			statusColumn.Key: string(cmd.Status),
		}
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

type commandsList struct {
	Model *table.Model[*dbmodels.Command]
	*services.AppService
	styles  *styles.Styles
	spinner *spinner.Model

	idColumn     *table.Column
	titleColumn  *table.Column
	scriptColumn *table.Column
	statusColumn *table.Column

	reloading bool
	height    int
	width     int
}

func (m *commandsList) getColumns(width int) []table.Column {
	slog.Debug("getColumns", "width", width)
	const columnsCount = 4
	w := width -
		columnsCount*m.styles.TableStyle.Cell.GetHorizontalPadding()*sidesCount
	m.idColumn.Width = idColumnPercentWidth * w / percent
	m.titleColumn.Width = titleColumnPercentWidth * w / percent
	m.scriptColumn.Width = scriptColumnPercentWidth * w / percent
	m.statusColumn.Width = statusColumnPercentWidth * w / percent
	return []table.Column{
		*m.idColumn,
		*m.titleColumn,
		*m.scriptColumn,
		*m.statusColumn,
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
		return m.handleReload()
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.BlurMsg:
		m.Model.Blur()
	case tea.FocusMsg:
		return m.handleFocus()
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

func (m *commandsList) handleReload() tea.Cmd {
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

			return tui.InfoMsg("reloading finished")
		},
	)
}

func (m *commandsList) View() string {
	if m.reloading {
		return "Pulling state " + m.spinner.View()
	}
	return m.Model.View()
}
