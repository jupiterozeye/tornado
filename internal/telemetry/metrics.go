// Package telemetry - Metrics types and aggregation utilities.
//
// This file defines helper types and functions for working with metrics,
// including aggregation, percentiles, and statistical calculations.
//
// TODO: Implement metrics utilities:
//   - [ ] Rolling average calculator
//   - [ ] Percentile calculation (P95, P99)
//   - [ ] Rate calculation (queries per second)
//   - [ ] Histogram for latency distribution
//
// Key Learning - Data Aggregation:
//
//	Real-world metrics need aggregation for meaningful display.
//	This file provides utilities to transform raw data into
//	display-ready statistics.
package telemetry

import (
	"math"
	"sort"
	"time"

	"github.com/jupiterozeye/tornado/internal/models"
)

// RollingAverage calculates a moving average over a window of values.
// Useful for smoothing out metric fluctuations.
//
// TODO: Implement rolling average
type RollingAverage struct {
	values   []float64
	size     int
	position int
	sum      float64
}

// NewRollingAverage creates a new rolling average calculator.
//
// TODO: Implement
func NewRollingAverage(size int) *RollingAverage {
	return &RollingAverage{
		values: make([]float64, size),
		size:   size,
	}
}

// Add adds a new value to the rolling average.
//
// TODO: Implement
func (r *RollingAverage) Add(value float64) {
	// TODO: Add value, update sum, calculate average
}

// Average returns the current rolling average.
//
// TODO: Implement
func (r *RollingAverage) Average() float64 {
	return r.sum / float64(r.size)
}

// Percentile calculates the given percentile of a slice of values.
// Useful for P95, P99 latency calculations.
//
// Key Learning - Percentiles:
//
//	P95 latency means 95% of requests are faster than this value.
//	More useful than average for understanding tail latency.
//
// TODO: Implement percentile calculation
func Percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// TODO: Implement proper percentile calculation
	// 1. Sort the values
	// 2. Calculate the index for the given percentile
	// 3. Return the value at that index

	sort.Float64s(values)
	index := int(float64(len(values)-1) * p / 100)
	return values[index]
}

// CalculateStats computes statistics from a slice of traffic snapshots.
// Returns aggregated metrics for display.
//
// TODO: Implement statistics calculation
func CalculateStats(snapshots []models.TrafficSnapshot) Stats {
	if len(snapshots) == 0 {
		return Stats{}
	}

	stats := Stats{}
	// TODO: Calculate:
	// - Average QPS
	// - Total queries
	// - Average latency
	// - P95/P99 latency
	// - Error rate

	return stats
}

// Stats holds aggregated statistics over a time period.
type Stats struct {
	// Time range
	Period time.Duration

	// Query metrics
	TotalQueries   int64
	AverageQPS     float64
	PeakQPS        float64
	AverageLatency time.Duration
	P95Latency     time.Duration
	P99Latency     time.Duration

	// Error metrics
	TotalErrors int64
	ErrorRate   float64

	// Connection metrics
	AverageConnections float64
	PeakConnections    int
}

// Histogram creates a histogram of values for bar charts.
// Returns buckets and their counts.
//
// TODO: Implement histogram generation
func Histogram(values []float64, buckets int) ([]float64, []int) {
	if len(values) == 0 {
		return nil, nil
	}

	// Find min and max
	min := values[0]
	max := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Create bucket boundaries
	bucketWidth := (max - min) / float64(buckets)
	boundaries := make([]float64, buckets+1)
	counts := make([]int, buckets)

	for i := 0; i <= buckets; i++ {
		boundaries[i] = min + float64(i)*bucketWidth
	}

	// Count values in each bucket
	for _, v := range values {
		idx := int((v - min) / bucketWidth)
		if idx >= buckets {
			idx = buckets - 1
		}
		counts[idx]++
	}

	return boundaries, counts
}

// RateCalculator calculates rates (per-second values).
// Useful for queries/second, bytes/second, etc.
//
// TODO: Implement rate calculation
type RateCalculator struct {
	lastValue   int64
	lastTime    time.Time
	currentRate float64
}

// NewRateCalculator creates a new rate calculator.
func NewRateCalculator() *RateCalculator {
	return &RateCalculator{
		lastTime: time.Now(),
	}
}

// Update updates the rate with a new value.
// Call this periodically with the current counter value.
//
// TODO: Implement rate calculation
func (r *RateCalculator) Update(value int64) float64 {
	now := time.Now()
	elapsed := now.Sub(r.lastTime).Seconds()

	if elapsed > 0 {
		delta := value - r.lastValue
		r.currentRate = float64(delta) / elapsed
	}

	r.lastValue = value
	r.lastTime = now

	return r.currentRate
}

// Rate returns the current calculated rate.
func (r *RateCalculator) Rate() float64 {
	return r.currentRate
}

// FormatBytes formats byte counts for display.
// Returns human-readable strings like "1.5 MB".
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return formatFloat(float64(bytes)/float64(GB), 1) + " GB"
	case bytes >= MB:
		return formatFloat(float64(bytes)/float64(MB), 1) + " MB"
	case bytes >= KB:
		return formatFloat(float64(bytes)/float64(KB), 1) + " KB"
	default:
		return formatFloat(float64(bytes), 0) + " B"
	}
}

// FormatDuration formats a duration for display.
// Returns human-readable strings like "1.5ms" or "2.3s".
func FormatDuration(d time.Duration) string {
	switch {
	case d >= time.Second:
		return formatFloat(d.Seconds(), 2) + "s"
	case d >= time.Millisecond:
		return formatFloat(float64(d)/float64(time.Millisecond), 2) + "ms"
	case d >= time.Microsecond:
		return formatFloat(float64(d)/float64(time.Microsecond), 2) + "µs"
	default:
		return formatFloat(float64(d)/float64(time.Nanosecond), 0) + "ns"
	}
}

// formatFloat formats a float with specified decimal places.
func formatFloat(f float64, decimals int) string {
	format := "%." + string(rune('0'+decimals)) + "f"
	return trimZeros(sprintf(format, f))
}

func sprintf(format string, a ...interface{}) string {
	return format // Simplified - use fmt.Sprintf in real code
}

func trimZeros(s string) string {
	// TODO: Implement proper zero trimming
	return s
}

// QueryTypeBucket categorizes query times for histograms.
// Used for the "query time distribution" chart.
type QueryTypeBucket string

const (
	BucketQuick    QueryTypeBucket = "quick"  // < 10ms
	BucketMedium   QueryTypeBucket = "medium" // 10ms - 100ms
	BucketSlow     QueryTypeBucket = "slow"   // 100ms - 1s
	BucketVerySlow QueryTypeBucket = "vslow"  // > 1s
)

// CategorizeLatency puts a latency into a bucket category.
func CategorizeLatency(d time.Duration) QueryTypeBucket {
	switch {
	case d < 10*time.Millisecond:
		return BucketQuick
	case d < 100*time.Millisecond:
		return BucketMedium
	case d < time.Second:
		return BucketSlow
	default:
		return BucketVerySlow
	}
}

// LatencyBuckets holds counts for each latency category.
type LatencyBuckets struct {
	Quick    int
	Medium   int
	Slow     int
	VerySlow int
}

// Add adds a latency to the appropriate bucket.
func (l *LatencyBuckets) Add(d time.Duration) {
	switch CategorizeLatency(d) {
	case BucketQuick:
		l.Quick++
	case BucketMedium:
		l.Medium++
	case BucketSlow:
		l.Slow++
	case BucketVerySlow:
		l.VerySlow++
	}
}

// ToSlice returns bucket counts as a slice for charting.
func (l *LatencyBuckets) ToSlice() []int {
	return []int{l.Quick, l.Medium, l.Slow, l.VerySlow}
}

// Labels returns bucket labels for charting.
func (l *LatencyBuckets) Labels() []string {
	return []string{"Quick", "Medium", "Slow", "V.Slow"}
}

// Math helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// clamp ensures a value is within bounds.
func clamp(value, minVal, maxVal float64) float64 {
	return math.Max(minVal, math.Min(maxVal, value))
}
