package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fchastanet/shell-command-bookmarker/app/models"
	"github.com/fchastanet/shell-command-bookmarker/internal/db"
	"github.com/fchastanet/shell-command-bookmarker/internal/framework/focus"

	// Import for side effects
	_ "embed"
)

//go:embed resources/sqlite.schema.sql
var sqliteSchema string

func initLogger(level slog.Level, logFileHandler io.Writer) {
	slogLevel := slog.SetLogLoggerLevel(level)
	opts := &slog.HandlerOptions{
		AddSource:   slogLevel == slog.LevelDebug,
		Level:       slogLevel,
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
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		return err
	}
	defer f.Close()
	level := slog.LevelError
	if os.Getenv("DEBUG") != "" {
		level = slog.LevelDebug
	}
	initLogger(level, f)

	dbPath := "db/shell-command-bookmarker.db"
	if os.Getenv("SHELL_CMD_BOOK_DB") != "" {
		dbPath = os.Getenv("SHELL_CMD_BOOK_DB")
	}
	dbAdapter := db.NewSQLiteAdapter(dbPath, sqliteSchema)
	if err := dbAdapter.Open(); err != nil {
		return err
	}
	defer dbAdapter.Close()

	focusManager := focus.NewFocusManager()
	m := models.NewAppModel(
		focusManager,
	)
	focusManager.SetRootComponents([]focus.Focusable{&m})

	if _, err := tea.NewProgram(
		m,
		tea.WithReportFocus(),
	).Run(); err != nil {
		slog.Error("Error running program", "error", err)
		return err
	}
	return nil
}
