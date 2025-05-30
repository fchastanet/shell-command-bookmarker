package keys

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

func GetFormKeyMap() *huh.KeyMap {
	defaultHuhKeyMap := huh.NewDefaultKeyMap()
	defaultHuhKeyMap.Confirm.Accept.SetKeys("y", "Y", "ctrl+c")
	defaultHuhKeyMap.Confirm.Accept.SetHelp("y/Ctrl+c", "confirm")
	defaultHuhKeyMap.Confirm.Reject.SetKeys("n", "N", "esc")
	defaultHuhKeyMap.Confirm.Reject.SetHelp("n/␛", "No")
	return defaultHuhKeyMap
}

func GetFormBindings() []*key.Binding {
	var bindings []*key.Binding
	defaultHuhKeyMap := GetFormKeyMap()
	bindings = append(
		bindings,
		&defaultHuhKeyMap.Quit,
		&defaultHuhKeyMap.Confirm.Accept,
		&defaultHuhKeyMap.Confirm.Reject,
		&defaultHuhKeyMap.Confirm.Toggle,
		&defaultHuhKeyMap.Confirm.Submit,
	)
	return bindings
}
