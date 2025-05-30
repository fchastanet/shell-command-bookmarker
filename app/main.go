//go:build sqlite_fts5 || fts5

package main

import (
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/args"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/styles"
	"github.com/fchastanet/shell-command-bookmarker/internal/models/top"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"
	"github.com/mattn/go-isatty"

	// Import for side effects
	_ "embed"
)

//go:embed resources/sqlite.schema.sql
var sqliteSchema string

func main() {
	if err := mainImpl(); err != nil {
		slog.Error("critical error", "error", err)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func mainImpl() error {
	if !isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		slog.Error("This program requires a terminal to run. Please run it in a terminal emulator.")
		return &services.InvalidTerminalError{}
	}

	var cli args.Cli
	err := args.ParseArgs(&cli)
	if err != nil {
		return err
	}

	// Handle shell integration script generation if requested
	if cli.GenerateBash || cli.GenerateZsh {
		shellIntegrationService := services.NewShellIntegrationService()

		if cli.GenerateBash {
			fmt.Print(shellIntegrationService.GenerateBashIntegration())
			return nil
		}

		if cli.GenerateZsh {
			fmt.Print(shellIntegrationService.GenerateZshIntegration())
			return nil
		}
	}

	appService := services.NewAppService(services.AppServiceConfig{
		SqliteSchema: sqliteSchema,
		MaxTasks:     1,
		DBPath:       string(cli.DBPath),
		Debug:        cli.Debug,
		OutputFile:   cli.OutputFile,
	})
	defer appService.Cleanup()
	err = appService.Init()
	if err != nil {
		return err
	}

	go func() {
		if err := appService.HistoryService.IngestHistory(); err != nil {
			slog.Error("Error ingesting history", "error", err)
			// Depending on requirements, you might want to signal this error back
			// to the main thread or handle it differently. For now, just logging.
		}
	}()

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

// The shell integration code has been moved to the ShellIntegrationService
