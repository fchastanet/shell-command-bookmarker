package tui

import (
	"fmt"

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

type InfoMsg string

func ReportInfo(msg string, args ...any) tea.Cmd {
	return CmdHandler(InfoMsg(fmt.Sprintf(msg, args...)))
}

// FilterFocusReqMsg is a request to focus the filter widget.
type FilterFocusReqMsg struct{}

// FilterBlurMsg is a request to un-focus the filter widget. It is not
// acknowledged.
type FilterBlurMsg struct{}

// FilterCloseMsg is a request to close the filter widget. It is not
// acknowledged.
type FilterCloseMsg struct{}

// FilterKeyMsg is a key entered by the user into the filter widget
type FilterKeyMsg tea.KeyMsg
