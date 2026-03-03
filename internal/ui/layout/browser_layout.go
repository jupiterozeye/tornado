// Package layout provides responsive layout management for the browser screen.
package layout

import (
	"charm.land/lipgloss/v2"
)

// Pane represents a UI pane with position and size
type Pane int

const (
	ExplorerPane Pane = iota
	QueryPane
	ResultsPane
)

// Layout manages the 3-pane layout calculations
// - Explorer: 22% width, 100% height (left side)
// - Query: 78% width, 50% height (top-right)
// - Results: 78% width, 50% height (bottom-right)
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
	width = int(float64(l.width) * 0.22)
	if width < 24 {
		width = 24 // Minimum width
	}
	if l.width > 0 && width > l.width-40 {
		width = l.width - 40 // Keep enough room for query/results panes
		if width < 16 {
			width = 16
		}
	}
	return 0, 0, width, l.height
}

// GetQueryBounds returns the bounds for the Query pane
func (l *Layout) GetQueryBounds() (x, y, width, height int) {
	_, _, explorerWidth, _ := l.GetExplorerBounds()
	x = explorerWidth
	y = 0
	width = l.width - explorerWidth
	height = int(float64(l.height) * 0.5)
	return x, y, width, height
}

// GetResultsBounds returns the bounds for the Results pane
func (l *Layout) GetResultsBounds() (x, y, width, height int) {
	_, _, explorerWidth, _ := l.GetExplorerBounds()
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
		Width(w).
		Height(h).
		MaxWidth(w).
		MaxHeight(h).
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
		Width(w).
		Height(h).
		MaxWidth(w).
		MaxHeight(h).
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
		Width(w).
		Height(h).
		MaxWidth(w).
		MaxHeight(h).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)
}
