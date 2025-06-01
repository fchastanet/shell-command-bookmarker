package services

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/fchastanet/shell-command-bookmarker/internal/args"
	"github.com/fchastanet/shell-command-bookmarker/internal/processors"
	"github.com/mattn/go-isatty"
)

type AppService struct {
	Config                  *AppServiceConfig
	DBService               *DBService
	LintService             *LintService
	HistoryService          *HistoryService
	LoggerService           *LoggerService
	ShellIntegrationService *ShellIntegrationService
	ShellDetectionService   ShellDetectionServiceInterface
	cleanupFunc             func()
}

type AppServiceConfig struct {
	DBPath       string
	SqliteSchema string
	OutputFile   string // Flag to indicate if we're in shell selection mode
	MaxTasks     int
	Debug        bool
}

func NewAppService() *AppService {
	return &AppService{
		Config:                  nil,
		cleanupFunc:             func() {},
		DBService:               nil,
		LintService:             nil,
		HistoryService:          nil,
		LoggerService:           nil,
		ShellIntegrationService: nil,
		ShellDetectionService:   nil,
	}
}

func (app *AppService) Self() *AppService {
	return app
}

func (app *AppService) Init(cfg AppServiceConfig) error {
	app.Config = &cfg

	app.LoggerService = NewLoggerService(cfg.Debug)
	if err := app.LoggerService.Init(); err != nil {
		slog.Error("Error initializing logger service", "error", err)
		return err
	}

	app.DBService = NewDBService(cfg.DBPath, cfg.SqliteSchema)

	// cleanup function to be invoked when app is terminated.
	cleanup := func() {
		// Perform cleanup tasks here
		// e.g., close database connections, release resources, etc.
		err := app.DBService.Close()
		if err != nil {
			slog.Error("Error closing database", "error", err)
		}
		app.LoggerService.Close() // #nosec G104
	}
	app.cleanupFunc = cleanup
	if err := app.DBService.Open(); err != nil {
		slog.Error("Error opening database", "error", err)
		return err
	}

	app.LintService = NewLintService()
	if err := app.LintService.Init(); err != nil {
		if errors.Is(err, ErrShellCheckNotFound) {
			slog.Warn("shellcheck command not found in PATH. Linting will be disabled.", "error", err)
		} else {
			slog.Error("Error creating LintService", "error", err)
			return err
		}
	}

	app.HistoryService = NewHistoryService(
		processors.NewHistoryIngestor(),
		app.DBService,
		app.LintService,
	)
	if err := app.HistoryService.Init(); err != nil {
		slog.Error("Error initializing history service", "error", err)
	}
	slog.Info("AppService initialized successfully", "dbPath", cfg.DBPath, "debug", cfg.Debug)

	app.ShellIntegrationService = NewShellIntegrationService()
	app.ShellDetectionService = NewShellDetectionService()

	return nil
}

func (app *AppService) Main(cli *args.Cli, sqliteSchema string) error {
	if err := app.IsTerminalCompatible(); err != nil {
		slog.Error("Terminal compatibility check failed", "error", err)
		return err
	}

	err := app.Init(AppServiceConfig{
		SqliteSchema: sqliteSchema,
		MaxTasks:     1,
		DBPath:       string(cli.DBPath),
		Debug:        cli.Debug,
		OutputFile:   cli.OutputFile,
	})
	if err != nil {
		slog.Error("Error initializing AppService", "error", err)
		return err
	}

	go func() {
		if err := app.GetHistoryService().IngestHistory(); err != nil {
			slog.Error("Error ingesting history", "error", err)
			// Depending on requirements, you might want to signal this error back
			// to the main thread or handle it differently. For now, just logging.
		}
	}()

	return nil
}

func (*AppService) IsTerminalCompatible() error {
	if !isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		slog.Error("This program requires a terminal to run. Please run it in a terminal emulator.")
		return &InvalidTerminalError{Err: nil}
	}
	return nil
}

func (app *AppService) IsShellSelectionMode() bool {
	return app.Config.OutputFile != ""
}

func (app *AppService) HandleShellIntegrationScriptGeneration(cli *args.Cli) bool {
	if !cli.GenerateBash && !cli.GenerateZsh && !cli.AutoDetect {
		return false
	}
	if cli.GenerateBash {
		fmt.Print(app.ShellIntegrationService.GenerateBashIntegration())
	}

	if cli.GenerateZsh {
		fmt.Print(app.ShellIntegrationService.GenerateZshIntegration())
	}

	if cli.AutoDetect {
		// Auto-detect shell type and generate appropriate integration script
		shellType := app.ShellDetectionService.DetectShell()

		slog.Debug("Auto-detected shell type", "shellType", shellType)

		if shellType == ShellTypeZsh {
			fmt.Print(app.ShellIntegrationService.GenerateZshIntegration())
		} else {
			// Default to bash integration or use detected bash shell
			fmt.Print(app.ShellIntegrationService.GenerateBashIntegration())
		}
	}
	return true
}

// GetHistoryService returns the HistoryService
func (app *AppService) GetHistoryService() *HistoryService {
	return app.HistoryService
}

// Cleanup executes the stored cleanup function
func (app *AppService) Cleanup() {
	if app.cleanupFunc != nil {
		app.cleanupFunc()
	}
}
