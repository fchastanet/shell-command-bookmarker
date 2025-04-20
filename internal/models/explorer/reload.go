package explorer

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	"github.com/leg100/pug/internal/tui"
)

// reload pug modules, resolving any differences between the modules on the
// user's disk, and those loaded in pug. Set firsttime to toggle whether this is
// the first time modules are being loaded.
func reload(db *services.DBService) tea.Cmd {
	return func() tea.Msg {
		cmds, err := db.GetAllCommands()
		if err != nil {
			return tui.ReportError(fmt.Errorf("reloading modules: %w", err))()
		}
		return tui.InfoMsg(
			fmt.Sprintf("loaded %d commands", len(cmds)),
		)
	}
}
