package services

import (
	"errors"
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/internal/processors"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
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

	return nil
}

// CreateTasks repeatedly invokes fn with each id in ids, creating a task for
// each invocation. If there is more than one id then a task group is created
// and the user sent to the task group's page; otherwise if only id is provided,
// the user is sent to the task's page.
func (app *AppService) CreateTasks(fn task.SpecFunc, ids ...resource.ID) tea.Cmd {
	return func() tea.Msg {
		switch len(ids) {
		case 0:
			return nil
		case 1:
			spec, err := fn(ids[0])
			if err != nil {
				return ErrorMsg(fmt.Errorf("creating task: %w", err))
			}
			task, err := h.Tasks.Create(spec)
			if err != nil {
				return ErrorMsg(fmt.Errorf("creating task: %w", err))
			}
			if task.Short {
				// Don't navigate the user to the task page for short tasks.
				return nil
			}
			return NewNavigationMsg(TaskKind, WithParent(task.ID))
		default:
			specs := make([]task.Spec, 0, len(ids))
			for _, id := range ids {
				spec, err := fn(id)
				if err != nil {
					h.Logger.Error("creating task spec", "error", err, "id", id)
					continue
				}
				specs = append(specs, spec)
			}
			return h.createTaskGroup(specs...)
		}
	}
}
