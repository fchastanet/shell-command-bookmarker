package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func CheckKey(msg tea.KeyMsg, binding *key.Binding) bool {
	if binding == nil {
		return false
	}
	if binding.Enabled() && key.Matches(msg, *binding) {
		return true
	}
	return false
}
