package sort

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
)

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
		key.WithKeys("right"),
		key.WithHelp("→", "next field"),
	)
	previousField := key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "previous field"),
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

func UpdateBindings[ElementType resource.Identifiable, FieldType string](
	k *KeyMap,
	state *State[ElementType, FieldType],
) {
	if k == nil {
		return
	}
	k.Sort.SetEnabled(state == nil || !state.IsEditActive)
	k.Apply.SetEnabled(state != nil && state.IsEditActive)
	k.Cancel.SetEnabled(state != nil && state.IsEditActive)
	k.NextField.SetEnabled(state != nil && state.IsEditActive)
	k.PreviousField.SetEnabled(state != nil && state.IsEditActive)
	k.NextComboValue.SetEnabled(state != nil && state.IsEditActive)
	k.PreviousComboValue.SetEnabled(state != nil && state.IsEditActive)
}
