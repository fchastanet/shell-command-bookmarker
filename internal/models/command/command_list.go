package command

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/keys"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/structure"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui/table"
)

type ListMaker struct {
	App     *services.AppService
	Styles  *styles.Styles
	Spinner *spinner.Model
}

func (mm *ListMaker) Make(_ resource.ID, width, height int) (structure.ChildModel, error) {
	commandColumn := table.Column{
		Key:            "command",
		Title:          "COMMAND",
		FlexFactor:     1,
		Width:          0,
		TruncationFunc: table.GetDefaultTruncationFunc(),
		RightAlign:     false,
	}

	columns := []table.Column{commandColumn}
	renderer := func(resource resource.Resource) table.RenderedRow {
		addr := fmt.Sprintf("%d", resource.GetMonotonicID().Serial)
		return table.RenderedRow{commandColumn.Key: addr}
	}
	tbl := table.New(
		mm.Styles.TableStyle,
		columns,
		renderer,
		width,
		height,
		table.WithSortFunc(resource.Sort),
		table.WithPreview[resource.Resource](structure.CommandKind),
	)
	m := &resourceList{
		AppService: mm.App,
		Model:      tbl,
		reloading:  false,
		spinner:    mm.Spinner,
		width:      width,
		height:     height,
	}
	return m, nil
}

type resourceList struct {
	table.Model[resource.Resource]
	*services.AppService

	reloading bool
	height    int
	width     int

	spinner *spinner.Model
}

func (m *resourceList) Init() tea.Cmd {
	return tea.Batch()
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

func (m *resourceList) HelpBindings() []key.Binding {
	resourcesKeys := GetResourcesKeyMap()
	commonKeys := keys.GetCommonKeyMap()
	bindings := []key.Binding{
		commonKeys.Delete,
		*resourcesKeys.Move,
		*resourcesKeys.Reload,
	}
	return bindings
}
