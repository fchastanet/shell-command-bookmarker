package db

import "fmt"

type DatabaseDirectoryCreationError struct {
	Directory  string
	InnerError error
}

func (e *DatabaseDirectoryCreationError) Error() string {
	return fmt.Sprintf(
		"failed to create database directory: %s (inner error %v)",
		e.Directory,
		e.InnerError,
	)
}

type DatabaseNotFoundError struct {
	DBFilePath string
}

func (e *DatabaseNotFoundError) Error() string {
	return fmt.Sprintf("unable to open database file(%s): file not found",
		e.DBFilePath,
	)
}

type DatabaseConnectionError struct {
	DBFilePath string
	InnerError error
}

func (e *DatabaseConnectionError) Error() string {
	return fmt.Sprintf("database connection failure for file: %s (inner error: %v)",
		e.DBFilePath,
		e.InnerError,
	)
}

type SchemaInitializationError struct {
	DBFilePath string
	InnerError error
}

func (e *SchemaInitializationError) Error() string {
	return fmt.Sprintf("schema initialization failure for database file: %s (inner error: %v)",
		e.DBFilePath,
		e.InnerError,
	)
}

type SchemaReadError struct {
	DBFilePath string
	InnerError error
}

func (e *SchemaReadError) Error() string {
	return fmt.Sprintf("schema read failure for database file: %s (inner error: %v)",
		e.DBFilePath,
		e.InnerError,
	)
}

type QueryExecutionError struct {
	DBFilePath string
	Query      string
	InnerError error
}

func (e *QueryExecutionError) Error() string {
	return fmt.Sprintf("query execution failure for database file: %s with query: %s (inner error: %v)",
		e.DBFilePath,
		e.Query,
		e.InnerError,
	)
}
