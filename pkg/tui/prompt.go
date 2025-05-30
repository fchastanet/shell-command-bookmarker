package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// YesNoPromptMsg enables the prompt widget.
type YesNoPromptMsg struct {
	form      *huh.Form
	yesAction PromptAction
}

type PromptAction func() tea.Cmd

// YesNoPrompt sends a message to enable the prompt widget, specifically
// asking the user for a yes/no answer. If yes is given then the action is
// invoked.
func YesNoPrompt(
	prompt string,
	keyMap *huh.KeyMap,
	yesAction PromptAction,
) tea.Cmd {
	group := huh.NewGroup(
		huh.NewConfirm().
			Title(prompt).
			Key("confirmKey").
			Affirmative("Yes!").
			Negative("No."),
	)
	form := huh.NewForm(group)
	form.WithKeyMap(keyMap)
	return CmdHandler(YesNoPromptMsg{
		form:      form,
		yesAction: yesAction,
	})
}

func (m YesNoPromptMsg) IsCompleted() bool {
	return m.form.State != huh.StateNormal
}

func (m YesNoPromptMsg) Init() tea.Cmd {
	return m.form.Init()
}

func (m YesNoPromptMsg) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	_, cmd := m.form.Update(msg)
	cmds = append(cmds, cmd)
	if m.form.State != huh.StateNormal {
		if m.form.GetBool("confirmKey") || m.form.State == huh.StateAborted {
			cmds = append(cmds, m.yesAction())
		}
	}
	return tea.Batch(cmds...)
}

func (m YesNoPromptMsg) View() string {
	formView := m.form.View()
	// Exclude inlined help
	lines := strings.Split(formView, "\n")
	return strings.Join(lines[:len(lines)-2], "\n")
}
