package services

import (
	"errors"
	"log/slog"

	"github.com/fchastanet/shell-command-bookmarker/internal/processors"
)

type AppService struct {
	Config         *AppServiceConfig
	DBService      *DBService
	LintService    *LintService
	HistoryService *HistoryService
	LoggerService  *LoggerService
	Cleanup        func()
}

type AppServiceConfig struct {
	MaxTasks     int
	Debug        bool
	DBPath       string
	SqliteSchema string
}

func NewAppService(cfg AppServiceConfig) *AppService {
	dbService := NewDBService(cfg.DBPath, cfg.SqliteSchema)

	lintService := NewLintService()
	loggerService := NewLoggerService(cfg.Debug)

	historyService := NewHistoryService(
		processors.NewHistoryIngestor(),
		dbService,
		lintService,
	)

	// cleanup function to be invoked when app is terminated.
	cleanup := func() {
		// Perform cleanup tasks here
		// e.g., close database connections, release resources, etc.
		dbService.Close()
		loggerService.Close()
	}

	return &AppService{
		Config:         &cfg,
		Cleanup:        cleanup,
		DBService:      dbService,
		LintService:    lintService,
		HistoryService: historyService,
		LoggerService:  loggerService,
	}
}

func (app *AppService) Init() error {
	if err := app.LoggerService.Init(); err != nil {
		slog.Error("Error initializing logger service", "error", err)
		return err
	}

	if err := app.LintService.Init(); err != nil {
		if errors.Is(err, ErrShellCheckNotFound) {
			slog.Warn("shellcheck command not found in PATH. Linting will be disabled.", "error", err)
		} else {
			slog.Error("Error creating LintService", "error", err)
			return err
		}
	}

	if err := app.DBService.Open(); err != nil {
		slog.Error("Error opening database", "error", err)
		return err
	}

	if err := app.HistoryService.Init(); err != nil {
		slog.Error("Error initializing history service", "error", err)
	}

	return nil
}
