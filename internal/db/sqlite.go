// Package db - SQLite implementation of the Database interface.
//
// This file implements the Database interface for SQLite databases.
// SQLite is a good first implementation because:
//   - No server setup required (just a file)
//   - Great for development and testing
//   - Built into Go's database/sql package (with driver)
//
// Key Learning - Implementing an Interface:
//   - This file will define a SQLiteDB struct
//   - That struct will have methods matching the Database interface
//   - Go will automatically recognize it implements Database
//   - No explicit "implements" declaration needed
//
// Key Learning - database/sql Package:
//   - Go's standard library for SQL databases
//   - Uses driver pattern (import the driver for side effects)
//   - Provides connection pooling automatically
//   - Uses prepared statements for security
//
// Dependencies:
//   - github.com/mattn/go-sqlite3 (CGO-based, most common)
//   - modernc.org/sqlite (Pure Go, no CGO, good for cross-compilation)
//
// TODO: Implement the SQLiteDB struct and all Database interface methods:
//   - [ ] Define SQLiteDB struct with *sql.DB field
//   - [ ] Implement NewSQLiteDB constructor
//   - [ ] Implement Connect method (opens database file)
//   - [ ] Implement Disconnect method (closes connection)
//   - [ ] Implement Query method
//   - [ ] Implement Exec method
//   - [ ] Implement ListTables method (query sqlite_master)
//   - [ ] Implement DescribeTable method
//   - [ ] Handle SQLite-specific errors and edge cases
//
// References:
//   - https://pkg.go.dev/database/sql
//   - https://github.com/mattn/go-sqlite3
//   - https://www.sqlite.org/lang.html (SQLite SQL syntax)
package db

import (
	"database/sql"

	"github.com/jupiterozeye/tornado/internal/models"
)

// SQLiteDB implements the Database interface for SQLite databases.
// It wraps the standard sql.DB connection pool.
//
// TODO: Add any SQLite-specific fields:
//   - File path (for display purposes)
//   - Connection state
//   - Prepared statement cache (optional optimization)
type SQLiteDB struct {
	// db is the connection pool to the SQLite database
	// TODO: Add this field
	// db *sql.DB

	// path is the file path to the SQLite database
	// TODO: Add this field for display in UI
	// path string

	// connected tracks whether we have an active connection
	// TODO: Consider if this is needed or if db != nil is sufficient
	// connected bool
}

// NewSQLiteDB creates a new SQLiteDB instance.
// Note: This doesn't connect yet - call Connect() to open the database.
//
// TODO: Implement constructor
func NewSQLiteDB() *SQLiteDB {
	return &SQLiteDB{}
}

// Connect opens the SQLite database file.
//
// Key Learning - SQL Driver Registration:
//   - Importing github.com/mattn/go-sqlite3 registers the "sqlite3" driver
//   - sql.Open("sqlite3", path) creates a connection pool
//   - The returned *sql.DB is safe for concurrent use
//
// TODO: Implement Connect method
// Steps:
//  1. Parse connection config for file path
//  2. Call sql.Open("sqlite3", path)
//  3. Ping to verify connection works
//  4. Store connection in SQLiteDB struct
//
// SQLite-specific considerations:
//   - If file doesn't exist, SQLite creates it (may want to check first)
//   - Use "_journal_mode=WAL" in DSN for better concurrency
//   - Use "_foreign_keys=on" to enable FK constraints
func (s *SQLiteDB) Connect(config models.ConnectionConfig) error {
	// TODO: Implement
	//
	// Example:
	// db, err := sql.Open("sqlite3", config.Path)
	// if err != nil {
	//     return fmt.Errorf("failed to open database: %w", err)
	// }
	//
	// // Verify connection works
	// if err := db.Ping(); err != nil {
	//     return fmt.Errorf("failed to connect to database: %w", err)
	// }
	//
	// s.db = db
	// s.path = config.Path
	// s.connected = true
	// return nil

	return nil
}

// Disconnect closes the SQLite database connection.
//
// TODO: Implement Disconnect method
func (s *SQLiteDB) Disconnect() error {
	// TODO: Implement
	// if s.db != nil {
	//     return s.db.Close()
	// }
	return nil
}

// IsConnected returns whether there's an active connection.
//
// TODO: Implement IsConnected method
func (s *SQLiteDB) IsConnected() bool {
	// TODO: Implement
	return false
}

// Query executes a SELECT query and returns results.
//
// Key Learning - Working with sql.Rows:
//   - Use rows.Next() to iterate
//   - Use rows.Columns() to get column names
//   - Use rows.Scan() to read values into variables
//   - Always call rows.Close() (use defer)
//
// TODO: Implement Query method
// Steps:
//  1. Execute query with db.QueryContext or db.Query
//  2. Get column names with rows.Columns()
//  3. Iterate and scan rows into a slice
//  4. Build and return QueryResult
//
// Challenges to handle:
//   - Different column types (int, float, string, blob, null)
//   - Large result sets (consider pagination)
//   - Query timeouts
func (s *SQLiteDB) Query(sql string) (*models.QueryResult, error) {
	// TODO: Implement
	return nil, nil
}

// Exec executes a statement that doesn't return rows.
//
// TODO: Implement Exec method
// Use db.Exec for INSERT/UPDATE/DELETE
// Return rows affected and last insert ID
func (s *SQLiteDB) Exec(sql string) (*models.ExecResult, error) {
	// TODO: Implement
	return nil, nil
}

// ListTables returns all user tables in the SQLite database.
//
// SQLite stores table metadata in sqlite_master:
//
//	SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'
//
// TODO: Implement ListTables method
func (s *SQLiteDB) ListTables() ([]string, error) {
	// TODO: Implement
	return nil, nil
}

// ListSchemas returns schema names.
// For SQLite, this is typically just ["main"].
//
// TODO: Implement ListSchemas method
func (s *SQLiteDB) ListSchemas() ([]string, error) {
	// TODO: Implement
	// SQLite uses "main" as the default schema name
	// Attached databases would have their own names
	return []string{"main"}, nil
}

// DescribeTable returns column information for a table.
//
// Use PRAGMA table_info(table_name) to get column details:
//   - cid: column id
//   - name: column name
//   - type: data type
//   - notnull: 1 if NOT NULL
//   - dflt_value: default value
//   - pk: 1 if primary key
//
// TODO: Implement DescribeTable method
func (s *SQLiteDB) DescribeTable(name string) (*models.TableSchema, error) {
	// TODO: Implement
	return nil, nil
}

// GetType returns "sqlite" to identify the database type.
func (s *SQLiteDB) GetType() string {
	return "sqlite"
}

// Ensure SQLiteDB implements Database interface at compile time.
// This is a compile-time check - if SQLiteDB doesn't implement all
// Database methods, this line will cause a compilation error.
//
// Key Learning - Interface Satisfaction Check:
//   - The underscore assigns to nothing (we don't need the variable)
//   - The type assertion checks the interface is satisfied
//   - This catches missing methods early, not at runtime
var _ Database = (*SQLiteDB)(nil)
