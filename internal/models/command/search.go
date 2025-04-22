package command

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/structure"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
)

type SearchMaker struct {
	App     *services.AppService
	Styles  *styles.Styles
	Spinner *spinner.Model
}

func (mm *SearchMaker) Make(_ resource.ID, width, height int) (structure.ChildModel, error) {
	return &search{
		App:     mm.App,
		Spinner: mm.Spinner,
		Styles:  mm.Styles,
		width:   width,
		height:  height,
		model:   nil,
	}, nil
}

type search struct {
	App     *services.AppService
	Spinner *spinner.Model
	Styles  *styles.Styles
	width   int
	height  int
	model   *textinput.Model
}

func (m *search) Init() tea.Cmd {
	model := textinput.New()
	model.Prompt = "Search: "
	model.SetValue("")
	model.Placeholder = ""
	model.PlaceholderStyle = *m.Styles.PromptStyle.PlaceHolder
	m.model = &model
	return tea.Batch()
}

func (m *search) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.FocusMsg:
		if m.model != nil {
			m.model.Focus()
		}
	case tea.BlurMsg:
		if m.model != nil {
			m.model.Blur()
		}
	}

	return tea.Batch(cmds...)
}

func (m *search) View() string {
	if m.model == nil {
		return ""
	}
	return m.model.View()
}

func (m *search) HelpBindings() []key.Binding {
	bindings := []key.Binding{}
	return bindings
}
