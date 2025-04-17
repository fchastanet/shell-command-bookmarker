package services

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fchastanet/shell-command-bookmarker/app/processors"
)

type HistoryIngestor interface {
	// IngestHistoryWithCallback reads the bash history file and ingests it into the database using a callback
	ParseBashHistory(historyFilePath string, callback func(processors.HistoryCommand) error) error
}

type HistoryService struct {
	ingestor  HistoryIngestor
	dbService *DBService
}

func NewHistoryService(
	ingestor HistoryIngestor,
	dbService *DBService,
) *HistoryService {
	return &HistoryService{
		ingestor:  ingestor,
		dbService: dbService,
	}
}

func getHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("Error getting home directory", "error", err)
		return "", err
	}
	return homeDir, nil
}

func getDefaultHistoryFilePath() (string, error) {
	homeDir, err := getHomeDir()
	if err != nil {
		slog.Error("Error getting home directory", "error", err)
		return "", err
	}
	historyFile := filepath.Join(homeDir, ".bash_history")
	if _, err := os.Stat(historyFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Error("History file does not exist", "file", historyFile, "error", err)
			return "", err
		}
		if errors.Is(err, os.ErrPermission) {
			slog.Error("Permission denied to access history file", "file", historyFile, "error", err)
			return "", err
		}
	}
	return historyFile, nil
}

func getHistoryFilePath() (string, error) {
	historyFile := os.Getenv("HISTFILE")
	if historyFile == "" {
		historyFile, err := getDefaultHistoryFilePath()
		if err != nil {
			slog.Error("Error getting default history file path", "error", err)
			return "", err
		}
		slog.Info("Using default history file path", "file", historyFile)
		return historyFile, nil
	}
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

func (s *HistoryService) IngestHistory() error {
	historyFilePath, err := getHistoryFilePath()
	if err != nil {
		slog.Error("Error getting history file path", "error", err)
		return err
	}

	if historyFilePath == "" {
		slog.Warn("No history file path provided")
		return nil
	}

	if err := s.ingestor.ParseBashHistory(historyFilePath, func(cmd processors.HistoryCommand) error {
		if err := s.dbService.SaveCommand(cmd); err != nil {
			slog.Error("Error saving command to database", "command", cmd, "error", err)
			return err
		}
		slog.Info("Command saved successfully", "command", cmd)
		return nil
	}); err != nil {
		slog.Error("Error ingesting history", "file", historyFilePath, "error", err)
		return err
	}

	return nil
}
