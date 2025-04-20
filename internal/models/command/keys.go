package command

import (
	"github.com/charmbracelet/bubbles/key"
)

type resourcesKeyMap struct {
	Move   key.Binding
	Reload key.Binding
	Enter  key.Binding
}

var resourcesKeys = resourcesKeyMap{
	Move: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "move"),
	),
	Reload: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "view resource"),
	),
}
