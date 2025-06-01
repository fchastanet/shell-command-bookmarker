//go:build sqlite_fts5 || fts5

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/fchastanet/shell-command-bookmarker/app/application"
	"github.com/fchastanet/shell-command-bookmarker/internal/args"
	"github.com/fchastanet/shell-command-bookmarker/internal/services"

	// Import for side effects
	_ "embed"
)

//go:embed resources/sqlite.schema.sql
var sqliteSchema string

func main() {
	appService := services.NewAppService()
	if err := mainImpl(appService); err != nil {
		slog.Error("critical error", "error", err)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// mainImpl contains the main application logic, extracted to facilitate testing
func mainImpl(appService services.AppServiceInterface) error {
	defer appService.Cleanup()

	var cli args.Cli
	err := args.ParseArgs(&cli)
	if err != nil {
		return err
	}

	if appService.HandleShellIntegrationScriptGeneration(&cli) {
		return nil
	}

	if err := appService.Main(&cli, sqliteSchema); err != nil {
		return err
	}

	// Skip launchApp in tests
	if _, ok := appService.(*services.AppService); !ok {
		return fmt.Errorf("expected *services.AppService, got %T", appService)
	}

	if err := application.LaunchApp(appService); err != nil {
		return err
	}
	return nil
}
