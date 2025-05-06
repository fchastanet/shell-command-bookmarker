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
	previousField := key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("Shift+⭾", "previous field"),
	)
	nextField := key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("⭾", "next field"),
	)
	saveKey := key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("Ctrl+s", "save"),
	)
	cancelKey := key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("␛", "cancel"),
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
		PreviousField: &previousField,
		NextField:     &nextField,
		Save:          &saveKey,
		Cancel:        &cancelKey,
		PreviousPage:  &previousPage,
		NextPage:      &nextPage,
	}
}
