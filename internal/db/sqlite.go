// This file implements the Database interface for SQLite databases.
// References:
//   - https://pkg.go.dev/database/sql
//   - https://www.sqlite.org/lang.html (SQLite SQL syntax)
package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jupiterozeye/tornado/internal/models"
	_ "modernc.org/sqlite"
)

// SQLiteDB implements the Database interface for SQLite databases.
// It wraps the standard sql.DB connection pool.
type SQLiteDB struct {
	// db is the connection pool to the SQLite database
	db *sql.DB
	// path is the file path to the SQLite database
	path string
	// connected tracks whether we have an active connection
	connected bool
}

// NewSQLiteDB creates a new SQLiteDB instance.
func NewSQLiteDB() *SQLiteDB {
	return &SQLiteDB{}
}

// Connect opens the SQLite database file.
func (s *SQLiteDB) Connect(config models.ConnectionConfig) error {
	db, err := sql.Open("sqlite", config.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for faster close/disconnect
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	s.db = db
	s.path = config.Path
	s.connected = true
	return nil
}

// Disconnect closes the SQLite database connection with timeout.
func (s *SQLiteDB) Disconnect() error {
	if s.db == nil {
		return nil
	}

	// Use a channel to close with timeout
	done := make(chan error, 1)
	go func() {
		done <- s.db.Close()
	}()

	// Wait for close with timeout
	select {
	case err := <-done:
		s.db = nil
		s.connected = false
		return err
	case <-time.After(2 * time.Second):
		// Timeout - force close by setting to nil
		s.db = nil
		s.connected = false
		return fmt.Errorf("disconnect timeout: connection may still be active")
	}
}

// IsConnected returns whether there's an active connection.
func (s *SQLiteDB) IsConnected() bool {
	switch s.connected {
	case true:
		return true
	default:
		return false
	}
}

// Query executes a SELECT query and returns results.
func (s *SQLiteDB) Query(sql string) (*models.QueryResult, error) {
	start := time.Now()

	// Execute query
	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get column metadata
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	columnTypes, err := rows.ColumnTypes()

	// Extract type names for the result
	typeNames := make([]string, len(columnTypes))
	for i, ct := range columnTypes {
		typeNames[i] = ct.DatabaseTypeName()
	}
	// Prepare result container
	var results [][]any

	// Iterate rows
	for rows.Next() {
		// Create a slice of pointer to scan into
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan row data
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		results = append(results, values)

	}
	return &models.QueryResult{
		Columns:       columns,
		ColumnTypes:   typeNames,
		Rows:          results,
		RowCount:      len(results),
		ExecutionTime: time.Since(start),
		Query:         sql,
	}, nil
}

// Exec executes a statement that doesn't return rows.
func (s *SQLiteDB) Exec(sql string) (*models.ExecResult, error) {
	if !s.connected || s.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	// Begin timer
	start := time.Now()
	// Execute SQL
	result, err := s.db.Exec(sql)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := result.RowsAffected()
	lastInsertId, _ := result.LastInsertId()

	// Return result
	return &models.ExecResult{
		RowsAffected:  rowsAffected,
		LastInsertID:  lastInsertId,
		ExecutionTime: time.Since(start),
		Query:         sql,
	}, nil
}

// ListTables returns all user tables in the SQLite database.
func (s *SQLiteDB) ListTables() ([]string, error) {
	if !s.connected || s.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}
	// Run query
	rows, err := s.db.Query(`
			SELECT name FROM sqlite_master
			WHERE type='table' AND name NOT LIKE 'sqlite_%'
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	// Populate list with tables
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}

	// Return the tables
	return tables, rows.Err()
}

// ListSchemas returns schema names.
// For SQLite, this is typically just ["main"].
func (s *SQLiteDB) ListSchemas() ([]string, error) {
	if !s.connected || s.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}
	return []string{"main"}, nil
}

// DescribeTable returns column information for a table.
func (s *SQLiteDB) DescribeTable(name string) (*models.TableSchema, error) {
	if !s.connected || s.db == nil {
		return nil, fmt.Errorf("not connected to databse")
	}

	// Query PRAMGA table_info
	rows, err := s.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", name))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []models.Column
	var primaryKeys []string

	for rows.Next() {
		var cid int
		var colName, colType string
		var notNull, pk int
		var dfltValue sql.NullString

		err := rows.Scan(&cid, &colName, &colType, &notNull, &dfltValue, &pk)
		if err != nil {
			return nil, err
		}

		col := models.Column{
			Name:     colName,
			Type:     colType,
			Nullable: notNull == 0, // nullable when 0
		}

		if dfltValue.Valid {
			col.DefaultValue = &dfltValue.String
		}

		if pk == 1 {
			col.IsPrimaryKey = true
			primaryKeys = append(primaryKeys, colName)
		}

		columns = append(columns, col)
	}

	return &models.TableSchema{
		Name:       name,
		Columns:    columns,
		PrimaryKey: primaryKeys,
	}, rows.Err()
}

// GetType returns "sqlite" to identify the database type.
func (s *SQLiteDB) GetType() string {
	return "sqlite"
}

// ListViews returns all views in the SQLite database.
func (s *SQLiteDB) ListViews() ([]string, error) {
	if !s.connected || s.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	rows, err := s.db.Query(`
		SELECT name FROM sqlite_master
		WHERE type='view' AND name NOT LIKE 'sqlite_%'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		views = append(views, name)
	}

	return views, rows.Err()
}

// ListIndexes returns all indexes for a specific table.
func (s *SQLiteDB) ListIndexes(tableName string) ([]string, error) {
	if !s.connected || s.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	rows, err := s.db.Query(fmt.Sprintf("PRAGMA index_list(%s)", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []string
	for rows.Next() {
		var seq int
		var name string
		var unique int
		var origin string
		var partial int

		err := rows.Scan(&seq, &name, &unique, &origin, &partial)
		if err != nil {
			return nil, err
		}
		indexes = append(indexes, name)
	}

	return indexes, rows.Err()
}

// ListTriggers returns all triggers in the SQLite database.
func (s *SQLiteDB) ListTriggers() ([]string, error) {
	if !s.connected || s.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	rows, err := s.db.Query(`
		SELECT name FROM sqlite_master
		WHERE type='trigger' AND name NOT LIKE 'sqlite_%'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var triggers []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		triggers = append(triggers, name)
	}

	return triggers, rows.Err()
}

// ListSequences returns all sequences in the SQLite database.
// For SQLite, this returns an empty slice as SQLite doesn't have a traditional sequence concept.
func (s *SQLiteDB) ListSequences() ([]string, error) {
	// SQLite uses AUTOINCREMENT, not sequences like PostgreSQL
	return []string{}, nil
}

// Ensure SQLiteDB implements Database interface at compile time.
var _ Database = (*SQLiteDB)(nil)
