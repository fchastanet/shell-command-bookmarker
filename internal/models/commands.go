package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/structure"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
)

// NavigateTo sends an instruction to navigate to a page with the given model
// kind, and optionally parent resource.
func NavigateTo(kind resource.Kind, opts ...structure.NavigateOption) tea.Cmd {
	return tui.CmdHandler(structure.NewNavigationMsg(kind, opts...))
}
