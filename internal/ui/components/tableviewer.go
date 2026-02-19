// Package components contains reusable UI components for Tornado.
//
// This file implements TableViewer - a reusable component for displaying
// tabular data (like query results or table contents).
//
// Key Learning - Component Composition:
//   - Build small, reusable components
//   - Each component is a complete Bubble Tea model
//   - Parent components embed and delegate to child components
//   - Use bubbles/table for the underlying functionality
//
// TODO: Implement the table viewer:
//   - [ ] Define TableViewer struct wrapping bubbles/table
//   - [ ] Implement NewTableViewer constructor
//   - [ ] Implement Init, Update, View methods
//   - [ ] Add keyboard navigation (arrows, page up/down)
//   - [ ] Add column resizing
//   - [ ] Add row selection
//   - [ ] Add sorting by column header click
//   - [ ] Add search/filter functionality
//
// Bubbles Components to Use:
//   - table.Model for rendering tables
//   - viewport.Model for scrolling (if needed)
//
// References:
//   - https://github.com/charmbracelet/bubbles#table
package components

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jupiterozeye/tornado/internal/models"
)

// TableViewer is a reusable component for displaying tabular data.
// It wraps bubbles/table and adds Tornado-specific functionality.
//
// TODO: Complete the implementation
type TableViewer struct {
	// TODO: Add these fields
	//
	// ===== Core Component =====
	// table    table.Model   // bubbles/table
	//
	// ===== Data =====
	// columns  []table.Column
	// rows     []table.Row
	// data     *models.QueryResult
	//
	// ===== UI State =====
	// width    int
	// height   int
	// focused  bool
	//
	// ===== Styling =====
	// styles   *styles.Styles
}

// NewTableViewer creates a new table viewer component.
//
// TODO: Initialize with default columns and empty data
func NewTableViewer(width, height int) *TableViewer {
	return &TableViewer{
		// TODO: Initialize fields
		// table: table.New(),
		// width: width,
		// height: height,
	}
}

// Init returns the initial command for the table viewer.
// Usually returns nil for static components.
func (t *TableViewer) Init() tea.Cmd {
	return nil
}

// Update handles messages for the table viewer.
//
// Key events to handle:
//   - Arrow keys: Navigate cells
//   - Page Up/Down: Scroll by page
//   - Home/End: Jump to start/end
//   - Enter: Select row (if selection enabled)
//   - s: Sort by current column
//   - /: Start search/filter
//
// TODO: Implement table interaction
func (t *TableViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// TODO: Implement message handling
	//
	// switch msg := msg.(type) {
	// case tea.KeyMsg:
	//     switch msg.String() {
	//     case "up", "k":
	//         t.table.MoveUp(1)
	//     case "down", "j":
	//         t.table.MoveDown(1)
	//     case "left", "h":
	//         t.table.MoveLeft(1)
	//     case "right", "l":
	//         t.table.MoveRight(1)
	//     // ... more key handlers
	//     }
	// }
	//
	// var cmd tea.Cmd
	// t.table, cmd = t.table.Update(msg)
	// return t, cmd

	return t, nil
}

// View renders the table viewer.
//
// TODO: Render table with styling
func (t *TableViewer) View() string {
	// TODO: Return rendered table
	// return t.table.View()
	return "TableViewer - TODO"
}

// SetData populates the table with query results.
//
// TODO: Implement data loading
func (t *TableViewer) SetData(data *models.QueryResult) {
	// TODO:
	// 1. Extract columns from data.Columns
	// 2. Convert rows to table.Row format
	// 3. Update table.Model with new data
	// 4. Adjust column widths
}

// SetColumns sets the table columns.
//
// TODO: Implement column configuration
func (t *TableViewer) SetColumns(columns []string) {
	// TODO: Convert to table.Column and set
}

// SetRows sets the table data rows.
//
// TODO: Implement row loading
func (t *TableViewer) SetRows(rows [][]interface{}) {
	// TODO: Convert to table.Row format and set
}

// Focus sets the focus state for styling.
func (t *TableViewer) Focus(focused bool) {
	// t.focused = focused
}

// SelectedRow returns the currently selected row index.
//
// TODO: Implement selection tracking
func (t *TableViewer) SelectedRow() int {
	return 0 // TODO: Return actual selection
}

// SetSize updates the table dimensions.
func (t *TableViewer) SetSize(width, height int) {
	// t.width = width
	// t.height = height
	// t.table.SetWidth(width)
	// t.table.SetHeight(height)
}

// Sort sorts the table by the specified column.
//
// TODO: Implement sorting
func (t *TableViewer) Sort(columnIndex int, ascending bool) {
	// TODO: Sort rows by column
}

// Filter filters rows by a search term.
//
// TODO: Implement filtering
func (t *TableViewer) Filter(term string) {
	// TODO: Filter visible rows
}

// ClearFilter removes any active filter.
func (t *TableViewer) ClearFilter() {
	// TODO: Reset to showing all rows
}
