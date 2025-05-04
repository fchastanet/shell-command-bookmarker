package keys

import "github.com/charmbracelet/bubbles/key"

type EditorKeyMap struct {
	PreviousField *key.Binding
	NextField     *key.Binding
	PreviousPage  *key.Binding
	NextPage      *key.Binding
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
	previousPage := key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("⇞", "previous page"),
	)
	nextPage := key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("⇟", "next page"),
	)

	return &EditorKeyMap{
		PreviousField: &upKey,
		NextField:     &downKey,
		Save:          &saveKey,
		Cancel:        &cancelKey,
		PreviousPage:  &previousPage,
		NextPage:      &nextPage,
	}
}
