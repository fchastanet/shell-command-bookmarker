package application

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/top"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
)

func LaunchApp(appService services.AppServiceInterface) error {
	myStyles := styles.NewStyles()
	myStyles.Init()

	m := top.NewModel(
		appService,
		myStyles,
	)

	if _, err := tea.NewProgram(
		&m,
		tea.WithReportFocus(),
	).Run(); err != nil {
		slog.Error("Error running program", "error", err)
		return err
	}
	return nil
}
