package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func CmdHandler(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

type ErrorMsg error

func ReportError(err error) tea.Cmd {
	return CmdHandler(ErrorMsg(err))
}
