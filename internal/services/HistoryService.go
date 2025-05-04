package services

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/fchastanet/shell-command-bookmarker/internal/processors"
	"github.com/fchastanet/shell-command-bookmarker/internal/services/models"
)

const (
	// MinCommandLength is the minimum length of a command to be ingested
	MinCommandLength = 6
)

type HistoryIngestor interface {
	// IngestHistoryWithCallback reads the bash history file and ingests it into the database using a callback
	ParseBashHistory(
		historyFilePath string, fromTimestamp time.Time,
		callback func(processors.HistoryCommand) (processors.CommandImportedStatus, error),
	) error
}

type HistoryService struct {
	ingestor          HistoryIngestor
	homeDir           string
	dbService         *DBService
	lintService       *LintService
	scriptRegexp      *regexp.Regexp
	ignoreLinesRegexp []*regexp.Regexp
}

func NewHistoryService(
	ingestor HistoryIngestor,
	dbService *DBService,
	lintService *LintService,
) *HistoryService {
	return &HistoryService{
		ingestor:          ingestor,
		dbService:         dbService,
		lintService:       lintService,
		homeDir:           "",
		scriptRegexp:      nil,
		ignoreLinesRegexp: nil,
	}
}

func (s *HistoryService) GetHistoryRows() ([]*models.Command, error) {
	// Only fetch commands with the statuses we want to display
	cmds, err := s.dbService.GetCommands(
		models.CommandStatusImported,
		models.CommandStatusSaved,
		models.CommandStatusBookmarked,
	)
	if err != nil {
		slog.Error("Error getting history rows", "error", err)
		return []*models.Command{}, err
	}

	return cmds, nil
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

func (s *HistoryService) Init() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("Error getting home directory", "error", err)
		return err
	}
	s.homeDir = homeDir
	return nil
}

func (s *HistoryService) getDefaultHistoryFilePath() (string, error) {
	historyFile := filepath.Join(s.homeDir, ".bash_history")
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

func (s *HistoryService) getHistoryFilePath() (string, error) {
	historyFile := os.Getenv("HISTFILE")
	if historyFile == "" {
		historyFile, err := s.getDefaultHistoryFilePath()
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

func (s *HistoryService) checkIfCommandShouldBeSaved(cmd processors.HistoryCommand) (processors.CommandImportedStatus, error) {
	if len(cmd.Command) < MinCommandLength {
		slog.Info("Command too short, skipping", "command", cmd)
		return processors.CommandImportedStatusSkipped, nil
	}
	if !s.getScriptRegexp().MatchString(cmd.Command) &&
		matchOneOfRegexp(cmd.Command, s.getIgnoreLinesRegexp()) {
		slog.Info("Command does not match any script or is ignored", "command", cmd)
		return processors.CommandImportedStatusFilteredOut, nil
	}

	existingCmd, err := s.dbService.GetCommandByScript(cmd.Command)
	if err != nil {
		slog.Error("Error getting command from database", "command", cmd, "error", err)
		return processors.CommandImportedStatusError, err
	}
	if existingCmd != nil {
		slog.Debug("Command already exists in database", "command", cmd)
		return processors.CommandImportedStatusAlreadyExists, nil
	}
	return processors.CommandImportedStatusNew, nil
}

func (s *HistoryService) IngestHistory() error {
	historyFilePath, err := s.getHistoryFilePath()
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

	if err := s.ingestor.ParseBashHistory(historyFilePath, maxCommandTimestamp, s.processCmd); err != nil {
		slog.Error("Error ingesting history", "file", historyFilePath, "error", err)
		return err
	}

	return nil
}

func (s *HistoryService) processCmd(historyCmd processors.HistoryCommand) (processors.CommandImportedStatus, error) {
	if importStatus, err := s.checkIfCommandShouldBeSaved(historyCmd); err != nil {
		return processors.CommandImportedStatusError, err
	} else if importStatus != processors.CommandImportedStatusNew {
		slog.Debug("Command already exists in database or is ignored", "command", historyCmd, "status", importStatus)
		return importStatus, nil
	}
	cmd := models.NewCommand(
		historyCmd.Command,
		historyCmd.Elapsed,
		historyCmd.Timestamp,
	)

	if s.lintService.IsLintingAvailable() {
		issues, err := s.lintService.LintScript(historyCmd.Command)
		if err != nil && len(issues) == 0 {
			slog.Error("Error linting command", "command", historyCmd, "error", err)
			cmd.LintStatus = models.LintStatusShellcheckFailed
			cmd.LintIssues = "[]"
		} else {
			cmd.LintStatus = s.lintService.GetLintResultingStatus(issues)
			cmd.LintIssues = s.lintService.FormatLintIssuesAsJSON(issues)
		}
		if cmd.LintStatus == models.LintStatusWarning {
			slog.Warn("Linting issues found", "command", historyCmd, "issues", issues)
		}
		if cmd.LintStatus == models.LintStatusError {
			slog.Error("Linting errors found", "command", historyCmd, "issues", issues)
		}
	}
	if err := s.dbService.SaveCommand(cmd); err != nil {
		slog.Error("Error saving command to database", "command", cmd, "error", err)
		return processors.CommandImportedStatusError, err
	}
	slog.Info("Command saved successfully", "command", cmd)
	return processors.CommandImportedStatusNew, nil
}

func (s *HistoryService) UpdateCommand(command *models.Command) (newCommand *models.Command, err error) {
	slog.Debug("Updating command", "id", command.ID, "status", command.Status)

	// If it's an IMPORTED command being updated, we need to handle duplication
	if command.Status == models.CommandStatusImported {
		originalCmd := *command

		// Step 1: Mark the original command as obsolete
		if err := s.duplicateCommandAsObsolete(command); err != nil {
			slog.Error("Failed to mark original command as obsolete", "id", command.ID, "error", err)
			return nil, err
		}

		newCommand, err = s.createNewCommandFromImportedCommand(command)
		if err != nil {
			slog.Error("Failed to create new command from imported command", "error", err)
			return nil, err
		}

		slog.Info("Created new saved command from imported command", "oldId", originalCmd.ID, "newId", newCommand.ID)
		return newCommand, nil
	}

	// For non-IMPORTED commands, just update directly
	command.ModificationDatetime = time.Now()
	// Lint the new command
	s.lintCommand(command)
	err = s.dbService.UpdateCommand(command)
	if err != nil {
		slog.Error("Error updating command in database", "id", command.ID, "error", err)
		return nil, err
	}
	return command, nil
}

func (s *HistoryService) createNewCommandFromImportedCommand(command *models.Command) (*models.Command, error) {
	// Step 2: Create a new command with SAVED status
	newCmd := models.Command{
		ID:                   0, // New ID will be generated
		Title:                command.Title,
		Description:          command.Description,
		Script:               command.Script,
		Status:               models.CommandStatusSaved,
		LintIssues:           command.LintIssues,
		LintStatus:           command.LintStatus,
		Elapsed:              command.Elapsed,
		CreationDatetime:     command.CreationDatetime,
		ModificationDatetime: time.Now(),
	}

	// Lint the new command
	s.lintCommand(&newCmd)

	// Save the new command
	if err := s.dbService.SaveCommand(&newCmd); err != nil {
		slog.Error("Error saving new command", "error", err)
		return nil, err
	}
	return &newCmd, nil
}

func (s *HistoryService) lintCommand(command *models.Command) {
	if command.Status != models.CommandStatusSaved {
		slog.Warn("Command is not in a state that can be linted", "id", command.ID, "status", command.Status)
		return
	}

	issues, err := s.lintService.LintScript(command.Script)
	if err != nil && len(issues) == 0 {
		slog.Error("Error linting command", "command", command, "error", err)
		command.LintStatus = models.LintStatusShellcheckFailed
		command.LintIssues = "[]"
	} else {
		command.LintStatus = s.lintService.GetLintResultingStatus(issues)
		command.LintIssues = s.lintService.FormatLintIssuesAsJSON(issues)
	}
	slog.Info("Command linted successfully",
		"command", command,
		"lintStatus", command.LintStatus,
		"lintIssues", command.LintIssues,
	)
}

func (s *HistoryService) duplicateCommandAsObsolete(command *models.Command) error {
	command.Status = models.CommandStatusObsolete
	command.ModificationDatetime = time.Now()
	if err := s.dbService.UpdateCommand(command); err != nil {
		slog.Error("Failed to mark original command as obsolete", "id", command.ID, "error", err)
		return err
	}
	slog.Info("Original command marked as obsolete", "id", command.ID)
	return nil
}

func matchOneOfRegexp(line string, regexps []*regexp.Regexp) bool {
	for _, r := range regexps {
		if r.MatchString(line) {
			return true
		}
	}
	return false
}
