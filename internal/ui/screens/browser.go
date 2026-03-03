// Package screens - Table browser screen for exploring database structure.
//
// This screen implements a 3-pane lazygit-style layout:
//   - Explorer (15% width, full height, left): Tree browser for database objects
//   - Query (85% width, 50% height, top-right): SQL editor with syntax highlighting
//   - Results (85% width, 50% height, bottom-right): Query results display
//
// Key bindings:
//   - e: Focus Explorer
//   - q: Focus Query
//   - r: Focus Results
//   - Ctrl+Enter: Execute query (when Query focused)
//
// Explorer navigation:
//   - j/k: Navigate up/down
//   - h: Collapse node or go to parent
//   - l/Enter: Expand node
//   - s: Select table (SELECT TOP 100)
//   - r: Refresh tree
//
// References:
//   - https://github.com/charmbracelet/bubbles#table
//   - https://github.com/charmbracelet/bubbles#list
package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jupiterozeye/tornado/internal/db"
	"github.com/jupiterozeye/tornado/internal/models"
	"github.com/jupiterozeye/tornado/internal/ui/components"
	"github.com/jupiterozeye/tornado/internal/ui/layout"
	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

// Pane represents which pane has focus
type Pane int

const (
	PaneExplorer Pane = iota
	PaneQuery
	PaneResults
)

// BrowserModel is the model for the database browser screen.
type BrowserModel struct {
	// Database connection
	db db.Database

	// Layout manager
	layoutManager *layout.Layout

	// Components
	explorer *components.ExplorerModel
	query    textarea.Model
	results  table.Model

	// State
	width       int
	height      int
	focusedPane Pane
	styles      *styles.Styles

	// Query results
	currentResults *models.QueryResult
	queryError     string
}

// NewBrowserModel creates a new browser screen model.
func NewBrowserModel(database db.Database) *BrowserModel {
	s := styles.Default()
	l := layout.New()

	// Create query editor
	query := textarea.New()
	query.Placeholder = "Enter SQL query..."
	query.SetHeight(10)
	query.SetWidth(80)
	query.ShowLineNumbers = true

	// Create results table
	results := table.New(
		table.WithColumns([]table.Column{}),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(10),
	)
	results.SetStyles(table.Styles{
		Header:   styles.TableHeader(),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
		Selected: lipgloss.NewStyle().Foreground(styles.Primary).Bold(true),
	})

	return &BrowserModel{
		db:            database,
		layoutManager: l,
		query:         query,
		results:       results,
		focusedPane:   PaneExplorer,
		styles:        s,
	}
}

// Init returns the initial command for the browser screen.
func (m *BrowserModel) Init() tea.Cmd {
	return m.initExplorer()
}

// Update handles messages for the browser screen.
func (m *BrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layoutManager.Update(msg.Width, msg.Height)
		m.updateComponentSizes()

	case tea.KeyMsg:
		// Global key bindings
		switch msg.String() {
		case "e":
			m.focusedPane = PaneExplorer
			m.updateFocus()
			return m, nil
		case "q":
			m.focusedPane = PaneQuery
			m.updateFocus()
			return m, nil
		case "r":
			m.focusedPane = PaneResults
			m.updateFocus()
			return m, nil
		case "ctrl+enter":
			if m.focusedPane == PaneQuery {
				return m, m.executeQuery()
			}
		}

		// Route to focused component
		return m.routeKeyMsg(msg)

	case components.TableSelectedMsg:
		// User pressed 's' on a table in explorer
		m.query.SetValue("SELECT * FROM " + msg.Name + " LIMIT 100;")
		m.focusedPane = PaneQuery
		m.updateFocus()
		return m, nil

	case QueryExecutedMsg:
		if msg.Err != nil {
			m.queryError = msg.Err.Error()
			m.currentResults = nil
		} else {
			m.queryError = ""
			m.currentResults = msg.Result
			m.updateResultsTable()
		}
		m.focusedPane = PaneResults
		m.updateFocus()
		return m, nil
	}

	// Pass messages to explorer
	if m.explorer != nil {
		_, cmd := m.explorer.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the browser screen.
func (m *BrowserModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Get pane styles
	explorerStyle := m.layoutManager.GetExplorerStyle(m.focusedPane == PaneExplorer)
	queryStyle := m.layoutManager.GetQueryStyle(m.focusedPane == PaneQuery)
	resultsStyle := m.layoutManager.GetResultsStyle(m.focusedPane == PaneResults)

	// Render explorer
	var explorerContent string
	if m.explorer != nil {
		explorerContent = m.explorer.View()
	} else {
		explorerContent = "Loading..."
	}
	explorerPane := explorerStyle.Render(explorerContent)

	// Render query editor
	queryPane := queryStyle.Render(m.query.View())

	// Render results
	resultsContent := m.renderResults()
	resultsPane := resultsStyle.Render(resultsContent)

	// Combine right side panes vertically
	rightSide := lipgloss.JoinVertical(
		lipgloss.Left,
		queryPane,
		resultsPane,
	)

	// Join explorer with right side horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		explorerPane,
		rightSide,
	)
}

// Helper methods

func (m *BrowserModel) initExplorer() tea.Cmd {
	return func() tea.Msg {
		_, _, w, h := m.layoutManager.GetExplorerBounds()
		m.explorer = components.NewExplorerModel(m.db, w, h)
		return m.explorer.Init()()
	}
}

func (m *BrowserModel) updateComponentSizes() {
	// Update explorer size
	if m.explorer != nil {
		_, _, w, h := m.layoutManager.GetExplorerBounds()
		m.explorer.SetSize(w, h)
	}

	// Update query editor size
	_, _, qw, qh := m.layoutManager.GetQueryBounds()
	m.query.SetWidth(qw - 4) // Account for borders/padding
	m.query.SetHeight(qh - 4)

	// Update results table size
	_, _, rw, rh := m.layoutManager.GetResultsBounds()
	m.results.SetWidth(rw - 4)
	m.results.SetHeight(rh - 4)
}

func (m *BrowserModel) updateFocus() {
	// Update explorer focus
	if m.explorer != nil {
		m.explorer.SetFocused(m.focusedPane == PaneExplorer)
	}

	// Update query focus
	if m.focusedPane == PaneQuery {
		m.query.Focus()
	} else {
		m.query.Blur()
	}

	// Update results focus
	if m.focusedPane == PaneResults {
		m.results.Focus()
	} else {
		m.results.Blur()
	}
}

func (m *BrowserModel) routeKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.focusedPane {
	case PaneExplorer:
		if m.explorer != nil {
			_, cmd := m.explorer.Update(msg)
			return m, cmd
		}
	case PaneQuery:
		var cmd tea.Cmd
		m.query, cmd = m.query.Update(msg)
		return m, cmd
	case PaneResults:
		var cmd tea.Cmd
		m.results, cmd = m.results.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *BrowserModel) executeQuery() tea.Cmd {
	query := m.query.Value()
	if query == "" {
		return nil
	}

	return func() tea.Msg {
		// Try to determine if it's a query or exec
		upperQuery := ""
		for _, r := range query {
			if r >= 'a' && r <= 'z' {
				upperQuery += string(r - 32)
			} else {
				upperQuery += string(r)
			}
		}

		// Check if it starts with SELECT, WITH, or EXPLAIN
		isQuery := false
		if len(upperQuery) >= 6 && upperQuery[:6] == "SELECT" {
			isQuery = true
		} else if len(upperQuery) >= 4 && upperQuery[:4] == "WITH" {
			isQuery = true
		} else if len(upperQuery) >= 7 && upperQuery[:7] == "EXPLAIN" {
			isQuery = true
		}

		if isQuery {
			result, err := m.db.Query(query)
			return QueryExecutedMsg{Result: result, Err: err}
		} else {
			_, err := m.db.Exec(query)
			if err != nil {
				return QueryExecutedMsg{Err: err}
			}
			// For exec statements, return empty result
			return QueryExecutedMsg{
				Result: &models.QueryResult{
					Columns:  []string{"Result"},
					RowCount: 0,
					Query:    query,
				},
			}
		}
	}
}

func (m *BrowserModel) updateResultsTable() {
	if m.currentResults == nil {
		m.results.SetColumns([]table.Column{})
		m.results.SetRows([]table.Row{})
		return
	}

	// Build columns
	columns := make([]table.Column, len(m.currentResults.Columns))
	for i, col := range m.currentResults.Columns {
		columns[i] = table.Column{
			Title: col,
			Width: 20, // TODO: Calculate based on content
		}
	}

	// Build rows
	rows := make([]table.Row, len(m.currentResults.Rows))
	for i, row := range m.currentResults.Rows {
		rowData := make([]string, len(row))
		for j, val := range row {
			if val == nil {
				rowData[j] = "NULL"
			} else {
				// Convert to string
				switch v := val.(type) {
				case string:
					rowData[j] = v
				case []byte:
					rowData[j] = string(v)
				default:
					rowData[j] = fmt.Sprintf("%v", val)
				}
			}
		}
		rows[i] = rowData
	}

	m.results.SetColumns(columns)
	m.results.SetRows(rows)
}

func (m *BrowserModel) renderResults() string {
	if m.queryError != "" {
		return m.styles.Error.Render("Error: " + m.queryError)
	}

	if m.currentResults == nil {
		return m.styles.Muted.Render("Results will appear here...")
	}

	// Show result info
	info := m.styles.Muted.Render(
		fmt.Sprintf("Query returned %d rows", m.currentResults.RowCount),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		info,
		m.results.View(),
	)
}
