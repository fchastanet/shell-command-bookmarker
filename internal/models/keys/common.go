package keys

import "github.com/charmbracelet/bubbles/key"

type CommonKeyMap struct {
	Delete *key.Binding
	Reload *key.Binding
	Edit   *key.Binding
	Back   *key.Binding
}

// Keys shared by several models.

func GetCommonKeyMap() *CommonKeyMap {
	deleteKey := key.NewBinding(
		key.WithKeys("delete"),
		key.WithHelp("delete", "delete"),
	)
	reload := key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload"),
	)
	edit := key.NewBinding(
		key.WithKeys("E"),
		key.WithHelp("E", "edit"),
	)
	back := key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	)
	return &CommonKeyMap{
		Delete: &deleteKey,
		Reload: &reload,
		Edit:   &edit,
		Back:   &back,
	}
}
