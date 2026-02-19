// Package models defines data structures used throughout Tornado.
//
// This file contains structures for query results, which are returned
// by database queries and displayed in the UI.
//
// Key Learning - Data vs Behavior:
//   - This package contains data structures (structs)
//   - No methods that change state - just data containers
//   - This separates "what" (data) from "how" (behavior in other packages)
//
// TODO: Define all query-related structures:
//   - [ ] QueryResult for SELECT results
//   - [ ] ExecResult for INSERT/UPDATE/DELETE results
//   - [ ] Column metadata structure
//   - [ ] Row representation
package models

import (
	"time"
)

// QueryResult represents the result of a SELECT query.
// It contains column information and row data.
//
// TODO: Complete this structure with fields for:
//   - Column names and types
//   - Row data
//   - Execution metadata
type QueryResult struct {
	// Columns holds the names of each column returned
	// TODO: Add this field
	// Columns []string

	// ColumnTypes holds the database type for each column
	// Useful for formatting and type-aware display
	// TODO: Add this field
	// ColumnTypes []string

	// Rows holds the actual data
	// Each row is a slice of interface{} because column types vary
	// TODO: Consider alternatives:
	//   - [][]any (flexible but type-unsafe)
	//   - []map[string]any (column-name accessible)
	//   - Custom Row type with methods
	// TODO: Add this field
	// Rows [][]any

	// RowCount is the total number of rows returned
	// TODO: Add this field
	// RowCount int

	// ExecutionTime is how long the query took
	// Displayed in the UI to show query performance
	// TODO: Add this field
	// ExecutionTime time.Duration

	// Query is the original SQL that produced this result
	// Useful for display and debugging
	// TODO: Add this field
	// Query string

	// HasMore indicates if there are more rows (for pagination)
	// TODO: Consider if pagination is needed
	// HasMore bool
}

// ExecResult represents the result of an executing statement
// (INSERT, UPDATE, DELETE, CREATE, etc.)
//
// TODO: Define this structure with fields for:
//   - Rows affected
//   - Last insert ID
//   - Execution time
type ExecResult struct {
	// RowsAffected is the number of rows modified
	// TODO: Add this field
	// RowsAffected int64

	// LastInsertID is the ID of the last inserted row (if applicable)
	// Not all databases support this (PostgreSQL uses RETURNING instead)
	// TODO: Add this field
	// LastInsertID int64

	// ExecutionTime is how long the statement took
	// TODO: Add this field
	// ExecutionTime time.Duration

	// Query is the original SQL that was executed
	// TODO: Add this field
	// Query string
}

// Column represents metadata about a database column.
// Used when describing table structure.
//
// TODO: Define this structure
type Column struct {
	// Name is the column name
	Name string

	// Type is the database type (VARCHAR, INTEGER, etc.)
	Type string

	// Nullable indicates if the column can contain NULL
	Nullable bool

	// DefaultValue is the column's default value (if any)
	DefaultValue *string

	// IsPrimaryKey indicates if this column is a primary key
	IsPrimaryKey bool

	// IsForeignKey indicates if this column references another table
	IsForeignKey bool

	// ForeignKeyTable is the referenced table (if IsForeignKey)
	ForeignKeyTable string

	// ForeignKeyColumn is the referenced column (if IsForeignKey)
	ForeignKeyColumn string
}

// QueryHistoryItem represents a single entry in query history.
// Useful for the query editor to show previous queries.
//
// TODO: Implement query history tracking
type QueryHistoryItem struct {
	Query        string
	ExecutedAt   time.Time
	Duration     time.Duration
	RowCount     int
	WasError     bool
	ErrorMessage string
}
