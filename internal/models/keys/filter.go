package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type FilterKeyMap struct {
	Blur  *key.Binding
	Close *key.Binding
}

// FilterKeyMap is a key map of keys available in filter mode.
func GetFilterKeyMap() *FilterKeyMap {
	blur := key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("⏎", "exit filter"),
	)
	closeKey := key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("␛/Ctrl+c", "clear filter"),
	)
	return &FilterKeyMap{
		Blur:  &blur,
		Close: &closeKey,
	}
}
