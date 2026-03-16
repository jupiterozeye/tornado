// Package layout provides responsive layout management for the browser screen.
package layout


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

