// Package models - Database connection configuration and state.
//
// This file defines structures for managing database connections,
// including configuration parameters and connection state tracking.
//
// TODO: Define all connection-related structures:
//   - [ ] ConnectionConfig for storing connection parameters
//   - [ ] ConnectionState for tracking active connections
//   - [ ] ConnectionHistory for recent connections list
package models

import (
	"time"
)

// ConnectionConfig holds all parameters needed to connect to a database.
// Different database types use different subsets of these fields.
//
// Key Learning - Configuration Structures:
//   - Group related configuration into a struct
//   - Makes passing configuration easier (one arg vs many)
//   - Can be serialized to/from files (JSON, YAML)
//
// TODO: Complete this structure with all necessary fields
type ConnectionConfig struct {
	// Type specifies the database type: "sqlite" or "postgres"
	Type string

	// Name is a friendly name for this connection
	// Displayed in the connection list and title bar
	Name string

	// ===== SQLite-specific fields =====

	// Path is the file path for SQLite databases
	Path string

	// ===== PostgreSQL-specific fields =====

	// Host is the PostgreSQL server hostname
	Host string

	// Port is the PostgreSQL server port
	Port int

	// User is the database username
	User string

	// Password is the database password
	// TODO: Consider security - should this be stored?
	// Options: store encrypted, use keychain, prompt each time
	Password string

	// Database is the database name to connect to
	Database string

	// SSLMode controls SSL/TLS connection settings
	// Values: disable, require, verify-ca, verify-full
	SSLMode string

	// Schema is the default schema to use (PostgreSQL)
	Schema string

	// ===== Connection options =====

	// Timeout is the connection timeout duration
	Timeout time.Duration

	// MaxOpenConns is the maximum number of open connections
	MaxOpenConns int

	// MaxIdleConns is the maximum number of idle connections
	MaxIdleConns int
}

// IsValid performs basic validation on the connection config.
//
// TODO: Implement validation logic
//   - Check that Type is valid (sqlite or postgres)
//   - For sqlite: Path must be set
//   - For postgres: Host, User, Database must be set
//   - Port should be valid (1-65535)
func (c *ConnectionConfig) IsValid() bool {
	// TODO: Implement validation
	return c.Type != ""
}

// ConnectionString returns a display-safe connection string.
// Passwords are masked for security.
//
// TODO: Implement this for displaying in the UI
func (c *ConnectionConfig) ConnectionString() string {
	// TODO: Implement based on database type
	// For SQLite: just show the path
	// For Postgres: host:port/database (mask password)
	return ""
}

// ConnectionState tracks the current state of a database connection.
//
// TODO: Define state tracking fields
type ConnectionState struct {
	// Connected indicates if currently connected
	Connected bool

	// ConnectedAt is when the connection was established
	ConnectedAt time.Time

	// ServerVersion is the database server version
	ServerVersion string

	// CurrentDatabase is the currently connected database name
	CurrentDatabase string

	// CurrentSchema is the current schema (PostgreSQL)
	CurrentSchema string

	// Error holds any connection error
	Error error
}

// ConnectionHistoryItem represents a saved/recent connection.
// Used for the connection screen's "recent connections" list.
//
// TODO: Implement connection history persistence
type ConnectionHistoryItem struct {
	// Config is the connection configuration
	Config ConnectionConfig

	// LastConnected is when this connection was last used
	LastConnected time.Time

	// UseCount is how many times this connection has been used
	UseCount int
}

// ConnectionHistory manages a list of recent connections.
//
// TODO: Implement history management
//   - Add new connections
//   - Load from file on startup
//   - Save to file on exit
//   - Limit to max items
type ConnectionHistory struct {
	Items []ConnectionHistoryItem
	Max   int
}

// Add adds a connection to the history.
//
// TODO: Implement this method
func (h *ConnectionHistory) Add(config ConnectionConfig) {
	// TODO: Add to history, update timestamps, limit size
}

// Save persists the connection history to a file.
//
// TODO: Implement file persistence
// Consider using:
//   - JSON (simple, human-readable)
//   - YAML (more flexible)
//   - OS keychain for passwords (secure)
func (h *ConnectionHistory) Save() error {
	// TODO: Implement
	return nil
}

// Load reads the connection history from a file.
//
// TODO: Implement file loading
func (h *ConnectionHistory) Load() error {
	// TODO: Implement
	return nil
}
