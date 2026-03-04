// Package config provides application configuration management with YAML persistence.
//
// This package follows XDG Base Directory Specification for config file location:
//   - Linux/macOS: ~/.config/tornado/config.yaml
//   - Windows: %APPDATA%\tornado\config.yaml
//
// Configuration includes:
//   - Theme preference
//   - Connection history (successful connections only, no passwords)
//   - Recent queries (last 20)
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jupiterozeye/tornado/internal/models"
	"gopkg.in/yaml.v3"
)

const (
	appName        = "tornado"
	maxConnections = 10
	maxQueries     = 20
	configFileName = "config.yaml"
)

// Config holds all application configuration.
type Config struct {
	mu sync.RWMutex

	// Theme is the currently selected theme name
	Theme string `yaml:"theme"`

	// Connections is the list of recent successful connections
	Connections []ConnectionEntry `yaml:"connections"`

	// Queries is the list of recent queries
	Queries []string `yaml:"queries"`

	// Internal - not persisted
	configPath string
}

// ConnectionEntry represents a saved connection (passwords not stored).
type ConnectionEntry struct {
	Name          string    `yaml:"name"`
	Type          string    `yaml:"type"`
	Path          string    `yaml:"path,omitempty"`
	Host          string    `yaml:"host,omitempty"`
	Port          int       `yaml:"port,omitempty"`
	User          string    `yaml:"user,omitempty"`
	Database      string    `yaml:"database,omitempty"`
	SSLMode       string    `yaml:"ssl_mode,omitempty"`
	LastConnected time.Time `yaml:"last_connected"`
	UseCount      int       `yaml:"use_count"`
}

// Global config instance
var (
	globalConfig *Config
	globalMu     sync.RWMutex
)

// Get returns the global config instance.
// Call Load() before using this.
func Get() *Config {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalConfig
}

// Load loads configuration from the standard config path.
// If the config file doesn't exist, it creates a default config.
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	cfg := &Config{
		Theme:       "nord", // Default theme
		Connections: make([]ConnectionEntry, 0),
		Queries:     make([]string, 0),
		configPath:  configPath,
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		if err := cfg.saveUnlocked(); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
	} else {
		// Load existing config
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	globalMu.Lock()
	globalConfig = cfg
	globalMu.Unlock()

	return cfg, nil
}

// Save persists the configuration to disk.
// Safe to call from outside the struct (acquires read lock).
func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.saveUnlocked()
}

// saveUnlocked persists without acquiring any mutex.
// Must only be called when the caller already holds the lock (or during init).
func (c *Config) saveUnlocked() error {
	// Ensure config directory exists
	configDir := filepath.Dir(c.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(c.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetTheme returns the current theme name.
func (c *Config) GetTheme() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Theme
}

// SetTheme sets and saves the theme.
func (c *Config) SetTheme(theme string) error {
	c.mu.Lock()
	c.Theme = theme
	err := c.saveUnlocked()
	c.mu.Unlock()
	return err
}

// AddConnection adds a successful connection to history.
// If the connection already exists, it updates the timestamp and use count.
func (c *Config) AddConnection(cfg models.ConnectionConfig) error {
	// Create entry from config (no password stored)
	entry := ConnectionEntry{
		Name:          cfg.Name,
		Type:          cfg.Type,
		Path:          cfg.Path,
		Host:          cfg.Host,
		Port:          cfg.Port,
		User:          cfg.User,
		Database:      cfg.Database,
		SSLMode:       cfg.SSLMode,
		LastConnected: time.Now(),
		UseCount:      1,
	}

	// If no name provided, generate one
	if entry.Name == "" {
		if entry.Type == "sqlite" {
			entry.Name = filepath.Base(entry.Path)
		} else {
			entry.Name = fmt.Sprintf("%s@%s", entry.User, entry.Host)
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if connection already exists
	for i, existing := range c.Connections {
		if connectionsEqual(existing, entry) {
			// Update existing entry
			c.Connections[i].LastConnected = entry.LastConnected
			c.Connections[i].UseCount++
			// Move to front
			c.moveConnectionToFront(i)
			return c.saveUnlocked()
		}
	}

	// Add new entry at the beginning
	c.Connections = append([]ConnectionEntry{entry}, c.Connections...)

	// Trim to max connections
	if len(c.Connections) > maxConnections {
		c.Connections = c.Connections[:maxConnections]
	}

	return c.saveUnlocked()
}

// GetConnections returns the list of recent connections.
func (c *Config) GetConnections() []ConnectionEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy
	result := make([]ConnectionEntry, len(c.Connections))
	copy(result, c.Connections)
	return result
}

// ToConnectionConfig converts a ConnectionEntry back to ConnectionConfig.
func (e ConnectionEntry) ToConnectionConfig() models.ConnectionConfig {
	return models.ConnectionConfig{
		Type:     e.Type,
		Name:     e.Name,
		Path:     e.Path,
		Host:     e.Host,
		Port:     e.Port,
		User:     e.User,
		Database: e.Database,
		SSLMode:  e.SSLMode,
	}
}

// AddQuery adds a query to recent queries history.
func (c *Config) AddQuery(query string) error {
	if query == "" {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if query already exists
	for i, existing := range c.Queries {
		if existing == query {
			// Move to front
			c.moveQueryToFront(i)
			return c.saveUnlocked()
		}
	}

	// Add new query at the beginning
	c.Queries = append([]string{query}, c.Queries...)

	// Trim to max queries
	if len(c.Queries) > maxQueries {
		c.Queries = c.Queries[:maxQueries]
	}

	return c.saveUnlocked()
}

// GetQueries returns the list of recent queries.
func (c *Config) GetQueries() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy
	result := make([]string, len(c.Queries))
	copy(result, c.Queries)
	return result
}

// Helper methods

func (c *Config) moveConnectionToFront(index int) {
	if index == 0 || index >= len(c.Connections) {
		return
	}
	item := c.Connections[index]
	copy(c.Connections[1:], c.Connections[:index])
	c.Connections[0] = item
}

func (c *Config) moveQueryToFront(index int) {
	if index == 0 || index >= len(c.Queries) {
		return
	}
	item := c.Queries[index]
	copy(c.Queries[1:], c.Queries[:index])
	c.Queries[0] = item
}

func connectionsEqual(a, b ConnectionEntry) bool {
	if a.Type != b.Type {
		return false
	}
	switch a.Type {
	case "sqlite":
		return a.Path == b.Path
	case "postgres":
		return a.Host == b.Host && a.Port == b.Port && a.User == b.User && a.Database == b.Database
	}
	return false
}

// getConfigPath returns the path to the config file using XDG conventions.
func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}
	return filepath.Join(configDir, appName, configFileName), nil
}
