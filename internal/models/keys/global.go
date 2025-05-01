package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type GlobalKeyMap struct {
	Search *key.Binding
	Filter *key.Binding
	Quit   *key.Binding
	Help   *key.Binding
	Debug  *key.Binding
}

func GetGlobalKeyMap() *GlobalKeyMap {
	search := key.NewBinding(
		key.WithKeys("ctrl+f", "f3", "ctrl+r"),
		key.WithHelp("ctrl+f/ctrl+r/F3", "search"),
	)
	filter := key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp(`/`, "filter"),
	)
	quit := key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc/ctrl+c", "exit"),
	)
	help := key.NewBinding(
		key.WithKeys("h", "?"),
		key.WithHelp("h/?", "close help"),
	)
	debug := key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "show debug info"),
	)

	return &GlobalKeyMap{
		Search: &search,
		Filter: &filter,
		Quit:   &quit,
		Help:   &help,
		Debug:  &debug,
	}
}
