// Package db provides database abstraction for Tornado.
//
// This package defines a common interface for database operations, allowing
// Tornado to work with multiple database backends (SQLite, PostgreSQL, etc.)
// without the UI code needing to know which database is being used.
//
// This is a key Go concept: PROGRAMMING TO INTERFACES
// By defining behaviors in an interface, we can swap implementations easily.
//
// Key Learning - Interfaces in Go:
//   - Interfaces are implicit (no "implements" keyword)
//   - An interface is a set of method signatures
//   - Any type that has those methods satisfies the interface
//   - This enables dependency injection and testing with mocks
//
// TODO: Define the Database interface with all necessary operations
// TODO: Create mock implementation for testing UI without real DB
//
// References:
//   - https://go.dev/tour/methods/9 (Interfaces)
//   - https://go.dev/doc/effective_go#interfaces
package db

import (
	"github.com/yourusername/tornado/internal/models"
)

// Database defines the common interface for all database backends.
// Any database driver (SQLite, Postgres, etc.) must implement these methods.
//
// Key Learning - Interface Design:
//   - Keep interfaces small and focused (the "interface segregation" principle)
//   - Methods should return errors for the caller to handle
//   - Consider what the UI actually needs, not what the database can do
//
// TODO: Review and finalize the interface methods:
//   - [ ] Add transaction support? (Begin, Commit, Rollback)
//   - [ ] Add connection pooling controls?
//   - [ ] Add database-specific feature detection?
type Database interface {
	// Connect establishes a connection to the database using the provided config.
	// Returns an error if connection fails.
	//
	// TODO: Define how connection config is passed
	// Options: Accept models.ConnectionConfig, or specific fields, or DSN string
	Connect(config models.ConnectionConfig) error

	// Disconnect closes the database connection gracefully.
	// Should be called when the application exits.
	Disconnect() error

	// IsConnected returns true if there's an active database connection.
	// Useful for UI state checks.
	IsConnected() bool

	// Query executes a SQL query and returns the results.
	// Used for SELECT statements and any custom SQL the user writes.
	//
	// TODO: Define what QueryResult contains:
	// - Column names
	// - Row data (as [][]any or map[string]any)
	// - Row count
	// - Execution time
	Query(sql string) (*models.QueryResult, error)

	// Exec executes a SQL statement that doesn't return rows.
	// Used for INSERT, UPDATE, DELETE, CREATE, etc.
	//
	// TODO: Define ExecResult:
	// - Rows affected
	// - Last insert ID (if applicable)
	Exec(sql string) (*models.ExecResult, error)

	// ListTables returns a list of all user tables in the database.
	// Used for the browser sidebar.
	ListTables() ([]string, error)

	// ListSchemas returns a list of schemas (Postgres) or databases (MySQL).
	// For SQLite, this might return an empty list or ["main"].
	ListSchemas() ([]string, error)

	// DescribeTable returns the schema for a specific table.
	// Includes column names, types, nullability, keys, etc.
	DescribeTable(name string) (*models.TableSchema, error)

	// GetType returns the database type (sqlite, postgres, etc.)
	// Useful for showing database-specific features or syntax hints.
	GetType() string
}

// Open creates a new database connection based on the config type.
// This is a factory function that returns the appropriate implementation.
//
// Key Learning - Factory Pattern:
//   - The caller doesn't need to know which implementation to create
//   - Centralized logic for choosing the right database driver
//
// TODO: Implement this factory function
// Example implementation:
//
//	func Open(config models.ConnectionConfig) (Database, error) {
//	    switch config.Type {
//	    case "sqlite":
//	        return NewSQLiteDB(config), nil
//	    case "postgres":
//	        return NewPostgresDB(config), nil
//	    default:
//	        return nil, fmt.Errorf("unsupported database type: %s", config.Type)
//	    }
//	}
func Open(config models.ConnectionConfig) (Database, error) {
	// TODO: Implement factory logic
	return nil, nil
}
