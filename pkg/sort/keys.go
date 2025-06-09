package sort

import "github.com/charmbracelet/bubbles/key"

// KeyMap contains keybindings for sort mode
type KeyMap struct {
	Sort               *key.Binding
	Apply              *key.Binding
	Cancel             *key.Binding
	NextField          *key.Binding
	PreviousField      *key.Binding
	NextComboValue     *key.Binding
	PreviousComboValue *key.Binding
}

// GetDefaultKeyMap returns the default sort key mapping
func GetDefaultKeyMap() *KeyMap {
	sort := key.NewBinding(
		key.WithKeys("s", "S"),
		key.WithHelp("s", "sort"),
	)
	apply := key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "apply sort"),
	)
	cancel := key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel sort"),
	)
	nextField := key.NewBinding(
		key.WithKeys("tab", "right"),
		key.WithHelp("tab/→", "next field"),
	)
	previousField := key.NewBinding(
		key.WithKeys("shift+tab", "left"),
		key.WithHelp("shift+tab/←", "previous field"),
	)
	nextComboValue := key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "combo next value"),
	)
	previousComboValue := key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "combo previous value"),
	)
	return &KeyMap{
		Sort:               &sort,
		Apply:              &apply,
		Cancel:             &cancel,
		NextField:          &nextField,
		PreviousField:      &previousField,
		NextComboValue:     &nextComboValue,
		PreviousComboValue: &previousComboValue,
	}
}
