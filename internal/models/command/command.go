package command

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	"github.com/leg100/pug/internal/plan"
	"github.com/leg100/pug/internal/resource"
	"github.com/leg100/pug/internal/state"
	"github.com/leg100/pug/internal/task"
	"github.com/leg100/pug/internal/tui"
	"github.com/leg100/pug/internal/tui/keys"
)

type CommandMaker struct {
	App     *services.AppService
	Spinner *spinner.Model

	disableBorders bool
}

func (mm *CommandMaker) Make(id resource.ID, width, height int) (tui.ChildModel, error) {
	stateResource, err := mm.App.DBService.GetCommandById(id)
	if err != nil {
		return nil, err
	}

	m := commandModel{
		states:   mm.States,
		plans:    mm.Plans,
		Helpers:  mm.Helpers,
		resource: stateResource,
		border:   !mm.disableBorders,
	}

	marshaled, err := json.MarshalIndent(stateResource.Attributes, "", "\t")
	if err != nil {
		return nil, err
	}
	m.viewport = tui.NewViewport(tui.ViewportOptions{
		Width:  width,
		Height: height,
		JSON:   true,
	})
	m.viewport.AppendContent(marshaled, true, false)

	return &m, nil
}

type commandModel struct {
	*tui.Helpers

	states *state.Service
	plans  *plan.Service

	viewport tui.Viewport
	resource *state.Resource
	border   bool
}

func (m *commandModel) Init() tea.Cmd {
	return nil
}

func (m *commandModel) Update(msg tea.Msg) tea.Cmd {
	var (
		cmd              tea.Cmd
		cmds             []tea.Cmd
		createRunOptions plan.CreateOptions
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Common.Delete):
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return m.states.Delete(workspaceID, m.resource.Address)
			}
			return tui.YesNoPrompt(
				"Delete resource?",
				m.CreateTasks(fn, m.resource.WorkspaceID),
			)
		case key.Matches(msg, keys.Common.PlanDestroy):
			// Create a targeted destroy plan.
			createRunOptions.Destroy = true
			fallthrough
		case key.Matches(msg, keys.Common.Plan):
			// Create a targeted plan.
			createRunOptions.TargetAddrs = []state.ResourceAddress{m.resource.Address}
			fn := func(workspaceID resource.ID) (task.Spec, error) {
				return m.plans.Plan(workspaceID, createRunOptions)
			}
			return m.CreateTasks(fn, m.resource.WorkspaceID)
		}
	case tea.WindowSizeMsg:
		m.viewport.SetDimensions(msg.Width, msg.Height)
		return nil
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m *commandModel) View() string {
	return m.viewport.View()
}

func (m *commandModel) BorderText() map[tui.BorderPosition]string {
	topLeft := fmt.Sprintf("%s %s",
		tui.Bold.Render("resource"),
		m.resource,
	)
	if m.resource.Tainted {
		topLeft += lipgloss.NewStyle().
			Foreground(tui.Red).
			Render(" (tainted)")
	}
	return map[tui.BorderPosition]string{
		tui.TopLeftBorder: topLeft,
	}
}

func (m commandModel) HelpBindings() []key.Binding {
	return []key.Binding{
		keys.Common.Plan,
		keys.Common.PlanDestroy,
		keys.Common.Delete,
		resourcesKeys.Move,
		resourcesKeys.Taint,
		resourcesKeys.Untaint,
	}
}
