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
	"strings"

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

type QueryMode int

const (
	QueryModeNormal QueryMode = iota
	QueryModeInsert
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
	queryMode   QueryMode
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
		queryMode:     QueryModeNormal,
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

	_, _, ew, eh := m.layoutManager.GetExplorerBounds()
	_, _, qw, qh := m.layoutManager.GetQueryBounds()
	_, _, rw, rh := m.layoutManager.GetResultsBounds()

	// Render explorer
	var explorerContent string
	if m.explorer != nil {
		explorerContent = m.explorer.View()
	} else {
		explorerContent = "Loading..."
	}
	explorerPane := m.renderPane("Explorer", "e", explorerContent, ew, eh, m.focusedPane == PaneExplorer)

	// Render query editor
	queryTitle := "Query [NORMAL]"
	if m.queryMode == QueryModeInsert {
		queryTitle = "Query [INSERT]"
	}
	queryPane := m.renderPane(queryTitle, "q", m.query.View(), qw, qh, m.focusedPane == PaneQuery)

	// Render results
	resultsContent := m.renderResults()
	resultsPane := m.renderPane("Results", "r", resultsContent, rw, rh, m.focusedPane == PaneResults)

	// Combine right side panes vertically
	rightSide := lipgloss.JoinVertical(
		lipgloss.Left,
		queryPane,
		resultsPane,
	)

	main := lipgloss.JoinHorizontal(
		lipgloss.Top,
		explorerPane,
		rightSide,
	)

	return lipgloss.JoinVertical(lipgloss.Left, main, m.renderContextFooter())
}

func (m *BrowserModel) decoratePane(title, key, content string, paneWidth, paneHeight int) string {
	innerWidth := paneWidth - 4   // borders + horizontal padding
	innerHeight := paneHeight - 3 // borders + header row
	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	header := m.styles.Muted.Render(fmt.Sprintf("(%s) %s", key, title))
	body := clipText(content, innerWidth, innerHeight)
	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func (m *BrowserModel) renderPane(title, key, content string, paneWidth, paneHeight int, focused bool) string {
	if paneWidth < 4 {
		paneWidth = 4
	}
	if paneHeight < 3 {
		paneHeight = 3
	}

	innerWidth := paneWidth - 2
	bodyHeight := paneHeight - 2
	if innerWidth < 1 {
		innerWidth = 1
	}
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	label := fmt.Sprintf("(%s) %s", key, title)
	top := makeTopBorder(label, innerWidth)
	body := clipText(content, innerWidth, bodyHeight)
	bodyLines := strings.Split(body, "\n")
	for len(bodyLines) < bodyHeight {
		bodyLines = append(bodyLines, "")
	}

	borderColor := styles.Border
	if focused {
		borderColor = styles.BorderFocus
	}
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	var out []string
	out = append(out, borderStyle.Render("╭"+top+"╮"))
	for i := 0; i < bodyHeight; i++ {
		line := padToWidth(truncateToWidth(bodyLines[i], innerWidth), innerWidth)
		out = append(out, borderStyle.Render("│")+line+borderStyle.Render("│"))
	}
	out = append(out, borderStyle.Render("╰"+strings.Repeat("─", innerWidth)+"╯"))

	return strings.Join(out, "\n")
}

func makeTopBorder(label string, width int) string {
	if width < 1 {
		return ""
	}
	segment := "─ " + label + " "
	segment = truncateToWidth(segment, width)
	if lipgloss.Width(segment) < width {
		segment += strings.Repeat("─", width-lipgloss.Width(segment))
	}
	return segment
}

func padToWidth(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func (m *BrowserModel) renderContextFooter() string {
	var text string
	switch m.focusedPane {
	case PaneExplorer:
		text = "Explorer: j/k Move  Enter Expand/Collapse  s Select TOP 100  r Refresh"
	case PaneQuery:
		if m.queryMode == QueryModeInsert {
			text = "Query INSERT: Esc Normal  Enter Newline"
		} else {
			text = "Query NORMAL: i Insert  Enter Execute  e Explorer  r Results"
		}
	case PaneResults:
		text = "Results: j/k Move  q Query  e Explorer"
	}

	line := truncateToWidth(text, m.width)
	line = padToWidth(line, m.width)
	return m.styles.StatusBar.Render(line)
}

func clipText(content string, width, height int) string {
	lines := strings.Split(content, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for i := range lines {
		lines[i] = truncateToWidth(lines[i], width)
	}
	return strings.Join(lines, "\n")
}

func truncateToWidth(s string, width int) string {
	if width < 1 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}

	max := width
	if width > 1 {
		max = width - 1
	}

	var b strings.Builder
	for _, r := range s {
		next := b.String() + string(r)
		if lipgloss.Width(next) > max {
			break
		}
		b.WriteRune(r)
	}
	if width > 1 {
		return b.String() + "…"
	}
	return b.String()
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
	m.query.SetWidth(qw - 6)
	m.query.SetHeight(qh - 6)

	// Update results table size
	_, _, rw, rh := m.layoutManager.GetResultsBounds()
	m.results.SetWidth(rw - 6)
	m.results.SetHeight(rh - 6)
}

func (m *BrowserModel) updateFocus() {
	// Update explorer focus
	if m.explorer != nil {
		m.explorer.SetFocused(m.focusedPane == PaneExplorer)
	}

	// Update query focus
	if m.focusedPane == PaneQuery {
		if m.queryMode == QueryModeInsert {
			m.query.Focus()
		} else {
			m.query.Blur()
		}
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
		switch m.queryMode {
		case QueryModeNormal:
			switch msg.String() {
			case "i", "a":
				m.queryMode = QueryModeInsert
				m.query.Focus()
				return m, nil
			case "enter":
				return m, m.executeQuery()
			default:
				return m, nil
			}
		case QueryModeInsert:
			if msg.String() == "esc" {
				m.queryMode = QueryModeNormal
				m.query.Blur()
				return m, nil
			}
			var cmd tea.Cmd
			m.query, cmd = m.query.Update(msg)
			return m, cmd
		}
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
