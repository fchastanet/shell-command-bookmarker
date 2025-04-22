package models

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/structure"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
)

type ErrOpeningEditor struct {
	Path string
	Err  error
}

func (e ErrOpeningEditor) Error() string {
	return fmt.Sprintf("opening %s in editor: %v", e.Path, e.Err)
}

var ErrCannotOpenEditor = errors.New("cannot open editor: environment variable EDITOR not set")

// NavigateTo sends an instruction to navigate to a page with the given model
// kind, and optionally parent resource.
func NavigateTo(kind resource.Kind, opts ...structure.NavigateOption) tea.Cmd {
	return tui.CmdHandler(structure.NewNavigationMsg(kind, opts...))
}

func ReportInfo(msg string, args ...any) tea.Cmd {
	return tui.CmdHandler(tui.InfoMsg(fmt.Sprintf(msg, args...)))
}

func OpenEditor(path string) tea.Cmd {
	// TODO: check for side effects of exec blocking the tui - do
	// messages get queued up?
	editor, ok := os.LookupEnv("EDITOR")
	if !ok {
		return tui.ReportError(ErrCannotOpenEditor)
	}
	cmd := exec.Command(editor, path)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return tui.ReportError(&ErrOpeningEditor{Path: path, Err: err})()
		}
		return nil
	})
}
