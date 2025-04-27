package db

import (
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"

	// Import for side effects
	// Build with: go build -tags "sqlite_fts5"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DirectoryPerm = 0o755
)

// SQLiteAdapter represents a connection to a SQLite database
type SQLiteAdapter struct {
	db     *sql.DB
	path   string
	schema string
}

type Driver interface {
	Ping() error
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type Adapter interface {
	Open() error
	Close() error
	GetDB() Driver
	BeginTx() (*sql.Tx, error)
}

// NewSQLiteAdapter creates a new SQLite adapter
func NewSQLiteAdapter(dbPath, schema string) Adapter {
	return &SQLiteAdapter{
		db:     nil,
		path:   dbPath,
		schema: schema,
	}
}

// Open opens the database connection and initializes the schema if needed
func (a *SQLiteAdapter) Open() error {
	// Create the directory if it doesn't exist
	dbDir := filepath.Dir(a.path)
	if err := os.MkdirAll(dbDir, DirectoryPerm); err != nil {
		return &DatabaseDirectoryCreationError{
			Directory:  dbDir,
			InnerError: err,
		}
	}

	// Check if the database file exists
	isNew := !fileExists(a.path)

	// Open the database connection with foreign keys and FTS5 enabled
	db, err := sql.Open("sqlite3", a.path+"?_foreign_keys=on&_sqlite_fts5=1")
	if err != nil {
		return &DatabaseNotFoundError{
			DBFilePath: a.path,
		}
	}
	a.db = db

	// Test the connection
	if err := db.Ping(); err != nil {
		return &DatabaseConnectionError{
			DBFilePath: a.path,
			InnerError: err,
		}
	}

	// Initialize the schema if the database is new
	if isNew {
		if err := a.initSchema(); err != nil {
			err := a.Close() // Close the DB if initialization fails
			if err != nil {
				slog.Error("Error closing database after schema initialization failure", "error", err)
			}
			return &SchemaInitializationError{
				DBFilePath: a.path,
				InnerError: err,
			}
		}
	}

	return nil
}

// Close closes the database connection
func (a *SQLiteAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// GetDB returns the database connection
func (a *SQLiteAdapter) GetDB() Driver {
	return a.db
}

// BeginTx starts a new transaction
func (a *SQLiteAdapter) BeginTx() (*sql.Tx, error) {
	return a.db.Begin()
}

// initSchema initializes the database schema
func (a *SQLiteAdapter) initSchema() error {
	// Execute the schema SQL
	_, err := a.db.Exec(a.schema)
	if err != nil {
		return &QueryExecutionError{
			DBFilePath: a.path,
			Query:      "Loading Schema",
			InnerError: err,
		}
	}

	return nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
