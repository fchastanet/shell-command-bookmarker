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

	historyFilePath, err := getHistoryFilePath()
	if err != nil {
		slog.Error("Error getting history file path", "error", err)
		return err
	}
	if historyFilePath != "" {
		if err := parseBashHistory(historyFilePath); err != nil {
			slog.Error("Error parsing bash history", "file", historyFilePath, "error", err)
		}
	}

	if _, err := tea.NewProgram(
		m,
		tea.WithReportFocus(),
	).Run(); err != nil {
		slog.Error("Error running program", "error", err)
		return err
	}
	return nil
}

func getHistoryFilePath() (string, error) {
	historyFile := os.Getenv("HISTFILE")
	if historyFile != "" {
		_, err := os.Stat(historyFile)
		if errors.Is(err, os.ErrNotExist) {
			slog.Error("History file does not exist", "file", historyFile, "error", err)
			return "", err
		}
		if errors.Is(err, os.ErrPermission) {
			slog.Error("Permission denied to access history file", "file", historyFile, "error", err)
			return "", err
		}
		return historyFile, nil
	}
	return "", nil
}

func parseBashHistory(historyFile string) error {
	// Create error channel to capture errors from goroutine
	errChan := make(chan error, 1)

	// Process bash history in a background goroutine
	go func() {
		err := processors.ParseBashHistory(historyFile, func(cmd processors.HistoryCommand) error {
			slog.Info("Parsed command", "command", cmd.Command, "timestamp", cmd.Timestamp, "elapsed", cmd.Elapsed)
			return nil
		})

		// Send error (or nil) to channel when done
		errChan <- err
	}()

	// Before returning from mainImpl, check if there were any errors in processing
	select {
	case err := <-errChan:
		if err != nil {
			slog.Error("Error parsing history file", "file", historyFile, "error", err)
			return err
		}
	default:
		// Processing still in progress, but we don't want to block here
		slog.Info("History file processing still in progress")
	}
	return nil
}
