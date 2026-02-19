// This package defines a common interface for database operations, allowing
// Tornado to work with multiple database backends (SQLite, PostgreSQL, etc.)
// without the UI code needing to know which database is being used.
//
// References:
//   - https://go.dev/tour/methods/9 (Interfaces)
//   - https://go.dev/doc/effective_go#interfaces
package db

import (
	"fmt"

	"github.com/jupiterozeye/tornado/internal/models"
)

// Database defines the common interface for all database backends.
// Any database driver (SQLite, Postgres, etc.) must implement these methods.
type Database interface {
	// Connect establishes a connection to the database using the provided config.
	// Returns an error if connection fails.
	Connect(config models.ConnectionConfig) error

	// Disconnect closes the database connection gracefully.
	Disconnect() error

	// IsConnected returns true if there's an active database connection.
	IsConnected() bool

	// Query executes a SQL query and returns the results.
	Query(sql string) (*models.QueryResult, error)

	// Exec executes a SQL statement that doesn't return rows.
	Exec(sql string) (*models.ExecResult, error)

	// ListTables returns a list of all user tables in the database.
	ListTables() ([]string, error)

	// ListSchemas returns a list of schemas (Postgres) or databases (MySQL).
	ListSchemas() ([]string, error)

	// DescribeTable returns the schema for a specific table.
	// Includes column names, types, nullability, keys, etc.
	DescribeTable(name string) (*models.TableSchema, error)

	// GetType returns the database type (sqlite, postgres, etc.)
	GetType() string
}

// Open creates a new database connection based on the config type.
// This is a factory function that returns the appropriate implementation.
func Open(config models.ConnectionConfig) (Database, error) {
	switch config.Type {
	case "sqlite":
		db := NewSQLiteDB()
		return db, db.Connect(config)
	case "postgres":
		db := NewPostgresDB()
		return db, db.Connect(config)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}
