package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type FilterKeyMap struct {
	Blur  key.Binding
	Close key.Binding
}

// FilterKeyMap is a key map of keys available in filter mode.
func GetFilterKeyMap() *FilterKeyMap {
	return &FilterKeyMap{
		Blur: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "exit filter"),
		),
		Close: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear filter"),
		),
	}
}
