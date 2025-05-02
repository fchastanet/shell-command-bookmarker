package keys

import "github.com/charmbracelet/bubbles/key"

type EditorKeyMap struct {
	PreviousField *key.Binding
	NextField     *key.Binding
	Save          *key.Binding
	Cancel        *key.Binding
}

// HelpBindings returns the key bindings for this model
func GetDefaultEditorKeyMap() *EditorKeyMap {
	upKey := key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	)
	downKey := key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	)
	saveKey := key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "save"),
	)
	cancelKey := key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	)

	return &EditorKeyMap{
		PreviousField: &upKey,
		NextField:     &downKey,
		Save:          &saveKey,
		Cancel:        &cancelKey,
	}
}
