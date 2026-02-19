// Package screens - Query editor screen for running custom SQL.
//
// This screen provides:
//   - SQL query editor (multi-line text input)
//   - Query execution with results display
//   - Query history
//   - Query templates/snippets
//
// Layout (envisioned):
//
//	┌─────────────────────────────────────────────────────────┐
//	│ [Tables] [Schemas]  [Query] [Dashboard]                 │
//	├─────────────────────────────────────────────────────────┤
//	│ SELECT * FROM users WHERE created_at > '2024-01-01';   │
//	│                                                         │
//	│                                                         │
//	├─────────────────────────────────────────────────────────┤
//	│  Results (123 rows, 0.045s)                            │
//	│  ┌──────────────────────────────────────────────────┐  │
//	│  │ id | username | email      | created_at        │  │
//	│  │ 1  | alice    | alice@...  | 2024-01-15        │  │
//	│  └──────────────────────────────────────────────────┘  │
//	│  [History] [Save] [Export]                             │
//	└─────────────────────────────────────────────────────────┘
//
// TODO: Implement the query screen:
//   - [ ] Define QueryModel struct
//   - [ ] Implement SQL editor (bubbles/textarea)
//   - [ ] Implement query execution
//   - [ ] Implement results display (bubbles/table)
//   - [ ] Add query history
//   - [ ] Add keyboard shortcuts (Ctrl+Enter to run)
//   - [ ] Add error handling and display
//   - [ ] Add query templates
//
// Bubbles Components to Use:
//   - textarea.Model for SQL input
//   - table.Model for results
//   - spinner.Model for query in progress
//
// References:
//   - https://github.com/charmbracelet/bubbles#text-area
package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jupiterozeye/tornado/internal/db"
	"github.com/jupiterozeye/tornado/internal/models"
)

// QueryModel is the model for the query editor screen.
//
// TODO: Add all necessary fields
type QueryModel struct {
	// TODO: Add these fields
	//
	// ===== Database =====
	// db            db.Database
	//
	// ===== UI State =====
	// width         int
	// height        int
	// focus         queryFocus  // editor or results
	//
	// ===== Editor =====
	// editor        textarea.Model  // bubbles/textarea for SQL
	// isExecuting   bool
	//
	// ===== Results =====
	// results       table.Model     // bubbles/table for results
	// queryResult   *models.QueryResult
	// executionTime time.Duration
	//
	// ===== History =====
	// history       []models.QueryHistoryItem
	// historyIndex  int
	// showHistory   bool
	//
	// ===== Error State =====
	// error         string
	//
	// ===== Styling =====
	// styles        *styles.Styles
}

// queryFocus represents which panel has focus.
type queryFocus int

const (
	focusEditor queryFocus = iota
	focusResults
)

// NewQueryModel creates a new query screen model.
//
// TODO: Initialize editor with SQL syntax highlighting hints
func NewQueryModel(database db.Database) *QueryModel {
	return &QueryModel{
		// TODO: Initialize fields
		// db: database,
		// focus: focusEditor,
	}
}

// Init returns the initial command for the query screen.
//
// TODO: Consider loading query history
func (m *QueryModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the query screen.
//
// Key events to handle:
//   - Tab: Switch between editor and results
//   - Ctrl+Enter or F5: Execute query
//   - Ctrl+H: Show history
//   - Ctrl+S: Save query to file
//   - Up/Down in history mode: Navigate history
//   - Esc: Close history/error overlay
//
// TODO: Implement complete query interaction
func (m *QueryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// TODO: Implement message handling
	//
	// switch msg := msg.(type) {
	// case tea.WindowSizeMsg:
	//     m.width = msg.Width
	//     m.height = msg.Height
	//     m.updateComponentSizes()
	//
	// case tea.KeyMsg:
	//     // Handle Ctrl+Enter for execution
	//     if msg.String() == "ctrl+enter" || msg.String() == "f5" {
	//         return m, m.executeQuery()
	//     }
	//     // ... more key handlers
	//
	// case QueryExecutedMsg:
	//     m.handleQueryResult(msg)
	// }

	return m, nil
}

// View renders the query screen.
//
// TODO: Implement split-pane layout:
//   - Top: SQL editor (textarea)
//   - Bottom: Results table or error message
func (m *QueryModel) View() string {
	// TODO: Implement view
	//
	// Structure:
	// editor := m.editor.View()
	// results := m.renderResults()
	// return lipgloss.JoinVertical(lipgloss.Left, editor, results)

	return "Query Screen\n\nTODO: Implement query editor"
}

// Helper methods
// TODO: Implement these

func (m *QueryModel) executeQuery() tea.Cmd {
	// TODO: Return async command to execute SQL
	// 1. Get SQL from editor
	// 2. Call db.Query or db.Exec
	// 3. Return QueryExecutedMsg with result
	return nil
}

func (m *QueryModel) handleQueryResult(msg QueryExecutedMsg) {
	// TODO: Update model with query results
}

func (m *QueryModel) updateComponentSizes() {
	// TODO: Update textarea and table with new dimensions
}

func (m *QueryModel) loadFromHistory() {
	// TODO: Load query from history
}

// QueryExecutedMsg is sent when a query finishes execution.
type QueryExecutedMsg struct {
	Result *models.QueryResult
	Err    error
}

// QueryHistoryMsg is sent when loading query history.
type QueryHistoryMsg struct {
	History []models.QueryHistoryItem
	Err     error
}
