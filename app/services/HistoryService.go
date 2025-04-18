package services

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/fchastanet/shell-command-bookmarker/app/processors"
)

const (
	// MinCommandLength is the minimum length of a command to be ingested
	MinCommandLength = 6
)

type HistoryIngestor interface {
	// IngestHistoryWithCallback reads the bash history file and ingests it into the database using a callback
	ParseBashHistory(historyFilePath string, fromTimestamp time.Time, callback func(processors.HistoryCommand) error) error
}

type HistoryService struct {
	ingestor          HistoryIngestor
	dbService         *DBService
	scriptRegexp      *regexp.Regexp
	ignoreLinesRegexp []*regexp.Regexp
}

func NewHistoryService(
	ingestor HistoryIngestor,
	dbService *DBService,
) *HistoryService {
	return &HistoryService{
		ingestor:          ingestor,
		dbService:         dbService,
		scriptRegexp:      nil,
		ignoreLinesRegexp: nil,
	}
}

func (s *HistoryService) getScriptRegexp() *regexp.Regexp {
	if s.scriptRegexp != nil {
		return s.scriptRegexp
	}
	s.scriptRegexp = regexp.MustCompile("[|&;><()\\[\\]{}$*?!+=,`]")
	return s.scriptRegexp
}

func (s *HistoryService) getIgnoreLinesRegexp() []*regexp.Regexp {
	if s.ignoreLinesRegexp != nil {
		return s.ignoreLinesRegexp
	}

	s.ignoreLinesRegexp = []*regexp.Regexp{
		regexp.MustCompile("^#"),
		regexp.MustCompile("( --version| --help)"),
		regexp.MustCompile("^(shutdown|export|kill|ln|man|mc|ls|ll|ps|source|which|command -v|cd|pwd|echo|cat|rm|mv|cp|touch|mkdir|rmdir|chmod|chown|top|killall|grep|find|locate|updatedb|z) "),
		regexp.MustCompile("^(code|vi|vim|nano|exit|logout|clear|history|alias|unalias|export|unset|set|env|source|bash|sh|zsh) "),
		regexp.MustCompile(`^(\./[^ ]+|exit|ls|alias|cd)$`),
		regexp.MustCompile(`^[A-Za-z0-9_]+=[^ ]+$`),
		regexp.MustCompile(`^\s*$`),
	}
	return s.ignoreLinesRegexp
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

func (s *HistoryService) checkIfCommandShouldBeSaved(cmd processors.HistoryCommand) (bool, error) {
	if len(cmd.Command) < MinCommandLength {
		slog.Info("Command too short, skipping", "command", cmd)
		return false, nil
	}
	if !s.getScriptRegexp().MatchString(cmd.Command) &&
		matchOneOfRegexp(cmd.Command, s.getIgnoreLinesRegexp()) {
		slog.Info("Command does not match any script or is ignored", "command", cmd)
		return false, nil
	}

	existingCmd, err := s.dbService.GetCommandByScript(cmd.Command)
	if err != nil {
		slog.Error("Error getting command from database", "command", cmd, "error", err)
		return false, err
	}
	if existingCmd != nil {
		slog.Debug("Command already exists in database", "command", cmd)
		return false, nil
	}
	return true, nil
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

	maxCommandTimestamp, err := s.dbService.GetMaxCommandTimestamp()
	if err != nil {
		slog.Debug("Error getting max command timestamp, fallback to 0", "error", err)
		maxCommandTimestamp = time.Time{}
	}
	slog.Debug("Max command timestamp", "timestamp", maxCommandTimestamp)

	if err := s.ingestor.ParseBashHistory(historyFilePath, maxCommandTimestamp, func(cmd processors.HistoryCommand) error {
		if ok, err := s.checkIfCommandShouldBeSaved(cmd); err != nil {
			return err
		} else if !ok {
			slog.Debug("Command already exists in database", "command", cmd)
			return nil
		}
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

func matchOneOfRegexp(line string, regexps []*regexp.Regexp) bool {
	for _, regexp := range regexps {
		if regexp.MatchString(line) {
			return true
		}
	}
	return false
}
