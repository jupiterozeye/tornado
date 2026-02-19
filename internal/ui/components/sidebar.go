// Package components - Sidebar navigation component.
//
// This file implements Sidebar - a vertical navigation panel showing:
//   - List of tables/views
//   - Schema selector (PostgreSQL)
//   - Quick actions
//
// Layout (envisioned):
//
//	┌──────────────┐
//	│ 📁 Tables    │
//	├──────────────┤
//	│ 🗃️ users     │
//	│ 🗃️ posts     │
//	│ 🗃️ comments  │
//	│              │
//	│ 📁 Views     │
//	├──────────────┤
//	│ 👁️ user_view │
//	│              │
//	│ [r] Refresh  │
//	└──────────────┘
//
// TODO: Implement the sidebar:
//   - [ ] Define Sidebar struct wrapping bubbles/list
//   - [ ] Implement NewSidebar constructor
//   - [ ] Implement Init, Update, View methods
//   - [ ] Add table grouping (tables, views, indexes)
//   - [ ] Add search/filter for tables
//   - [ ] Add schema switching (PostgreSQL)
//   - [ ] Add refresh button
//   - [ ] Handle selection events
//
// Bubbles Components to Use:
//   - list.Model for the table list
//
// References:
//   - https://github.com/charmbracelet/bubbles#list
package components

import (
	tea "github.com/charmbracelet/bubbletea"

)

// Sidebar is a navigation component for browsing database objects.
// It shows tables, views, and other database objects in a list.
//
// TODO: Complete the implementation
type Sidebar struct {
	// TODO: Add these fields
	//
	// ===== Core Component =====
	// list         list.Model  // bubbles/list
	//
	// ===== Data =====
	// tables       []string
	// views        []string
	// currentIndexes []string
	// schemas      []string
	// currentSchema string
	//
	// ===== UI State =====
	// width        int
	// height       int
	// focused      bool
	// filterText   string
	//
	// ===== Styling =====
	// styles       *styles.Styles
}

// TableItem represents a table in the sidebar list.
// Implements list.Item interface from bubbles.
type TableItem struct {
	Name     string
	Type     string // "table", "view", "index"
	RowCount int
	Desc     string // description (renamed to avoid conflict with Description method)
}

// FilterValue implements list.Item interface for filtering.
func (t TableItem) FilterValue() string {
	return t.Name
}

// Title implements list.DefaultItem interface.
func (t TableItem) Title() string {
	return t.Name
}

// Description implements list.DefaultItem interface.
func (t TableItem) Description() string {
	return t.Desc
}

// NewSidebar creates a new sidebar component.
//
// TODO: Initialize list with default items
func NewSidebar(width, height int) *Sidebar {
	return &Sidebar{
		// TODO: Initialize fields
		// list: list.New([]list.Item{}, list.NewDefaultDelegate(), width, height),
		// width: width,
		// height: height,
	}
}

// Init returns the initial command for the sidebar.
func (s *Sidebar) Init() tea.Cmd {
	return nil
}

// Update handles messages for the sidebar.
//
// Key events to handle:
//   - Up/Down: Navigate items
//   - Enter: Select item (send TableSelectedMsg)
//   - /: Start filter
//   - Esc: Clear filter
//   - r: Refresh list
//   - s: Open schema selector (PostgreSQL)
//
// TODO: Implement sidebar interaction
func (s *Sidebar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// TODO: Implement message handling
	//
	// switch msg := msg.(type) {
	// case tea.KeyMsg:
	//     switch msg.String() {
	//     case "enter":
	//         if item, ok := s.list.SelectedItem().(TableItem); ok {
	//             return s, func() tea.Msg {
	//                 return TableSelectedMsg{Name: item.Name}
	//             }
	//         }
	//     }
	// }
	//
	// var cmd tea.Cmd
	// s.list, cmd = s.list.Update(msg)
	// return s, cmd

	return s, nil
}

// View renders the sidebar.
//
// TODO: Render list with styling
func (s *Sidebar) View() string {
	// TODO: Return rendered list
	// return s.list.View()
	return "Sidebar - TODO"
}

// SetTables sets the list of tables to display.
//
// TODO: Implement
func (s *Sidebar) SetTables(tables []string) {
	// s.tables = tables
	// Convert to list.Items and update
}

// SetViews sets the list of views to display.
//
// TODO: Implement
func (s *Sidebar) SetViews(views []string) {
	// s.views = views
}

// SetSchemas sets the available schemas (PostgreSQL).
//
// TODO: Implement
func (s *Sidebar) SetSchemas(schemas []string) {
	// s.schemas = schemas
}

// SetCurrentSchema sets the current schema.
//
// TODO: Implement
func (s *Sidebar) SetCurrentSchema(schema string) {
	// s.currentSchema = schema
}

// Focus sets the focus state.
func (s *Sidebar) Focus(focused bool) {
	// s.focused = focused
}

// SetSize updates the sidebar dimensions.
func (s *Sidebar) SetSize(width, height int) {
	// s.width = width
	// s.height = height
	// s.list.SetSize(width, height)
}

// SelectedTable returns the currently selected table name.
//
// TODO: Implement
func (s *Sidebar) SelectedTable() string {
	// if item, ok := s.list.SelectedItem().(TableItem); ok {
	//     return item.Name
	// }
	return ""
}

// Refresh reloads the table list from the database.
//
// TODO: Implement
func (s *Sidebar) Refresh() tea.Cmd {
	// Return command to reload tables
	return nil
}

// Filter filters the table list by name.
//
// TODO: Implement
func (s *Sidebar) Filter(term string) {
	// s.list.Filter(term)
}

// ClearFilter clears the filter.
//
// TODO: Implement
func (s *Sidebar) ClearFilter() {
	// s.list.ResetFilter()
}

// TableSelectedMsg is sent when a table is selected in the sidebar.
type TableSelectedMsg struct {
	Name string
}

// SchemaSelectedMsg is sent when a schema is selected.
type SchemaSelectedMsg struct {
	Name string
}

// SidebarRefreshMsg is sent when the sidebar list should be refreshed.
type SidebarRefreshMsg struct{}
