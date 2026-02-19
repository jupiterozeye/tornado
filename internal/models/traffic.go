// Package models - Traffic metrics for real-time charts.
//
// This file defines structures for tracking database traffic and performance
// metrics. These are used by the telemetry package to collect data and by
// the dashboard screen to display charts via ntcharts.
//
// Key Learning - Metrics Design:
//   - Think about what's useful to measure
//   - Consider aggregation (per-second, per-minute averages)
//   - Plan for real-time updates
//
// TODO: Define all metric-related structures:
//   - [ ] TrafficSnapshot for point-in-time metrics
//   - [ ] TrafficHistory for storing historical data
//   - [ ] QueryMetrics for individual query tracking
package models

import (
	"time"
)

// TrafficSnapshot represents database metrics at a point in time.
// These are collected periodically and used to draw charts.
//
// Key Learning - Time Series Data:
//   - Each snapshot has a timestamp
//   - Snapshots can be aggregated over time windows
//   - Used directly by ntcharts line charts
//
// TODO: Complete this structure with useful metrics
type TrafficSnapshot struct {
	// Timestamp is when this snapshot was taken
	Timestamp time.Time

	// ===== Connection Metrics =====

	// ActiveConnections is the current number of connections
	// For SQLite, this is typically 1 (single connection)
	ActiveConnections int

	// TotalConnections is the total connections made since start
	TotalConnections int64

	// ===== Query Metrics =====

	// QueriesPerSecond is the queries executed in the last second
	QueriesPerSecond float64

	// TotalQueries is the total queries executed since connection
	TotalQueries int64

	// AverageQueryTime is the average query execution time
	AverageQueryTime time.Duration

	// ===== Query Type Breakdown =====

	// SelectCount is the number of SELECT queries
	SelectCount int64

	// InsertCount is the number of INSERT queries
	InsertCount int64

	// UpdateCount is the number of UPDATE queries
	UpdateCount int64

	// DeleteCount is the number of DELETE queries
	DeleteCount int64

	// ===== Performance Metrics =====

	// SlowQueries is the count of queries exceeding slow query threshold
	SlowQueries int64

	// SlowQueryThreshold is what's considered "slow"
	SlowQueryThreshold time.Duration

	// ===== Data Transfer =====

	// BytesRead is total bytes read from database
	BytesRead int64

	// BytesWritten is total bytes written to database
	BytesWritten int64

	// ===== Error Tracking =====

	// ErrorCount is the total number of errors
	ErrorCount int64

	// LastError is the most recent error message
	LastError string

	// LastErrorTime is when the last error occurred
	LastErrorTime time.Time
}

// TrafficHistory maintains a time-series of traffic snapshots.
// This is used by charts to show historical data.
//
// TODO: Implement ring buffer or fixed-size history
// Consider using a circular buffer to limit memory usage.
type TrafficHistory struct {
	// Snapshots holds the historical data
	Snapshots []TrafficSnapshot

	// MaxSize is the maximum number of snapshots to keep
	MaxSize int

	// currentIndex is the next write position (for circular buffer)
	currentIndex int
}

// NewTrafficHistory creates a new history buffer.
//
// TODO: Implement
func NewTrafficHistory(maxSize int) *TrafficHistory {
	return &TrafficHistory{
		Snapshots: make([]TrafficSnapshot, 0, maxSize),
		MaxSize:   maxSize,
	}
}

// Add adds a new snapshot to the history.
//
// TODO: Implement circular buffer logic
func (h *TrafficHistory) Add(snapshot TrafficSnapshot) {
	// TODO: Implement circular buffer
	// If at capacity, overwrite oldest
}

// Last returns the most recent N snapshots.
//
// TODO: Implement
func (h *TrafficHistory) Last(n int) []TrafficSnapshot {
	// TODO: Return last n snapshots for charting
	return nil
}

// QueryMetrics tracks metrics for a single query execution.
// Used for detailed query analysis and slow query logging.
//
// TODO: Implement query-level tracking
type QueryMetrics struct {
	// Query is the SQL statement
	Query string

	// StartedAt is when the query began
	StartedAt time.Time

	// Duration is how long the query took
	Duration time.Duration

	// RowsAffected is the number of rows (for SELECT, this is returned rows)
	RowsAffected int64

	// WasError indicates if the query failed
	WasError bool

	// ErrorMessage contains error details if failed
	ErrorMessage string

	// QueryType is the type: SELECT, INSERT, UPDATE, DELETE, etc.
	QueryType string
}

// IsSlow returns true if this query exceeded the slow query threshold.
//
// TODO: Implement
func (m *QueryMetrics) IsSlow(threshold time.Duration) bool {
	return m.Duration > threshold
}

// TrafficStats represents aggregated statistics over a time period.
// Used for displaying summary information on the dashboard.
//
// TODO: Define aggregation logic
type TrafficStats struct {
	// Period is the time range these stats cover
	Period time.Duration

	// QueriesTotal is the total queries in the period
	QueriesTotal int64

	// QueriesPerSecond is the average QPS over the period
	QueriesPerSecond float64

	// AverageLatency is the average query time
	AverageLatency time.Duration

	// P95Latency is the 95th percentile latency
	P95Latency time.Duration

	// P99Latency is the 99th percentile latency
	P99Latency time.Duration

	// ErrorRate is the percentage of queries that failed
	ErrorRate float64
}

// ChartDataPoint is a simple x,y point for charting.
// ntcharts expects data in this or similar format.
//
// TODO: Use appropriate ntcharts types
type ChartDataPoint struct {
	X float64 // Usually time (unix timestamp)
	Y float64 // The metric value
}
