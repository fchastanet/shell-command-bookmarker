package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type GlobalKeyMap struct {
	Search           *key.Binding
	Filter           *key.Binding
	ShrinkPaneHeight *key.Binding
	GrowPaneHeight   *key.Binding
	ShrinkPaneWidth  *key.Binding
	GrowPaneWidth    *key.Binding
	ClosePane        *key.Binding
	Quit             *key.Binding
	Help             *key.Binding
	Debug            *key.Binding
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
	shrinkPaneHeight := key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "reduce height"),
	)
	growPaneHeight := key.NewBinding(
		key.WithKeys("+"),
		key.WithHelp("+", "increase height"),
	)
	shrinkPaneWidth := key.NewBinding(
		key.WithKeys("<"),
		key.WithHelp("<", "reduce width"),
	)
	growPaneWidth := key.NewBinding(
		key.WithKeys(">"),
		key.WithHelp(">", "increase width"),
	)
	closePane := key.NewBinding(
		key.WithKeys("X"),
		key.WithHelp("X", "close pane"),
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
		Search:           &search,
		Filter:           &filter,
		ShrinkPaneHeight: &shrinkPaneHeight,
		GrowPaneHeight:   &growPaneHeight,
		ShrinkPaneWidth:  &shrinkPaneWidth,
		GrowPaneWidth:    &growPaneWidth,
		ClosePane:        &closePane,
		Quit:             &quit,
		Help:             &help,
		Debug:            &debug,
	}
}
