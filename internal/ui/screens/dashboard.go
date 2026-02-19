// Package screens - Dashboard screen for traffic visualization.
//
// This screen provides real-time monitoring of database activity:
//   - Queries per second over time (line chart)
//   - Query time distribution (bar chart)
//   - Connection statistics
//   - Slow query tracking
//
// Layout (envisioned):
//
//	┌─────────────────────────────────────────────────────────┐
//	│ [Tables] [Schemas]  [Query] [Dashboard]                 │
//	├─────────────────────────────────────────────────────────┤
//	│  Queries/sec (last 60s)    │  Query Time Distribution   │
//	│  ┌─────────────────────┐   │  ┌─────────────────────┐   │
//	│  │     ╭─╮             │   │  │  Quick (<10ms)  ████│   │
//	│  │    ╭╯ ╰╮            │   │  │  Medium (10-100)███ │   │
//	│  │   ╭╯   ╰╮      ╭──╮ │   │  │  Slow (>100ms)  ███│   │
//	│  │───╯      ╰──────╯  ╰─│   │  └─────────────────────┘   │
//	│  └─────────────────────┘   │                            │
//	├────────────────────────────┴────────────────────────────┤
//	│  Stats: 1234 queries | 45 errors | 12 slow | 2.3ms avg │
//	│  Last Error: duplicate key value violates unique...     │
//	└─────────────────────────────────────────────────────────┘
//
// Key Learning - Real-time Updates:
//   - Use tea.Tick or tea.Every for periodic updates
//   - The telemetry collector sends metrics messages
//   - Charts update automatically with new data
//
// TODO: Implement the dashboard screen:
//   - [ ] Define DashboardModel struct
//   - [ ] Implement real-time metrics subscription
//   - [ ] Implement line chart for queries/sec (ntcharts)
//   - [ ] Implement bar chart for query distribution (ntcharts)
//   - [ ] Implement sparklines for mini-stats
//   - [ ] Add stats summary display
//   - [ ] Add slow query list
//   - [ ] Add error log
//
// ntcharts Components to Use:
//   - streamlinechart for real-time queries/sec
//   - barchart for query time distribution
//   - sparkline for mini metrics
//
// References:
//   - https://github.com/NimbleMarkets/ntcharts
//   - https://github.com/NimbleMarkets/ntcharts/blob/main/examples/README.md
package screens

import (
	tea "github.com/charmbracelet/bubbletea"
	"time"

	"github.com/jupiterozeye/tornado/internal/db"
	"github.com/jupiterozeye/tornado/internal/models"
	"github.com/jupiterozeye/tornado/internal/telemetry"
	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

// DashboardModel is the model for the dashboard/monitoring screen.
//
// TODO: Add all necessary fields
type DashboardModel struct {
	// TODO: Add these fields
	//
	// ===== Database =====
	// db            db.Database
	//
	// ===== Telemetry =====
	// collector     *telemetry.Collector
	// currentStats  *models.TrafficSnapshot
	// history       *models.TrafficHistory
	//
	// ===== UI State =====
	// width         int
	// height        int
	//
	// ===== Charts =====
	// Use ntcharts components:
	// qpsChart      *streamlinechart.Model    // queries per second
	// distChart     *barchart.Model           // query time distribution
	// connectionsSparkline *sparkline.Model   // connection count
	//
	// ===== Display Options =====
	// timeRange     time.Duration   // how much history to show
	// paused        bool            // pause updates
	//
	// ===== Styling =====
	// styles        *styles.Styles
}

// NewDashboardModel creates a new dashboard screen model.
//
// TODO: Initialize charts and start metrics collection
func NewDashboardModel(database db.Database) *DashboardModel {
	return &DashboardModel{
		// TODO: Initialize fields
		// db: database,
		// timeRange: 60 * time.Second,
		// history: models.NewTrafficHistory(60), // 60 data points
	}
}

// Init returns the initial command for the dashboard screen.
// Should start the telemetry collection loop.
//
// Key Learning - Periodic Updates:
//
//	Use tea.Tick or tea.Every to send periodic update messages.
//	This creates a loop: Update -> Cmd -> Msg -> Update -> ...
//
// TODO: Start telemetry collection
func (m *DashboardModel) Init() tea.Cmd {
	// TODO: Return command to start collecting metrics
	// return tea.Tick(time.Second, func(t time.Time) tea.Msg {
	//     return MetricsTickMsg{Time: t}
	// })
	return nil
}

// Update handles messages for the dashboard screen.
//
// Key events to handle:
//   - p: Pause/resume updates
//   - +/-: Change time range
//   - r: Reset statistics
//   - e: Export metrics to file
//
// Messages to handle:
//   - tea.KeyMsg: Keyboard input
//   - tea.WindowSizeMsg: Terminal resize
//   - MetricsTickMsg: Time to collect new metrics
//   - TrafficUpdateMsg: New metrics from collector
//
// TODO: Implement complete dashboard interaction
func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// TODO: Implement message handling
	//
	// switch msg := msg.(type) {
	// case tea.WindowSizeMsg:
	//     m.width = msg.Width
	//     m.height = msg.Height
	//     m.updateChartSizes()
	//
	// case tea.KeyMsg:
	//     switch msg.String() {
	//     case "p":
	//         m.paused = !m.paused
	//     case "+":
	//         m.timeRange += 30 * time.Second
	//     case "-":
	//         if m.timeRange > 30*time.Second {
	//             m.timeRange -= 30 * time.Second
	//         }
	//     case "r":
	//         m.resetStats()
	//     }
	//
	// case MetricsTickMsg:
	//     // Collect new metrics and schedule next tick
	//     return m, tea.Batch(
	//         m.collectMetrics(),
	//         tea.Tick(time.Second, func(t time.Time) tea.Msg {
	//             return MetricsTickMsg{Time: t}
	//         }),
	//     )
	//
	// case TrafficUpdateMsg:
	//     m.handleTrafficUpdate(msg.Metrics)
	// }

	return m, nil
}

// View renders the dashboard screen.
//
// TODO: Implement chart layout:
//   - Top row: Line chart (QPS) and bar chart (distribution)
//   - Bottom: Stats summary and sparklines
func (m *DashboardModel) View() string {
	// TODO: Implement view with ntcharts
	//
	// Structure:
	// qpsChart := m.renderQPSChart()
	// distChart := m.renderDistributionChart()
	// stats := m.renderStats()
	// return lipgloss.JoinVertical(...)
	//
	// ntcharts example:
	// slc := streamlinechart.New(width, height)
	// for _, v := range dataPoints {
	//     slc.Push(v)
	// }
	// slc.Draw()
	// return slc.View()

	return "Dashboard Screen\n\nTODO: Implement traffic charts"
}

// Helper methods
// TODO: Implement these

func (m *DashboardModel) collectMetrics() tea.Cmd {
	// TODO: Query database for current metrics
	// Return TrafficUpdateMsg with collected data
	return nil
}

func (m *DashboardModel) handleTrafficUpdate(metrics models.TrafficSnapshot) {
	// TODO: Update charts with new data
}

func (m *DashboardModel) updateChartSizes() {
	// TODO: Resize ntcharts components
}

func (m *DashboardModel) resetStats() {
	// TODO: Reset all statistics to zero
}

// MetricsTickMsg is sent periodically to trigger metrics collection.
type MetricsTickMsg struct {
	Time time.Time
}

// TrafficUpdateMsg contains new traffic metrics.
type TrafficUpdateMsg struct {
	Metrics models.TrafficSnapshot
}
