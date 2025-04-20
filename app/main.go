//go:build sqlite_fts5 || fts5
// +build sqlite_fts5 fts5

package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/app/models"
	"github.com/fchastanet/shell-command-bookmarker/app/processors"
	"github.com/fchastanet/shell-command-bookmarker/app/services"

	// Import for side effects
	_ "embed"
)

const WriteFileMode = 0o644

//go:embed resources/sqlite.schema.sql
var sqliteSchema string

func initLogger(level slog.Level, logFileHandler io.Writer) {
	slog.SetLogLoggerLevel(level)
	opts := &slog.HandlerOptions{
		AddSource:   level == slog.LevelDebug,
		Level:       level,
		ReplaceAttr: nil,
	}
	handler := slog.NewTextHandler(logFileHandler, opts)

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func main() {
	if err := mainImpl(); err != nil {
		slog.Error("critical error", "error", err)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func mainImpl() error {
	f, err := tea.LogToFile("logs/error.log", "debug")
	if err != nil {
		return err
	}
	defer f.Close()
	level := slog.LevelError
	debug := os.Getenv("DEBUG")
	var dump *os.File
	if debug != "" {
		level = slog.LevelDebug
		var err error
		dump, err = os.OpenFile("logs/debug.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, WriteFileMode)
		if err != nil {
			slog.Error("Error opening debug log file", "error", err)
			os.Exit(1)
		}
		defer dump.Close()
	}
	initLogger(level, f)

	dbPath := "db/shell-command-bookmarker.db"
	if os.Getenv("SHELL_CMD_BOOK_DB") != "" {
		dbPath = os.Getenv("SHELL_CMD_BOOK_DB")
	}
	dbService := services.NewDBService(dbPath, sqliteSchema)
	defer dbService.Close()

	lintService, err := services.NewLintService()
	if err != nil {
		if errors.Is(err, services.ErrShellCheckNotFound) {
			slog.Warn("shellcheck command not found in PATH. Linting will be disabled.", "error", err)
		} else {
			slog.Error("Error creating LintService", "error", err)

			return err
		}
	}

	historyService := services.NewHistoryService(
		processors.NewHistoryIngestor(),
		dbService,
		lintService,
	)
	if err := dbService.Open(); err != nil {
		slog.Error("Error opening database", "error", err)
		return err
	}
	go func() {
		if err := historyService.IngestHistory(); err != nil {
			slog.Error("Error ingesting history", "error", err)
			// Depending on requirements, you might want to signal this error back
			// to the main thread or handle it differently. For now, just logging.
		}
	}()

	m := models.NewAppModel(
		historyService,
		dump,
	)

	if _, err := tea.NewProgram(
		m,
		tea.WithReportFocus(),
	).Run(); err != nil {
		slog.Error("Error running program", "error", err)
		return err
	}
	return nil
}
