package services

import "github.com/fchastanet/shell-command-bookmarker/internal/db"

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
