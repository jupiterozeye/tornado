// Package layout provides responsive layout management for the browser screen.
package layout

import (
	"github.com/charmbracelet/lipgloss"
)

// Pane represents a UI pane with position and size
type Pane int

const (
	ExplorerPane Pane = iota
	QueryPane
	ResultsPane
)

// Layout manages the 3-pane layout calculations
// - Explorer: 15% width, 100% height (left side)
// - Query: 85% width, 50% height (top-right)
// - Results: 85% width, 50% height (bottom-right)
type Layout struct {
	width  int
	height int
}

// New creates a new layout manager
func New() *Layout {
	return &Layout{}
}

// Update updates the layout dimensions
func (l *Layout) Update(width, height int) {
	l.width = width
	l.height = height
}

// GetExplorerBounds returns the bounds for the Explorer pane
func (l *Layout) GetExplorerBounds() (x, y, width, height int) {
	width = int(float64(l.width) * 0.15)
	if width < 20 {
		width = 20 // Minimum width
	}
	return 0, 0, width, l.height
}

// GetQueryBounds returns the bounds for the Query pane
func (l *Layout) GetQueryBounds() (x, y, width, height int) {
	explorerWidth, _, _, _ := l.GetExplorerBounds()
	x = explorerWidth
	y = 0
	width = l.width - explorerWidth
	height = int(float64(l.height) * 0.5)
	return x, y, width, height
}

// GetResultsBounds returns the bounds for the Results pane
func (l *Layout) GetResultsBounds() (x, y, width, height int) {
	explorerWidth, _, _, _ := l.GetExplorerBounds()
	_, _, _, queryHeight := l.GetQueryBounds()
	x = explorerWidth
	y = queryHeight
	width = l.width - explorerWidth
	height = l.height - queryHeight
	return x, y, width, height
}

// GetExplorerStyle returns the style for the Explorer pane with focus indicator
func (l *Layout) GetExplorerStyle(focused bool) lipgloss.Style {
	_, _, w, h := l.GetExplorerBounds()

	borderColor := lipgloss.Color("238")
	if focused {
		borderColor = lipgloss.Color("99") // Purple when focused
	}

	return lipgloss.NewStyle().
		Width(w-2). // Account for borders
		Height(h-2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)
}

// GetQueryStyle returns the style for the Query pane with focus indicator
func (l *Layout) GetQueryStyle(focused bool) lipgloss.Style {
	_, _, w, h := l.GetQueryBounds()

	borderColor := lipgloss.Color("238")
	if focused {
		borderColor = lipgloss.Color("99")
	}

	return lipgloss.NewStyle().
		Width(w-2).
		Height(h-2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)
}

// GetResultsStyle returns the style for the Results pane with focus indicator
func (l *Layout) GetResultsStyle(focused bool) lipgloss.Style {
	_, _, w, h := l.GetResultsBounds()

	borderColor := lipgloss.Color("238")
	if focused {
		borderColor = lipgloss.Color("99")
	}

	return lipgloss.NewStyle().
		Width(w-2).
		Height(h-2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)
}
