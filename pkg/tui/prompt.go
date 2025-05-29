package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PromptMsg enables the prompt widget.
type PromptMsg struct {
	// Action to carry out when key is pressed.
	Action PromptAction
	// Key that when pressed triggers the action and closes the prompt.
	Key *key.Binding
	// Cancel is a key that when pressed skips the action and closes the prompt.
	Cancel *key.Binding
	// Prompt to display to the user.
	Prompt string
	// Set initial value for the user to edit.
	InitialValue string
	// Set placeholder text in prompt
	Placeholder string
	// CancelAnyOther, if true, checks if any key other than that specified in
	// Key is pressed. If so then the action is skipped and the prompt is
	// closed. Overrides Cancel key binding.
	CancelAnyOther bool
}

type PromptAction func(text string) tea.Cmd

type PromptStyle struct {
	ThickBorder *lipgloss.Style
	Regular     *lipgloss.Style
	PlaceHolder *lipgloss.Style
	Height      int
}

// YesNoPrompt sends a message to enable the prompt widget, specifically
// asking the user for a yes/no answer. If yes is given then the action is
// invoked.
func YesNoPrompt(prompt string, quit bool, action tea.Cmd) tea.Cmd {
	cancel := key.NewBinding(key.WithKeys("n", "N"))
	yes := key.NewBinding(key.WithKeys("y", "Y"), key.WithHelp("y", "confirm"))
	if quit {
		yes = key.NewBinding(key.WithKeys("y", "Y", "ctrl+c"), key.WithHelp("y/ctrl+c", "confirm and quit"))
	}

	return CmdHandler(PromptMsg{
		Prompt:         prompt + " (y/N): ",
		InitialValue:   "",
		Placeholder:    "",
		Cancel:         &cancel,
		CancelAnyOther: true,
		Action: func(_ string) tea.Cmd {
			return action
		},
		Key: &yes,
	})
}
