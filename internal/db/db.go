// This package defines a common interface for database operations
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

	// ListViews returns a list of all views in the database.
	ListViews() ([]string, error)

	// ListIndexes returns a list of indexes for a specific table.
	ListIndexes(tableName string) ([]string, error)

	// ListTriggers returns a list of all triggers in the database.
	ListTriggers() ([]string, error)

	// ListSequences returns a list of all sequences in the database.
	// (Primarily for PostgreSQL, returns empty for SQLite)
	ListSequences() ([]string, error)

	// GetType returns the database type (sqlite, postgres, etc.)
	GetType() string
}

// Open creates a new database connection based on the config type.
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
