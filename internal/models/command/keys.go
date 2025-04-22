package command

import (
	"github.com/charmbracelet/bubbles/key"
)

type ResourcesKeyMap struct {
	Move   *key.Binding
	Reload *key.Binding
	Enter  *key.Binding
}

func GetResourcesKeyMap() *ResourcesKeyMap {
	move := key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "move"),
	)
	reload := key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload"),
	)
	enter := key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "view resource"),
	)
	return &ResourcesKeyMap{
		Move:   &move,
		Reload: &reload,
		Enter:  &enter,
	}
}
