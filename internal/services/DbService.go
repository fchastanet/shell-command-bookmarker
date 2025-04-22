package services

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/fchastanet/shell-command-bookmarker/pkg/db"
)

type CommandStatus string

const (
	CommandStatusImported   CommandStatus = "IMPORTED"
	CommandStatusBookmarked CommandStatus = "BOOKMARKED"
)

type DBService struct {
	dbAdapter  db.Adapter
	dbPath     string
	schemaPath string
}

type Command struct {
	ID                   int
	Title                string
	Description          string
	Script               string
	Status               CommandStatus
	LintIssues           string
	LintStatus           LintStatus
	Elapsed              int
	CreationDatetime     time.Time
	ModificationDatetime time.Time
}

type CommandStatusEnum struct {
	Imported   CommandStatus
	Bookmarked CommandStatus
}

func NewDBService(
	dbPath string,
	schemaPath string,
) *DBService {
	return &DBService{
		dbAdapter:  db.NewSQLiteAdapter(dbPath, schemaPath),
		dbPath:     dbPath,
		schemaPath: schemaPath,
	}
}

func (s *DBService) Open() error {
	return s.dbAdapter.Open()
}

func (s *DBService) Close() error {
	return s.dbAdapter.Close()
}

func (s *DBService) GetDBAdapter() db.Adapter {
	return s.dbAdapter
}

func (s *DBService) SaveCommand(command *Command) error {
	// Use Exec instead of Query for INSERT statements
	_, err := s.dbAdapter.GetDB().Exec(
		`INSERT INTO command (
			title, description, script, status,
			lint_issues, lint_status, elapsed,
			creation_datetime, modification_datetime
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		command.Title, command.Description, command.Script, string(command.Status),
		command.LintIssues, string(command.LintStatus), command.Elapsed,
		command.CreationDatetime.Format(time.DateTime), command.ModificationDatetime.Format(time.DateTime),
	)
	if err != nil {
		return err
	}
	return nil
}

// GetCommandByID retrieves a command by its database ID
func (s *DBService) GetCommandByID(id int) (*Command, error) {
	slog.Debug("Retrieving command by id from database", "id", id)
	// Use QueryRow for single row retrieval
	row := s.dbAdapter.GetDB().QueryRow(
		`SELECT id, title, description, script, status,
			lint_issues, lint_status, elapsed,
			creation_datetime, modification_datetime
			FROM command WHERE id = ? LIMIT 1`,
		id,
	)
	if row == nil {
		slog.Debug("No command found in database", "id", id)
		return nil, nil
	}
	return s.getCommandFromRow(row)
}

func (s *DBService) GetCommandByScript(script string) (*Command, error) {
	slog.Debug("Retrieving command by script from database", "script", script)
	// Use QueryRow for single row retrieval
	row := s.dbAdapter.GetDB().QueryRow(
		`SELECT id, title, description, script, status,
			lint_issues, lint_status, elapsed,
			creation_datetime, modification_datetime
			FROM command WHERE script = ? LIMIT 1`,
		script,
	)
	if row == nil {
		slog.Debug("No command found in database", "script", script)
		return nil, nil
	}

	return s.getCommandFromRow(row)
}

func (s *DBService) getCommandFromRow(row *sql.Row) (*Command, error) {
	var command Command
	var creationDateStr string
	var modificationDateStr string

	err := row.Scan(
		&command.ID,
		&command.Title,
		&command.Description,
		&command.Script,
		&command.Status,
		&command.LintIssues,
		&command.LintStatus,
		&command.Elapsed,
		&creationDateStr,
		&modificationDateStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		// Handle other scan errors
		slog.Error("Error scanning command from database", "error", err)
		return nil, err
	}

	command.CreationDatetime, err = time.Parse(time.DateTime, creationDateStr)
	if err != nil {
		return nil, err
	}

	command.ModificationDatetime, err = time.Parse(time.DateTime, modificationDateStr)
	if err != nil {
		return nil, err
	}
	return &command, nil
}

func (s *DBService) GetMaxCommandTimestamp() (time.Time, error) {
	var maxTimestampStr string
	var maxTimestamp time.Time

	// Query for the maximum creation_datetime
	row := s.dbAdapter.GetDB().QueryRow("SELECT IFNULL(MAX(creation_datetime), '1970-01-01 00:00:00') FROM command")
	err := row.Scan(&maxTimestampStr)
	if err != nil {
		// Handle case where table might be empty or other scan errors
		// If the table is empty, Scan might return sql.ErrNoRows, depending on the driver and MAX behavior with NULLs.
		// We might want to return time.Zero or a specific error in that case.
		// For simplicity, returning zero time and the error for now.
		return time.Time{}, err
	}

	// Parse the timestamp string into time.Time
	// Assuming the format stored is time.DateTime ("2006-01-02 15:04:05")
	maxTimestamp, err = time.Parse(time.DateTime, maxTimestampStr)
	if err != nil {
		return time.Time{}, err
	}

	return maxTimestamp, nil
}

func (s *DBService) GetAllCommands() ([]*Command, error) {
	var commands []*Command
	var creationDateStr string
	var modificationDateStr string

	rows, err := s.dbAdapter.GetDB().Query(
		`SELECT id, title, description, script, status,
			lint_issues, lint_status, elapsed,
			creation_datetime, modification_datetime
			FROM command`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		command := Command{
			ID:                   0,
			Title:                "",
			Description:          "",
			Script:               "",
			Status:               "",
			LintIssues:           "",
			LintStatus:           "",
			Elapsed:              0,
			CreationDatetime:     time.Time{},
			ModificationDatetime: time.Time{},
		}
		err := rows.Scan(
			&command.ID,
			&command.Title,
			&command.Description,
			&command.Script,
			&command.Status,
			&command.LintIssues,
			&command.LintStatus,
			&command.Elapsed,
			&creationDateStr,
			&modificationDateStr,
		)
		if err != nil {
			return nil, err
		}

		command.CreationDatetime, err = time.Parse(time.DateTime, creationDateStr)
		if err != nil {
			return nil, err
		}

		command.ModificationDatetime, err = time.Parse(time.DateTime, modificationDateStr)
		if err != nil {
			return nil, err
		}
		slog.Debug("Command retrieved from database", "command", command)
		commands = append(commands, &command)
	}
	return commands, nil
}
