// This file contains structures for query results, which are returned
// by database queries and displayed in the UI.
package models

import (
	"time"
)

// QueryResult represents the result of a SELECT query.
// It contains column information and row data.
type QueryResult struct {
	// Columns holds the names of each column returned
	Columns []string

	// ColumnTypes holds the database type for each column
	// Useful for formatting and type-aware display
	ColumnTypes []string

	// Rows holds the actual data
	// Each row is a slice of interface{} because column types vary
	// TODO: Consider alternatives:
	//   - [][]any (flexible but type-unsafe)
	//   - []map[string]any (column-name accessible)
	//   - Custom Row type with methods
	Rows [][]any

	// RowCount is the total number of rows returned
	RowCount int

	// ExecutionTime is how long the query took
	// Displayed in the UI to show query performance
	ExecutionTime time.Duration

	// Query is the original SQL that produced this result
	// Useful for display and debugging
	Query string

	// HasMore indicates if there are more rows (for pagination)
	HasMore bool
}

// ExecResult represents the result of an executing statement
type ExecResult struct {
	// RowsAffected is the number of rows modified
	RowsAffected int64

	// LastInsertID is the ID of the last inserted row (if applicable)
	// TODO: make this db dependent postgres uses RETURNING
	LastInsertID int64

	// ExecutionTime is how long the statement took
	ExecutionTime time.Duration

	// Query is the original SQL that was executed
	Query string
}

// Column represents metadata about a database column.
// Used when describing table structure.
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

// TableSchema represents the structure of a database table.
// Returned by DescribeTable method of Database interface.
type TableSchema struct {
	// Name is the table name
	Name string

	// Columns is the list of columns in the table
	Columns []Column

	// PrimaryKey is the name of the primary key column(s)
	PrimaryKey []string

	// Indexes contains information about table indexes
	Indexes []IndexInfo

	// RowCount is an estimate of the number of rows
	RowCount int64
}

// IndexInfo represents information about a database index.
type IndexInfo struct {
	Name    string
	Columns []string
	Unique  bool
}

// QueryHistoryItem represents a single entry in query history.
// Useful for the query editor to show previous queries.
type QueryHistoryItem struct {
	Query        string
	ExecutedAt   time.Time
	Duration     time.Duration
	RowCount     int
	WasError     bool
	ErrorMessage string
}
