package command

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/keys"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/structure"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	"github.com/fchastanet/shell-command-bookmarker/internal/services/models"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui/table"
)

type ListMaker struct {
	App     *services.AppService
	Styles  *styles.Styles
	Spinner *spinner.Model
}

// commandReloadedMsg is sent when a command reload has finished.
type commandReloadedMsg struct {
	err error
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

	m := &resourceList{
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

	renderer := func(cmd *models.Command) table.RenderedRow {
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
		table.WithSortFunc(models.CommandSorter),
		table.WithPreview[*models.Command](structure.CommandKind),
	)
	m.Model = &tbl
	return m, nil
}

type resourceList struct {
	Model *table.Model[*models.Command]
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

func (m *resourceList) getColumns(width int) []table.Column {
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

func (m *resourceList) Init() tea.Cmd {
	return func() tea.Msg {
		return tea.FocusMsg{}
	}
}

func (m *resourceList) Update(msg tea.Msg) tea.Cmd {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case commandReloadedMsg:
		m.reloading = false
		if msg.err != nil {
			return tui.ReportError(fmt.Errorf("reloading state failed: %w", msg.err))
		}
		return tui.ReportInfo("reloading finished")
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.Model.SetWidth(m.width)
		m.Model.SetHeight(m.height)
		m.Model.SetColumns(m.getColumns(m.width))

	case tea.BlurMsg:
		m.Model.Blur()
	case tea.FocusMsg:
		return func() tea.Msg {
			m.Model.SetColumns(m.getColumns(m.width))
			rows, err := m.HistoryService.GetHistoryRows()
			if err != nil {
				slog.Error("Error getting history rows", "error", err)
				return nil
			}
			m.Model.Focus()
			// type conversion to []*models.Command
			return table.BulkInsertMsg[*models.Command](rows)
		}
	}

	// Handle keyboard and mouse events in the table widget
	m.Model, cmd = m.Model.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m *resourceList) View() string {
	if m.reloading {
		return "Pulling state " + m.spinner.View()
	}
	return m.Model.View()
}

func (m *resourceList) HelpBindings() []*key.Binding {
	resourcesKeys := GetResourcesKeyMap()
	commonKeys := keys.GetCommonKeyMap()
	bindings := []*key.Binding{
		commonKeys.Delete,
		resourcesKeys.Move,
		resourcesKeys.Reload,
	}
	return bindings
}
