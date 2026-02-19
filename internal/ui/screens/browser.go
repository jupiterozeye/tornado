// Package screens - Table browser screen for exploring database structure.
//
// This screen is the main interface for:
//   - Browsing database tables and schemas
//   - Viewing table structure (columns, types, keys)
//   - Browsing table data with pagination
//
// Layout (envisioned):
//
//	┌─────────────────────────────────────────────────────────┐
//	│ [Tables] [Schemas]                           [Query] [Dashboard] │
//	├──────────────┬────────────────────────────────────────────────┤
//	│ users        │  Table: users                                 │
//	│ posts        │  ┌──────────────────────────────────────────┐ │
//	│ comments     │  │ id | username | email      | created_at │ │
//	│ tags         │  │ 1  | alice    | alice@...  | 2024-01-01 │ │
//	│              │  │ 2  | bob      | bob@...    | 2024-01-02 │ │
//	│              │  └──────────────────────────────────────────┘ │
//	│              │  [← Prev] Page 1/10 [Next →]                  │
//	└──────────────┴────────────────────────────────────────────────┘
//
// TODO: Implement the browser screen:
//   - [ ] Define BrowserModel struct
//   - [ ] Implement sidebar with table list (bubbles/list)
//   - [ ] Implement table viewer (bubbles/table)
//   - [ ] Implement pagination controls
//   - [ ] Add schema switching (PostgreSQL)
//   - [ ] Add column sorting
//   - [ ] Add row filtering/search
//
// Bubbles Components to Use:
//   - list.Model for the sidebar table list
//   - table.Model for displaying table data
//   - paginator.Model for page navigation
//
// References:
//   - https://github.com/charmbracelet/bubbles#table
//   - https://github.com/charmbracelet/bubbles#list
package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jupiterozeye/tornado/internal/db"
	"github.com/jupiterozeye/tornado/internal/models"
	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

// BrowserModel is the model for the database browser screen.
//
// TODO: Add all necessary fields:
//   - Database reference
//   - Sidebar state (table list)
//   - Main content state (table viewer)
//   - Current selection
type BrowserModel struct {
	// TODO: Add these fields
	//
	// ===== Database =====
	// db            db.Database
	//
	// ===== UI State =====
	// width         int
	// height        int
	// focus         browserFocus  // sidebar or content
	//
	// ===== Sidebar =====
	// sidebarWidth  int
	// tableList     list.Model    // bubbles/list for tables
	// schemas       []string
	// currentSchema string
	//
	// ===== Content =====
	// tableViewer   table.Model   // bubbles/table for data
	// currentTable  string
	// tableSchema   *models.TableSchema
	//
	// ===== Pagination =====
	// currentPage   int
	// totalPages    int
	// pageSize      int
	// totalRows     int64
	//
	// ===== Styling =====
	// styles        *styles.Styles
}

// browserFocus represents which panel has focus.
type browserFocus int

const (
	focusSidebar browserFocus = iota
	focusContent
)

// NewBrowserModel creates a new browser screen model.
//
// TODO: Initialize with database connection
func NewBrowserModel(database db.Database) *BrowserModel {
	return &BrowserModel{
		// TODO: Initialize fields
		// db: database,
		// focus: focusSidebar,
		// pageSize: 50,
	}
}

// Init returns the initial command for the browser screen.
// Should load the table list from the database.
//
// TODO: Return command to load tables
func (m *BrowserModel) Init() tea.Cmd {
	// TODO: Return command to load table list
	// return m.loadTables()
	return nil
}

// Update handles messages for the browser screen.
//
// Key events to handle:
//   - Tab: Switch between sidebar and content
//   - Arrow keys/j/k: Navigate in focused panel
//   - Enter: Select table (in sidebar) or view details
//   - /: Start search/filter
//   - s: Open schema selector (PostgreSQL)
//   - r: Refresh table list
//   - Esc: Clear selection or return to sidebar
//
// TODO: Implement complete browser interaction
func (m *BrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// TODO: Implement message handling
	//
	// switch msg := msg.(type) {
	// case tea.WindowSizeMsg:
	//     m.width = msg.Width
	//     m.height = msg.Height
	//     m.updateComponentSizes()
	//
	// case tea.KeyMsg:
	//     switch msg.String() {
	//     case "tab":
	//         m.toggleFocus()
	//     case "enter":
	//         return m.handleSelect()
	//     case "r":
	//         return m, m.refreshTables()
	//     // ... more key handlers
	//     default:
	//         return m.updateFocused(msg)
	//     }
	//
	// case TablesLoadedMsg:
	//     m.tableList.SetItems(msg.Tables)
	//
	// case TableDataLoadedMsg:
	//     m.updateTableViewer(msg.Data)
	// }

	return m, nil
}

// View renders the browser screen.
//
// TODO: Implement two-panel layout:
//   - Left: Sidebar with table list (bubbles/list)
//   - Right: Table data viewer (bubbles/table)
//   - Use lipgloss.JoinHorizontal to combine panels
func (m *BrowserModel) View() string {
	// TODO: Implement view
	//
	// Structure:
	// sidebar := m.renderSidebar()
	// content := m.renderContent()
	// return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)

	return "Browser Screen\n\nTODO: Implement table browser"
}

// Helper methods
// TODO: Implement these

func (m *BrowserModel) toggleFocus() {
	// TODO: Switch focus between sidebar and content
}

func (m *BrowserModel) loadTables() tea.Cmd {
	// TODO: Return async command to fetch table list
	return nil
}

func (m *BrowserModel) loadTableData(table string) tea.Cmd {
	// TODO: Return async command to fetch table data with pagination
	return nil
}

func (m *BrowserModel) updateComponentSizes() {
	// TODO: Update bubbles components with new dimensions
}

// TablesLoadedMsg is sent when the table list is loaded.
type TablesLoadedMsg struct {
	Tables []string
	Err    error
}

// TableDataLoadedMsg is sent when table data is loaded.
type TableDataLoadedMsg struct {
	TableName string
	Data      *models.QueryResult
	Err       error
}
