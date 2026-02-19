// Package telemetry collects and aggregates database metrics.
//
// This file implements the metrics collector, which runs as a background
// goroutine to periodically gather performance data from the database.
//
// Key Learning - Goroutines and Channels:
//   - Goroutines are lightweight threads managed by Go runtime
//   - Channels are typed conduits for communication between goroutines
//   - Use select with multiple channels for non-blocking operations
//   - Always provide a way to stop goroutines (done channel pattern)
//
// Key Learning - Bubble Tea Commands:
//   - A tea.Cmd is a function that returns a tea.Msg
//   - Commands run asynchronously in goroutines
//   - When complete, the returned Msg is sent to Update()
//
// TODO: Implement the collector:
//   - [ ] Define Collector struct with channels
//   - [ ] Implement NewCollector constructor
//   - [ ] Implement Start method (returns tea.Cmd)
//   - [ ] Implement Stop method
//   - [ ] Implement periodic collection with ticker
//   - [ ] Query database for metrics
//   - [ ] Send metrics via channel
//
// References:
//   - https://go.dev/tour/concurrency/1 (Goroutines)
//   - https://go.dev/tour/concurrency/2 (Channels)
//   - https://github.com/charmbracelet/bubbletea#commands
package telemetry

import (
	"time"

	"github.com/charmbracelet/bubbletea"

	"github.com/jupiterozeye/tornado/internal/db"
	"github.com/jupiterozeye/tornado/internal/models"
)

// Collector periodically collects database metrics.
// It runs in a background goroutine and sends updates via a channel.
//
// Key Learning - Stopping Goroutines:
//
//	Use a done channel or context to signal the goroutine to stop.
//	This prevents goroutine leaks when the program exits.
//
// TODO: Complete the collector implementation
type Collector struct {
	// db is the database to collect metrics from
	// TODO: Add this field
	// db db.Database

	// interval is how often to collect metrics
	// TODO: Add this field
	// interval time.Duration

	// ===== Concurrency Control =====

	// done is closed to signal the collector to stop
	// Key Learning: Close channel pattern for shutdown
	// TODO: Add this field
	// done chan struct{}

	// ===== Output =====

	// metrics is the channel where metrics are sent
	// Key Learning: Channel for goroutine communication
	// TODO: Add this field
	// metrics chan models.TrafficSnapshot

	// ===== State =====

	// current holds the most recent metrics
	// TODO: Add this field
	// current *models.TrafficSnapshot

	// history stores historical data for charts
	// TODO: Add this field
	// history *models.TrafficHistory

	// queryCount tracks total queries since collection started
	// TODO: Add this field
	// queryCount int64
}

// NewCollector creates a new metrics collector.
//
// TODO: Initialize collector with database and interval
func NewCollector(database db.Database, interval time.Duration) *Collector {
	return &Collector{
		// TODO: Initialize fields
		// db:       database,
		// interval: interval,
		// done:     make(chan struct{}),
		// metrics:  make(chan models.TrafficSnapshot, 10),
		// history:  models.NewTrafficHistory(60), // 60 data points
	}
}

// Start begins collecting metrics.
// It returns a tea.Cmd that the Bubble Tea runtime will execute.
//
// Key Learning - tea.Cmd Pattern:
//
//	A Cmd is just a function: func() tea.Msg
//	The function runs in a goroutine.
//	When it returns, the Msg is sent to Update.
//
// This pattern allows long-running operations (like collecting
// metrics every second) to not block the UI.
//
// TODO: Implement collection loop
func (c *Collector) Start() tea.Cmd {
	// TODO: Return a command that:
	// 1. Creates a ticker for the interval
	// 2. Loops until done channel is closed
	// 3. Collects metrics on each tick
	// 4. Returns MetricsUpdateMsg
	//
	// Example:
	// return func() tea.Msg {
	//     ticker := time.NewTicker(c.interval)
	//     defer ticker.Stop()
	//
	//     for {
	//         select {
	//         case <-c.done:
	//             return nil
	//         case <-ticker.C:
	//             snapshot := c.collect()
	//             return MetricsUpdateMsg{Snapshot: snapshot}
	//         }
	//     }
	// }
	return nil
}

// Stop signals the collector to stop.
// Close the done channel to signal the goroutine to exit.
//
// Key Learning - Graceful Shutdown:
//
//	Always provide a way to stop background goroutines.
//	Closing a channel is a broadcast signal - all receivers wake up.
//
// TODO: Implement stop logic
func (c *Collector) Stop() {
	// TODO: Close the done channel
	// close(c.done)
}

// collect gathers current metrics from the database.
//
// TODO: Implement metric collection
// Different metrics for different databases:
//   - SQLite: Use PRAGMA commands, query sqlite_master
//   - PostgreSQL: Query pg_stat_activity, pg_stat_statements
func (c *Collector) collect() models.TrafficSnapshot {
	snapshot := models.TrafficSnapshot{
		Timestamp: time.Now(),
	}

	// TODO: Collect metrics based on database type
	// switch c.db.GetType() {
	// case "sqlite":
	//     c.collectSQLiteMetrics(&snapshot)
	// case "postgres":
	//     c.collectPostgresMetrics(&snapshot)
	// }

	return snapshot
}

// collectSQLiteMetrics gathers SQLite-specific metrics.
//
// TODO: Implement SQLite metric collection
// Useful queries:
//   - PRAGMA cache_size, PRAGMA page_count
//   - SELECT COUNT(*) FROM sqlite_master
func (c *Collector) collectSQLiteMetrics(snapshot *models.TrafficSnapshot) {
	// TODO: Query SQLite for metrics
}

// collectPostgresMetrics gathers PostgreSQL-specific metrics.
//
// TODO: Implement PostgreSQL metric collection
// Useful queries:
//   - SELECT * FROM pg_stat_activity
//   - SELECT * FROM pg_stat_database
//   - SELECT * FROM pg_stat_statements (if available)
func (c *Collector) collectPostgresMetrics(snapshot *models.TrafficSnapshot) {
	// TODO: Query PostgreSQL for metrics
}

// GetMetrics returns the most recent metrics snapshot.
//
// TODO: Implement getter
func (c *Collector) GetMetrics() *models.TrafficSnapshot {
	return nil // TODO: Return current metrics
}

// GetHistory returns historical metrics for charting.
//
// TODO: Implement getter
func (c *Collector) GetHistory() []models.TrafficSnapshot {
	return nil // TODO: Return history
}

// RecordQuery records a query execution for metrics.
// Call this after each query execution.
//
// TODO: Implement query tracking
func (c *Collector) RecordQuery(queryType string, duration time.Duration, err error) {
	// TODO: Update metrics with query information
	// c.queryCount++
	// Update current snapshot with query stats
}

// MetricsUpdateMsg is sent when new metrics are collected.
// This message triggers the dashboard to update its charts.
type MetricsUpdateMsg struct {
	Snapshot models.TrafficSnapshot
}

// CollectorErrMsg is sent when metric collection fails.
type CollectorErrMsg struct {
	Error error
}
