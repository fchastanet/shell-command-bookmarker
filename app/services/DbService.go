package services

import (
	"time"

	"github.com/fchastanet/shell-command-bookmarker/app/processors"
	"github.com/fchastanet/shell-command-bookmarker/internal/db"
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

func (s *DBService) SaveCommand(command processors.HistoryCommand) error {
	date := command.Timestamp.Format(time.DateTime)
	// Use Exec instead of Query for INSERT statements
	_, err := s.dbAdapter.GetDB().Exec(
		"INSERT INTO command (title, description, script, elapsed, creation_datetime, modification_datetime, status) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?)",
		"", "",
		command.Command, command.Elapsed, date, date,
		string(CommandStatusImported),
	)
	if err != nil {
		return err
	}
	return nil
}
