package keys

import "github.com/charmbracelet/bubbles/key"

type common struct {
	Delete key.Binding
	Reload key.Binding
	Edit   key.Binding
	Back   key.Binding
}

// Keys shared by several models.
var Common = common{
	Delete: key.NewBinding(
		key.WithKeys("delete"),
		key.WithHelp("delete", "delete"),
	),
	Reload: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload"),
	),
	Edit: key.NewBinding(
		key.WithKeys("E"),
		key.WithHelp("E", "edit"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
}
