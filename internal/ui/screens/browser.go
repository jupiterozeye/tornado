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
//   - https://charm.land/bubbles/v2#table
//   - https://charm.land/bubbles/v2#list
package screens

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	xansi "github.com/charmbracelet/x/ansi"

	"github.com/jupiterozeye/tornado/internal/config"
	"github.com/jupiterozeye/tornado/internal/db"
	"github.com/jupiterozeye/tornado/internal/models"
	"github.com/jupiterozeye/tornado/internal/ui/components"
	"github.com/jupiterozeye/tornado/internal/ui/layout"
	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

// Pane represents which pane has focus
type Pane int

const (
	PaneNone     Pane = -1
	PaneExplorer Pane = iota
	PaneQuery
	PaneResults
)

type QueryMode int

const (
	QueryModeNormal QueryMode = iota
	QueryModeInsert
	QueryModeVisual
	QueryModeVisualLine
)

type browserThemeItem struct{ name string }

func (i browserThemeItem) Title() string       { return i.name }
func (i browserThemeItem) Description() string { return "" }
func (i browserThemeItem) FilterValue() string { return i.name }

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
	width         int
	height        int
	focusedPane   Pane
	queryMode     QueryMode
	styles        *styles.Styles
	showExplorer  bool
	leaderActive  bool // menu popup is visible
	themeMenu     bool
	themeList     list.Model
	themeOriginal string
	statusMsg     string
	maximizedPane Pane

	// Query results
	currentResults      *models.QueryResult
	filteredResults     *models.QueryResult // For filtered view
	queryError          string
	resultsFilter       string // Fuzzy filter text
	resultsFilterActive bool   // Filter input mode active
	resultsCursorCol    int    // Selected column in results table
	resultsScrollCol    int    // Horizontal scroll offset for table
	showPreview         bool   // Preview popup visible
	previewContent      string // Content to preview
	previewTitle        string // Title for preview popup
	showCopyMenu        bool   // Copy menu popup visible

	// Autocomplete
	autocomplete *AutocompleteModel

	// Schema cache for autocomplete
	tables  []string
	columns map[string][]string

	// Visual mode selection tracking
	visualStart struct {
		row int
		col int
	}
	visualEnd struct {
		row int
		col int
	}
	yankBuffer string

	// Pending normal-mode key for multi-key commands (gg, dd, yy)
	pendingNormalKey string

	// Context for cancelling background operations
	ctx    context.Context
	cancel context.CancelFunc

	// Track if cleanup has been called
	cleanedUp bool
}

// NewBrowserModel creates a new browser screen model.
func NewBrowserModel(database db.Database) *BrowserModel {
	s := styles.Default()
	l := layout.New()

	// Create query editor
	query := textarea.New()
	query.Placeholder = ""
	query.SetHeight(10)
	query.SetWidth(80)
	query.ShowLineNumbers = false // Hide line numbers to remove scrollbar appearance
	applyTextAreaStyles(&query)

	// Create results table
	results := table.New(
		table.WithColumns([]table.Column{}),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(10),
	)
	applyTableStyles(&results)

	themeItems := make([]list.Item, 0, len(styles.AvailableThemes()))
	for _, t := range styles.AvailableThemes() {
		themeItems = append(themeItems, browserThemeItem{name: t})
	}

	// Create custom delegate with proper background styling
	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(styles.Text).
		Background(styles.BgDark)
	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(styles.BgDark)
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(styles.Primary).
		Background(styles.BgDark).
		Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(styles.BgDark)
	delegate.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(styles.BgDark)
	delegate.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(styles.BgDark)
	delegate.Styles.FilterMatch = lipgloss.NewStyle().
		Foreground(styles.Primary).
		Background(styles.BgDark).
		Bold(true)
	// Hide help by setting empty help functions
	delegate.ShortHelpFunc = func() []key.Binding { return nil }
	delegate.FullHelpFunc = func() [][]key.Binding { return nil }
	delegate.SetSpacing(0)
	delegate.SetHeight(1)

	themeList := list.New(themeItems, delegate, 20, len(themeItems))
	themeList.SetShowStatusBar(false)
	themeList.SetFilteringEnabled(false)
	themeList.SetShowTitle(false) // Hide the internal title since we use the dialog box title
	// Hide help/navigation commands and spacing
	themeList.SetShowHelp(false)
	delegate.SetSpacing(0)
	// Set proper background styles for the list
	themeList.Styles.TitleBar = lipgloss.NewStyle().Background(styles.BgDark)
	themeList.Styles.Title = lipgloss.NewStyle().Background(styles.BgDark)
	themeList.Styles.Spinner = lipgloss.NewStyle().Background(styles.BgDark)
	themeList.Styles.Filter = textinput.Styles{
		Focused: textinput.StyleState{
			Text:        lipgloss.NewStyle().Foreground(styles.Text).Background(styles.BgDark),
			Placeholder: lipgloss.NewStyle().Foreground(styles.TextMuted).Background(styles.BgDark),
			Prompt:      lipgloss.NewStyle().Foreground(styles.Primary).Background(styles.BgDark),
		},
		Blurred: textinput.StyleState{
			Text:        lipgloss.NewStyle().Foreground(styles.Text).Background(styles.BgDark),
			Placeholder: lipgloss.NewStyle().Foreground(styles.TextMuted).Background(styles.BgDark),
			Prompt:      lipgloss.NewStyle().Foreground(styles.Primary).Background(styles.BgDark),
		},
	}
	themeList.Styles.DefaultFilterCharacterMatch = lipgloss.NewStyle().
		Foreground(styles.Primary).
		Background(styles.BgDark).
		Bold(true)
	themeList.Styles.StatusBar = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(styles.BgDark)
	themeList.Styles.StatusEmpty = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(styles.BgDark)
	themeList.Styles.StatusBarActiveFilter = lipgloss.NewStyle().
		Foreground(styles.Text).
		Background(styles.BgDark)
	themeList.Styles.StatusBarFilterCount = lipgloss.NewStyle().
		Foreground(styles.Primary).
		Background(styles.BgDark)
	themeList.Styles.NoItems = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(styles.BgDark)
	themeList.Styles.PaginationStyle = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(styles.BgDark)
	themeList.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(styles.BgDark)
	themeList.Styles.ActivePaginationDot = lipgloss.NewStyle().
		Foreground(styles.Primary).
		Background(styles.BgDark)
	themeList.Styles.InactivePaginationDot = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(styles.BgDark)
	themeList.Styles.ArabicPagination = lipgloss.NewStyle().
		Foreground(styles.Text).
		Background(styles.BgDark)
	themeList.Styles.DividerDot = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(styles.BgDark)

	ctx, cancel := context.WithCancel(context.Background())

	m := &BrowserModel{
		db:            database,
		layoutManager: l,
		query:         query,
		results:       results,
		focusedPane:   PaneExplorer,
		queryMode:     QueryModeNormal,
		styles:        s,
		showExplorer:  true,
		themeList:     themeList,
		maximizedPane: PaneNone,
		autocomplete:  NewAutocompleteModel(),
		columns:       make(map[string][]string),
		ctx:           ctx,
		cancel:        cancel,
	}

	return m
}

// SchemaLoadedMsg is sent when schema loading completes in the background.
type SchemaLoadedMsg struct {
	Tables  []string
	Columns map[string][]string
}

// ExplorerInitMsg is sent when the explorer has been initialized in the background.
type ExplorerInitMsg struct {
	Explorer *components.ExplorerModel
	InnerMsg tea.Msg
}

// loadSchemaCmd returns a tea.Cmd that loads the schema in a goroutine.
func (m *BrowserModel) loadSchemaCmd() tea.Cmd {
	db := m.db
	ctx := m.ctx
	return func() tea.Msg {
		if db == nil {
			return SchemaLoadedMsg{}
		}

		// Check context before starting
		select {
		case <-ctx.Done():
			return SchemaLoadedMsg{}
		default:
		}

		// Use the Database interface instead of SQLite-specific SQL
		tables, err := db.ListTables()
		if err != nil {
			return SchemaLoadedMsg{}
		}

		// Check context after query
		select {
		case <-ctx.Done():
			return SchemaLoadedMsg{}
		default:
		}

		columns := make(map[string][]string)

		// Query for columns of each table using the Database interface
		for _, table := range tables {
			// Check context before each table query
			select {
			case <-ctx.Done():
				return SchemaLoadedMsg{Tables: tables, Columns: columns}
			default:
			}

			schema, err := db.DescribeTable(table)
			if err != nil {
				continue
			}
			var cols []string
			for _, col := range schema.Columns {
				cols = append(cols, col.Name)
			}
			columns[table] = cols
		}

		return SchemaLoadedMsg{Tables: tables, Columns: columns}
	}
}

// Init returns the initial command for the browser screen.
func (m *BrowserModel) Init() tea.Cmd {
	return tea.Batch(m.initExplorer(), m.loadSchemaCmd())
}

// Update handles messages for the browser screen.
func (m *BrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		usableHeight := msg.Height - 1 // reserve one line for contextual footer
		if usableHeight < 3 {
			usableHeight = 3
		}
		m.layoutManager.Update(msg.Width, usableHeight)
		m.updateComponentSizes()

	case tea.KeyPressMsg:
		// Handle preview dialog first
		if m.showPreview {
			if msg.String() == "esc" || msg.String() == "q" {
				m.showPreview = false
				m.statusMsg = ""
				return m, nil
			}
			return m, nil
		}

		if m.themeMenu {
			return m.handleThemeMenuKey(msg)
		}

		if m.showCopyMenu {
			return m.handleCopyMenuKey(msg)
		}

		if m.leaderActive {
			return m.handleLeaderKey(msg)
		}

		if m.focusedPane == PaneExplorer {
			handled, cmd := m.handleExplorerActionKey(msg)
			if handled {
				return m, cmd
			}
		}

		// When in Query pane with INSERT mode, route all keys directly to query editor
		// This prevents global shortcuts like 'e', 'q', 'r' from interfering with typing
		if m.focusedPane == PaneQuery && m.queryMode == QueryModeInsert {
			// Check if autocomplete is visible and handle its keys
			if m.autocomplete.Visible {
				handled, suggestion := m.autocomplete.HandleKey(msg)
				if handled {
					if suggestion != "" {
						// Apply the suggestion
						m.applyAutocompleteSuggestion(suggestion)
					}
					return m, nil
				}
			}

			switch msg.String() {
			case "ctrl+enter":
				return m, m.executeQuery()
			case "esc":
				m.autocomplete.Visible = false
				m.queryMode = QueryModeNormal
				m.query.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.query, cmd = m.query.Update(msg)
				// Trigger autocomplete after typing (use text length as cursor pos)
				cursorPos := len(m.query.Value())
				return m, tea.Batch(cmd, TriggerAutocomplete(m.query.Value(), cursorPos))
			}
		}

		// When filtering results, route all keys to filter input/navigation.
		// This prevents global shortcuts (e/q/r/space...) from hijacking typing.
		if m.focusedPane == PaneResults && m.resultsFilterActive {
			return m.handleResultsKey(msg)
		}

		// Global key bindings (only processed when NOT in INSERT mode)
		switch msg.String() {
		case "space":
			// Show leader menu immediately
			m.leaderActive = true
			m.themeMenu = false
			m.statusMsg = ""
			return m, nil
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
			m.resultsCursorCol = 0        // Reset column cursor for new results
			m.resultsScrollCol = 0        // Reset horizontal scroll
			m.resultsFilterActive = false // Exit filter mode
			m.resultsFilter = ""
			m.filteredResults = nil
			m.updateResultsTable()
		}
		m.focusedPane = PaneResults
		m.updateFocus()
		return m, nil

	case SchemaLoadedMsg:
		m.tables = msg.Tables
		m.columns = msg.Columns
		return m, nil

	case ExplorerInitMsg:
		m.explorer = msg.Explorer
		return m, func() tea.Msg { return msg.InnerMsg }

	case TriggerAutocompleteMsg:
		// Only process if still in INSERT mode and focused on query
		if m.focusedPane != PaneQuery || m.queryMode != QueryModeInsert {
			return m, nil
		}
		// Check if query text matches current state
		if msg.QueryText != m.query.Value() {
			return m, nil
		}

		ctx := getQueryContext(msg.QueryText, msg.CursorPos)
		suggestions := getSuggestions(ctx, m.tables, m.columns)

		if len(suggestions) > 0 {
			m.autocomplete.Suggestions = suggestions
			m.autocomplete.Selected = 0
			m.autocomplete.Visible = true
			m.autocomplete.TriggerPos = msg.CursorPos
		} else {
			m.autocomplete.Visible = false
		}
		return m, nil
	}

	// Pass messages to explorer
	if m.explorer != nil {
		_, cmd := m.explorer.Update(msg)
		return m, cmd
	}

	return m, nil
}

// renderHighlightedQuery renders the query editor content with SQL syntax highlighting
// and a visible cursor, replacing the textarea's default View().
func (m *BrowserModel) renderHighlightedQuery(width, height int) string {
	if width < 1 || height < 1 {
		return ""
	}

	text := m.query.Value()
	lines := strings.Split(text, "\n")

	cursorRow := m.query.Line()
	cursorCol := m.query.Column()
	scrollY := m.query.ScrollYOffset()

	visibleEnd := scrollY + height
	if visibleEnd > len(lines) {
		visibleEnd = len(lines)
	}

	showCursor := m.focusedPane == PaneQuery
	isVisual := m.queryMode == QueryModeVisual || m.queryMode == QueryModeVisualLine

	// Compute visual selection range in absolute char indices
	var selStart, selEnd int
	if isVisual {
		selStart, selEnd, _ = m.selectedQueryRange()
	}

	bgStyle := lipgloss.NewStyle().Background(styles.BgDark)
	cursorLineBg := lipgloss.NewStyle().Background(styles.BgLight)
	blockCursorStyle := lipgloss.NewStyle().Reverse(true)
	insertCursorStyle := lipgloss.NewStyle().Foreground(styles.Text).Background(styles.BgDark)
	visualStyle := lipgloss.NewStyle().Background(styles.PrimaryBg).Foreground(styles.TextBold)

	var rendered []string

	// Track block comment state from beginning through all lines before visible area
	inBlockComment := false
	for i := 0; i < scrollY && i < len(lines); i++ {
		_, inBlockComment = components.HighlightSQL(lines[i], inBlockComment)
	}

	// Track cumulative char offset for visual selection mapping
	lineStartIdx := 0
	for i := 0; i < scrollY && i < len(lines); i++ {
		lineStartIdx += len([]rune(lines[i])) + 1 // +1 for \n
	}

	for i := scrollY; i < visibleEnd; i++ {
		line := lines[i]
		lineRunes := []rune(line)
		lineLen := len(lineRunes)
		isCursorLine := showCursor && i == cursorRow
		preLineComment := inBlockComment

		highlighted, stillInComment := components.HighlightSQL(line, inBlockComment)
		inBlockComment = stillInComment

		// Check if this line has any visual selection
		lineEndIdx := lineStartIdx + lineLen
		hasSelection := isVisual && lineStartIdx < selEnd && lineEndIdx > selStart

		if !isCursorLine && !hasSelection {
			rendered = append(rendered, bgStyle.Render(padToWidth(highlighted, width)))
			lineStartIdx = lineEndIdx + 1
			continue
		}

		if hasSelection {
			// Render line with visual selection highlighting
			// Calculate which columns of this line are selected
			selColStart := 0
			if selStart > lineStartIdx {
				selColStart = selStart - lineStartIdx
			}
			selColEnd := lineLen
			if selEnd < lineEndIdx {
				selColEnd = selEnd - lineStartIdx
			}
			if selColStart < 0 {
				selColStart = 0
			}
			if selColEnd > lineLen {
				selColEnd = lineLen
			}

			var parts string
			bc := preLineComment
			if selColStart > 0 {
				var seg string
				seg, bc = components.HighlightSQL(string(lineRunes[:selColStart]), bc)
				parts += seg
			}
			if selColEnd > selColStart {
				selText := string(lineRunes[selColStart:selColEnd])
				parts += visualStyle.Render(selText)
				_, bc = components.HighlightSQL(selText, bc)
			}
			if selColEnd < lineLen {
				seg, _ := components.HighlightSQL(string(lineRunes[selColEnd:]), bc)
				parts += seg
			}
			if lineLen == 0 {
				// Empty selected line: show a highlighted space
				parts = visualStyle.Render(" ")
			}
			rendered = append(rendered, bgStyle.Render(padToWidth(parts, width)))
			lineStartIdx = lineEndIdx + 1
			continue
		}

		// Cursor line rendering
		if m.queryMode == QueryModeInsert {
			// INSERT mode: thin bar cursor inserted between characters
			if lineLen == 0 {
				// Empty line: just show the bar cursor
				bar := insertCursorStyle.Render("│")
				rendered = append(rendered, cursorLineBg.Render(padToWidth(bar, width)))
			} else {
				col := cursorCol
				if col < 0 {
					col = 0
				}
				if col > lineLen {
					col = lineLen
				}
				var beforeHL, afterHL string
				bc := preLineComment
				if col > 0 {
					beforeHL, bc = components.HighlightSQL(string(lineRunes[:col]), bc)
				}
				bar := insertCursorStyle.Render("│")
				if col < lineLen {
					afterHL, _ = components.HighlightSQL(string(lineRunes[col:]), bc)
				}
				fullLine := padToWidth(beforeHL+bar+afterHL, width)
				rendered = append(rendered, cursorLineBg.Render(fullLine))
			}
			lineStartIdx = lineEndIdx + 1
			continue
		}

		// NORMAL mode: block cursor on the character
		if lineLen == 0 {
			cursorStr := blockCursorStyle.Background(styles.BgLight).Render(" ")
			pad := ""
			if width > 1 {
				pad = cursorLineBg.Render(strings.Repeat(" ", width-1))
			}
			rendered = append(rendered, cursorStr+pad)
			lineStartIdx = lineEndIdx + 1
			continue
		}

		col := cursorCol
		if col >= lineLen {
			col = lineLen - 1
		}
		if col < 0 {
			col = 0
		}

		var beforeHL, afterHL string
		bc := preLineComment
		if col > 0 {
			beforeHL, bc = components.HighlightSQL(string(lineRunes[:col]), bc)
		}
		_, bc2 := components.HighlightSQL(string(lineRunes[col:col+1]), bc)
		if col+1 < lineLen {
			afterHL, _ = components.HighlightSQL(string(lineRunes[col+1:]), bc2)
		}

		cursorRendered := blockCursorStyle.Background(styles.BgLight).Render(string(lineRunes[col : col+1]))
		fullLine := padToWidth(beforeHL+cursorRendered+afterHL, width)
		rendered = append(rendered, cursorLineBg.Render(fullLine))
		lineStartIdx = lineEndIdx + 1
	}

	// Pad remaining lines
	for len(rendered) < height {
		rendered = append(rendered, bgStyle.Render(strings.Repeat(" ", width)))
	}

	return strings.Join(rendered, "\n")
}

// View renders the browser screen.
func (m *BrowserModel) View() tea.View {
	if m.width == 0 || m.height == 0 {
		return tea.View{Content: "Loading..."}
	}

	ew, eh, qw, qh, rw, rh := m.paneDimensions()

	// Render explorer
	var explorerContent string
	if m.explorer != nil {
		explorerContent = m.explorer.View().Content
	} else {
		explorerContent = "Loading..."
	}
	explorerPane := ""
	if m.showExplorer {
		explorerPane = m.renderPane("Explorer", "e", explorerContent, ew, eh, m.focusedPane == PaneExplorer, styles.BgDefault)
	}

	// Render query editor - ensure textarea has consistent background
	queryTitle := "Query [NORMAL]"
	switch m.queryMode {
	case QueryModeInsert:
		queryTitle = "Query [INSERT]"
	case QueryModeVisual:
		queryTitle = "Query [VISUAL]"
	case QueryModeVisualLine:
		queryTitle = "Query [VISUAL LINE]"
	}

	// Render query with SQL syntax highlighting
	queryInnerW := maxInt(1, qw-4)
	queryInnerH := maxInt(1, qh-3)
	queryView := m.renderHighlightedQuery(queryInnerW, queryInnerH)
	queryPane := m.renderPane(queryTitle, "q", queryView, qw, qh, m.focusedPane == PaneQuery, styles.BgDark)

	// Render results
	resultsContent := m.renderResults()
	resultsPane := m.renderPane("Results", "r", resultsContent, rw, rh, m.focusedPane == PaneResults, styles.BgDefault)

	// Combine right side panes vertically
	rightSide := lipgloss.JoinVertical(
		lipgloss.Left,
		queryPane,
		resultsPane,
	)

	main := rightSide
	if m.showExplorer {
		main = lipgloss.JoinHorizontal(lipgloss.Top, explorerPane, rightSide)
	}

	if m.maximizedPane != PaneNone {
		switch m.maximizedPane {
		case PaneExplorer:
			main = m.renderPane("Explorer", "e", explorerContent, m.width, m.mainHeight(), m.focusedPane == PaneExplorer, styles.BgDefault)
		case PaneQuery:
			maxW := maxInt(1, m.width-4)
			maxH := maxInt(1, m.mainHeight()-3)
			maxQueryView := m.renderHighlightedQuery(maxW, maxH)
			main = m.renderPane(queryTitle, "q", maxQueryView, m.width, m.mainHeight(), m.focusedPane == PaneQuery, styles.BgDark)
		case PaneResults:
			main = m.renderPane("Results", "r", resultsContent, m.width, m.mainHeight(), m.focusedPane == PaneResults, styles.BgDefault)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, main, m.renderContextFooter())
	base := lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content,
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(styles.BgDefault)))

	// Use lipgloss compositing for overlays
	view := tea.View{Content: base, AltScreen: true}
	if m.leaderActive {
		view.Content = m.renderWithLeaderMenu(base)
	}
	if m.themeMenu {
		view.Content = m.renderWithThemeMenu(base)
	}
	if m.autocomplete.Visible {
		view.Content = m.renderWithAutocomplete(base)
	}
	if m.showPreview {
		view.Content = m.renderWithPreview(base)
	}
	if m.showCopyMenu {
		view.Content = m.renderWithCopyMenu(base)
	}

	return view
}

func (m *BrowserModel) renderPane(title, key, content string, paneWidth, paneHeight int, focused bool, bodyBg color.Color) string {
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

	borderColor := styles.Border
	if focused {
		borderColor = styles.BorderFocus
	}
	borderStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Background(styles.BgDefault)
	bodyStyle := lipgloss.NewStyle().
		Background(bodyBg).
		Width(innerWidth)

	// Pad lines with proper background - render empty lines with background
	for len(bodyLines) < bodyHeight {
		bodyLines = append(bodyLines, bodyStyle.Render(strings.Repeat(" ", innerWidth)))
	}

	var out []string
	out = append(out, borderStyle.Render("╭"+top+"╮"))
	for i := 0; i < bodyHeight; i++ {
		line := truncateToWidth(bodyLines[i], innerWidth)
		out = append(out, borderStyle.Render("│")+bodyStyle.Render(line)+borderStyle.Render("│"))
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

func (m *BrowserModel) mainHeight() int {
	h := m.height - 1
	if h < 3 {
		h = 3
	}
	return h
}

func (m *BrowserModel) paneDimensions() (ew, eh, qw, qh, rw, rh int) {
	mainH := m.mainHeight()
	if !m.showExplorer {
		ew, eh = 0, 0
		qw = m.width
		rw = m.width
		qh = mainH / 2
		rh = mainH - qh
		return
	}

	_, _, ew, eh = m.layoutManager.GetExplorerBounds()
	_, _, qw, qh = m.layoutManager.GetQueryBounds()
	_, _, rw, rh = m.layoutManager.GetResultsBounds()
	return
}

func (m *BrowserModel) renderContextFooter() string {
	if m.leaderActive {
		line := truncateToWidth("COMMANDS: e Explorer  f Maximize  c Connect  x Disconnect  t Theme  h Help  / Search  q Quit", m.width)
		line = padToWidth(line, m.width)
		return m.styles.StatusBar.Render(line)
	}

	var text string
	switch m.focusedPane {
	case PaneExplorer:
		text = m.explorerFooterText()
	case PaneQuery:
		switch m.queryMode {
		case QueryModeInsert:
			text = "Query INSERT: Esc→Normal  Enter Newline  Ctrl+Enter Execute"
		case QueryModeVisual:
			text = "Query VISUAL: y Yank  d Delete  c Change  >/< Indent  Esc→Normal"
		case QueryModeVisualLine:
			text = "Query VISUAL LINE: y Yank  d Delete  c Change  >/< Indent  Esc→Normal"
		default:
			text = "Query NORMAL: i/a Insert  o Open  h/j/k/l Move  w/b Word  dd Del  yy Yank  p Paste  u Undo"
		}
	case PaneResults:
		if m.resultsFilterActive {
			text = fmt.Sprintf("Filter: %s_ | Esc to clear", m.resultsFilter)
		} else if m.resultsFilter != "" {
			text = fmt.Sprintf("Filter: '%s' | Esc to clear", m.resultsFilter)
		} else if m.showPreview {
			text = "Preview: Esc or q to close"
		} else if m.showCopyMenu {
			text = "Copy Menu: c Cell, y Row, a All, e Export, Esc Cancel"
		} else {
			text = "Results: h/l Col  j/k Row  v Preview  d Delete  y Copy  / Filter  x Clear"
		}
	}

	line := truncateToWidth(text, m.width)
	if m.statusMsg != "" {
		status := " | " + m.statusMsg
		line = truncateToWidth(line+status, m.width)
	}
	line = padToWidth(line, m.width)
	return m.styles.StatusBar.Render(line)
}

func (m *BrowserModel) explorerFooterText() string {
	node := (*components.TreeNode)(nil)
	if m.explorer != nil {
		node = m.explorer.CurrentNode()
	}
	if node == nil {
		return "Explorer: j/k Move  Enter Expand/Collapse  s Select TOP 100  r Refresh  Commands <space>  Help ?"
	}

	switch node.Type {
	case components.NodeRoot:
		return "Disconnect: x  New: n  Edit: e  Move: m  Delete: d  Refresh: f  Commands: <space>  Help: ?"
	case components.NodeTable:
		return "Columns: enter  Select TOP 100: s  Refresh: f  Commands: <space>  Help: ?"
	default:
		return "Expand/Collapse: enter  Select TOP 100: s  Refresh: f  Commands: <space>  Help: ?"
	}
}

// handleLeaderKey handles key presses when the leader menu popup is visible.
func (m *BrowserModel) handleLeaderKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.leaderActive = false
		m.statusMsg = ""
		return m, nil
	}

	// Check if we're in copy mode from Results pane
	if m.statusMsg == "COPY: c Cell, y Row, a All" {
		return m.handleCopyCommand(msg.String())
	}

	// Dismiss menu and execute the command
	m.leaderActive = false
	return m.executeLeaderCommand(msg.String())
}

// handleCopyCommand handles copy menu options
func (m *BrowserModel) handleCopyCommand(key string) (tea.Model, tea.Cmd) {
	m.leaderActive = false

	switch key {
	case "c":
		// Copy cell
		return m.copyCell()
	case "y":
		// Copy row
		return m.copyRow()
	case "a":
		// Copy all
		return m.copyAll()
	case "e":
		// Export placeholder
		m.statusMsg = "Export: not implemented yet"
		return m, nil
	default:
		m.statusMsg = ""
		return m, nil
	}
}

func (m *BrowserModel) activeResultSet() *models.QueryResult {
	if m.currentResults == nil {
		return nil
	}
	if m.resultsFilter != "" && m.filteredResults != nil {
		return m.filteredResults
	}
	return m.currentResults
}

func (m *BrowserModel) writeClipboard(value string) bool {
	if err := clipboard.WriteAll(value); err != nil {
		m.statusMsg = "Clipboard failed: " + err.Error()
		return false
	}
	return true
}

// copyCell copies the currently selected cell to clipboard
func (m *BrowserModel) copyCell() (tea.Model, tea.Cmd) {
	row := m.results.SelectedRow()
	if row == nil || m.activeResultSet() == nil {
		m.statusMsg = "No cell selected"
		return m, nil
	}

	// Use the tracked column index
	colIdx := m.resultsCursorCol
	if colIdx >= len(row) {
		colIdx = 0
	}

	value := fmt.Sprintf("%v", row[colIdx])
	m.yankBuffer = value
	if !m.writeClipboard(value) {
		return m, nil
	}
	displayValue := value
	if len(displayValue) > 30 {
		displayValue = displayValue[:27] + "..."
	}
	m.statusMsg = fmt.Sprintf("Copied cell: %s", displayValue)
	return m, nil
}

// copyRow copies the currently selected row to clipboard
func (m *BrowserModel) copyRow() (tea.Model, tea.Cmd) {
	row := m.results.SelectedRow()
	if row == nil {
		m.statusMsg = "No row selected"
		return m, nil
	}

	var values []string
	for _, cell := range row {
		values = append(values, fmt.Sprintf("%v", cell))
	}
	value := strings.Join(values, "\t")
	m.yankBuffer = value
	if !m.writeClipboard(value) {
		return m, nil
	}
	m.statusMsg = "Copied row to clipboard"
	return m, nil
}

// copyAll copies all results to clipboard
func (m *BrowserModel) copyAll() (tea.Model, tea.Cmd) {
	active := m.activeResultSet()
	if active == nil {
		m.statusMsg = "No results to copy"
		return m, nil
	}

	var lines []string

	// Add header
	lines = append(lines, strings.Join(active.Columns, "\t"))

	// Add rows
	for _, row := range active.Rows {
		var values []string
		for _, cell := range row {
			values = append(values, fmt.Sprintf("%v", cell))
		}
		lines = append(lines, strings.Join(values, "\t"))
	}

	value := strings.Join(lines, "\n")
	m.yankBuffer = value
	if !m.writeClipboard(value) {
		return m, nil
	}
	m.statusMsg = fmt.Sprintf("Copied %d rows to clipboard", len(active.Rows))
	return m, nil
}

// handleCopyMenuKey handles key presses in the copy menu
func (m *BrowserModel) handleCopyMenuKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.showCopyMenu = false
		m.statusMsg = ""
		return m, nil
	}

	m.showCopyMenu = false

	switch msg.String() {
	case "c":
		return m.copyCell()
	case "y":
		return m.copyRow()
	case "a":
		return m.copyAll()
	case "e":
		m.statusMsg = "Export: not implemented yet"
		return m, nil
	default:
		m.statusMsg = ""
		return m, nil
	}
}

// executeLeaderCommand runs a leader command by key. Used both during
// leader-pending (direct key combo) and from the leader menu popup.
func (m *BrowserModel) executeLeaderCommand(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "e":
		m.showExplorer = !m.showExplorer
		if !m.showExplorer && m.focusedPane == PaneExplorer {
			m.focusedPane = PaneQuery
		}
		m.maximizedPane = PaneNone
		m.updateComponentSizes()
		m.statusMsg = "Toggled explorer"
		return m, nil
	case "f":
		if m.maximizedPane == m.focusedPane {
			m.maximizedPane = PaneNone
			m.statusMsg = "Restored split layout"
		} else {
			m.maximizedPane = m.focusedPane
			m.statusMsg = "Maximized focused pane"
		}
		m.updateComponentSizes()
		return m, nil
	case "c":
		return m, func() tea.Msg { return RequestConnectMsg{} }
	case "x":
		return m, func() tea.Msg { return RequestConnectMsg{} }
	case "t":
		m.themeMenu = true
		m.themeOriginal = styles.CurrentTheme()
		for i, t := range styles.AvailableThemes() {
			if t == m.themeOriginal {
				m.themeList.Select(i)
				break
			}
		}
		m.leaderActive = false
		m.statusMsg = "Preview with j/k, Enter to save, Esc to cancel"
		return m, nil
	case "h", "?":
		m.statusMsg = "Help: e/q/r focus, space command menu, enter run query in NORMAL"
		return m, nil
	case "/":
		m.statusMsg = "Search: not implemented yet"
		return m, nil
	case "q":
		return m, tea.Quit
	default:
		m.statusMsg = ""
		return m, nil
	}
}

func (m *BrowserModel) applyTheme(name string) bool {
	if !styles.SetTheme(name) {
		return false
	}
	m.styles = styles.Default()
	applyTextAreaStyles(&m.query)
	applyTableStyles(&m.results)
	m.updateThemeListStyles()
	return true
}

func (m *BrowserModel) handleThemeMenuKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Handle all navigation manually to prevent list's default wrapping
	switch msg.String() {
	case "esc":
		if m.themeOriginal != "" {
			m.applyTheme(m.themeOriginal)
		}
		m.themeMenu = false
		m.themeOriginal = ""
		m.statusMsg = ""
		return m, nil
	case "enter":
		if it, ok := m.themeList.SelectedItem().(browserThemeItem); ok {
			if m.applyTheme(it.name) {
				m.statusMsg = "Theme: " + it.name
				if cfg := config.Get(); cfg != nil {
					go cfg.SetTheme(it.name)
				}
			}
		}
		m.themeMenu = false
		m.themeOriginal = ""
		return m, nil
	}

	selected := m.themeList.Index()
	totalItems := len(m.themeList.Items())

	switch msg.String() {
	case "j", "down", "ctrl+n":
		if selected < totalItems-1 {
			m.themeList.Select(selected + 1)
			if it, ok := m.themeList.SelectedItem().(browserThemeItem); ok {
				m.applyTheme(it.name)
			}
		}
		return m, nil
	case "k", "up", "ctrl+p":
		if selected > 0 {
			m.themeList.Select(selected - 1)
			if it, ok := m.themeList.SelectedItem().(browserThemeItem); ok {
				m.applyTheme(it.name)
			}
		}
		return m, nil
	default:
		// Ignore all other keys
		return m, nil
	}
}

func (m *BrowserModel) handleExplorerActionKey(msg tea.KeyPressMsg) (bool, tea.Cmd) {
	node := (*components.TreeNode)(nil)
	if m.explorer != nil {
		node = m.explorer.CurrentNode()
	}

	switch msg.String() {
	case "x":
		return true, func() tea.Msg { return RequestConnectMsg{} }
	case "n":
		m.statusMsg = "New connection: coming soon"
		return true, nil
	case "m":
		m.statusMsg = "Move: coming soon"
		return true, nil
	case "d":
		m.statusMsg = "Delete: coming soon"
		return true, nil
	case "f":
		if m.explorer != nil {
			_, cmd := m.explorer.Update(tea.KeyPressMsg{Text: "r"})
			m.statusMsg = "Explorer refreshed"
			return true, cmd
		}
		return true, nil
	case "e":
		if node != nil && node.Type == components.NodeRoot {
			m.statusMsg = "Edit connection: coming soon"
			return true, nil
		}
	}

	return false, nil
}

// renderWithLeaderMenu uses lipgloss compositing to overlay the leader menu
func (m *BrowserModel) renderWithLeaderMenu(base string) string {
	menuContent := buildLeaderMenuContent()
	menu := renderStyledBox("Commands", menuContent, "esc Close", 38)

	// Position in bottom right
	boxH := len(strings.Split(menu, "\n"))
	boxW := 38
	x := m.width - boxW - 2
	y := m.height - boxH - 1
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	// Use lipgloss compositing
	baseLayer := lipgloss.NewLayer(base)
	menuLayer := lipgloss.NewLayer(menu).X(x).Y(y).Z(1)

	comp := lipgloss.NewCompositor(baseLayer, menuLayer)
	return comp.Render()
}

// renderWithThemeMenu uses lipgloss compositing to overlay the theme menu centered
func (m *BrowserModel) renderWithThemeMenu(base string) string {
	// Manually render theme list to ensure proper width and background
	innerWidth := 22 // 24 - 2 for borders
	bg := styles.BgDark

	var lines []string
	themes := styles.AvailableThemes()
	totalThemes := len(themes)
	cursor := m.themeList.Index()
	// Calculate visible count based on available space (max 10)
	visibleCount := totalThemes
	if visibleCount > 10 {
		visibleCount = 10
	}

	// Calculate viewport to show items around cursor
	startIdx := 0
	if cursor >= visibleCount {
		startIdx = cursor - visibleCount + 1
	}
	endIdx := startIdx + visibleCount
	if endIdx > totalThemes {
		endIdx = totalThemes
	}

	for i := startIdx; i < endIdx; i++ {
		theme := themes[i]
		// Determine style based on selection
		var style lipgloss.Style
		if i == cursor {
			style = lipgloss.NewStyle().
				Foreground(styles.Primary).
				Background(bg).
				Bold(true)
		} else {
			style = lipgloss.NewStyle().
				Foreground(styles.Text).
				Background(bg)
		}

		// Render theme name and pad to fill width
		themeWidth := lipgloss.Width(theme)
		if themeWidth < innerWidth {
			theme = theme + strings.Repeat(" ", innerWidth-themeWidth)
		}
		lines = append(lines, style.Render(theme))
	}

	menu := renderDialogBox("Themes", lines, "", 24)

	// Center position
	boxH := len(strings.Split(menu, "\n"))
	boxW := 24
	x := (m.width - boxW) / 2
	y := (m.height - boxH) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	baseLayer := lipgloss.NewLayer(base)
	menuLayer := lipgloss.NewLayer(menu).X(x).Y(y).Z(1)

	comp := lipgloss.NewCompositor(baseLayer, menuLayer)
	return comp.Render()
}

// renderWithAutocomplete overlays the autocomplete dropdown near the query editor
func (m *BrowserModel) renderWithAutocomplete(base string) string {
	menu := m.autocomplete.Render()
	if menu == "" {
		return base
	}

	// Position autocomplete near the query pane
	_, _, _, _, _, _ = m.paneDimensions()
	boxW := m.autocomplete.Width

	// Position in the query pane area
	var x, y int
	if m.showExplorer {
		// Query pane is on the right side
		_, _, ew, _ := m.layoutManager.GetExplorerBounds()
		x = ew + 4 // After explorer + some padding
	} else {
		x = 2
	}
	// Position near the top of query pane
	y = 4

	// Ensure it doesn't go off screen
	if x+boxW > m.width {
		x = m.width - boxW - 2
	}
	if x < 0 {
		x = 0
	}

	baseLayer := lipgloss.NewLayer(base)
	menuLayer := lipgloss.NewLayer(menu).X(x).Y(y).Z(1)

	comp := lipgloss.NewCompositor(baseLayer, menuLayer)
	return comp.Render()
}

// renderWithCopyMenu overlays the copy menu in the bottom right (like leader menu)
func (m *BrowserModel) renderWithCopyMenu(base string) string {
	bg := styles.BgDark
	keyStyle := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Background(bg)
	textStyle := lipgloss.NewStyle().Background(bg).Foreground(styles.Text)

	// Ensure all menu items have proper background - wrap entire line in textStyle
	menuContent := []string{
		textStyle.Render("  " + keyStyle.Render("c") + textStyle.Render("  Copy Cell")),
		textStyle.Render("  " + keyStyle.Render("y") + textStyle.Render("  Copy Row")),
		textStyle.Render("  " + keyStyle.Render("a") + textStyle.Render("  Copy All")),
		textStyle.Render(""),
		textStyle.Render("  " + keyStyle.Render("e") + textStyle.Render("  Export...")),
		textStyle.Render(""),
		textStyle.Render("  " + keyStyle.Render("esc") + textStyle.Render(" Cancel")),
	}

	menu := renderStyledBox("Copy", menuContent, "", 30)

	// Position in bottom right (same as leader menu)
	boxH := len(strings.Split(menu, "\n"))
	boxW := 30
	x := m.width - boxW - 2
	y := m.height - boxH - 1
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	baseLayer := lipgloss.NewLayer(base)
	menuLayer := lipgloss.NewLayer(menu).X(x).Y(y).Z(1)

	comp := lipgloss.NewCompositor(baseLayer, menuLayer)
	return comp.Render()
}

func buildLeaderMenuLines() []string {
	bg := styles.BgDark
	keyStyle := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Background(bg)
	headStyle := lipgloss.NewStyle().Foreground(styles.Secondary).Bold(true).Background(bg)
	textStyle := lipgloss.NewStyle().Background(bg).Foreground(styles.Text)

	return []string{
		headStyle.Render("Navigation"),
		textStyle.Render("  " + keyStyle.Render("e") + textStyle.Render("  Toggle Explorer")),
		textStyle.Render("  " + keyStyle.Render("f") + textStyle.Render("  Toggle Maximize")),
		textStyle.Render(""),
		headStyle.Render("Connection"),
		textStyle.Render("  " + keyStyle.Render("c") + textStyle.Render("  Connect")),
		textStyle.Render("  " + keyStyle.Render("x") + textStyle.Render("  Disconnect")),
		textStyle.Render(""),
		headStyle.Render("Other"),
		textStyle.Render("  " + keyStyle.Render("t") + textStyle.Render("  Change Theme")),
		textStyle.Render("  " + keyStyle.Render("h") + textStyle.Render("  Help")),
		textStyle.Render("  " + keyStyle.Render("/") + textStyle.Render("  Search")),
		textStyle.Render("  " + keyStyle.Render("q") + textStyle.Render("  Quit")),
	}
}

// buildLeaderMenuContent creates styled content for the leader menu popup
func buildLeaderMenuContent() []string {
	return buildLeaderMenuLines()
}

// renderStyledBox creates a styled popup box similar to renderDialogBox but optimized for menus
func renderStyledBox(title string, body []string, subtitle string, width int) string {
	if width < 14 {
		width = 14
	}

	innerWidth := width - 2
	bg := styles.BgDark
	borderStyle := lipgloss.NewStyle().
		Foreground(styles.BorderFocus).
		Background(bg)
	bodyStyle := lipgloss.NewStyle().
		Background(bg).
		Foreground(styles.Text).
		Width(innerWidth)

	out := make([]string, 0, len(body)+2)
	out = append(out, borderStyle.Render("╭"+makeMenuTopBorder(title, innerWidth)+"╮"))

	for _, line := range body {
		// Ensure each line has proper background by wrapping styled content
		line = truncateToWidth(line, innerWidth)
		// Re-render the line with background to ensure no terminal bleed
		lineWithBg := lipgloss.NewStyle().Background(bg).Render(line)
		out = append(out, borderStyle.Render("│")+bodyStyle.Render(lineWithBg)+borderStyle.Render("│"))
	}

	out = append(out, borderStyle.Render("╰"+makeMenuBottomBorder(subtitle, innerWidth)+"╯"))
	return strings.Join(out, "\n")
}

func makeMenuTopBorder(label string, width int) string {
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

func makeMenuBottomBorder(label string, width int) string {
	if width < 1 {
		return ""
	}
	if label == "" {
		return strings.Repeat("─", width)
	}
	segment := " " + label + " ─"
	if lipgloss.Width(segment) > width {
		return strings.Repeat("─", width)
	}
	left := strings.Repeat("─", width-lipgloss.Width(segment))
	return left + segment
}

// renderWithPreview overlays the preview dialog
func (m *BrowserModel) renderWithPreview(base string) string {
	if !m.showPreview {
		return base
	}

	// Format content with word wrapping
	boxWidth := minInt(60, m.width-10)
	contentLines := wrapText(m.previewContent, boxWidth-4)

	preview := renderDialogBox(m.previewTitle, contentLines, "esc Close", boxWidth)

	// Center position
	boxH := len(strings.Split(preview, "\n"))
	boxW := boxWidth
	x := (m.width - boxW) / 2
	y := (m.height - boxH) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	baseLayer := lipgloss.NewLayer(base)
	previewLayer := lipgloss.NewLayer(preview).X(x).Y(y).Z(1)

	comp := lipgloss.NewCompositor(baseLayer, previewLayer)
	return comp.Render()
}

// wrapText wraps text to fit within maxWidth
func wrapText(text string, maxWidth int) []string {
	if maxWidth < 1 {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		if lipgloss.Width(currentLine+" "+word) <= maxWidth {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)
	return lines
}

type RequestConnectMsg struct{}

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
	return xansi.Truncate(s, width, "")
}

// Helper methods

func (m *BrowserModel) initExplorer() tea.Cmd {
	db := m.db
	lm := m.layoutManager
	return func() tea.Msg {
		_, _, w, h := lm.GetExplorerBounds()
		explorer := components.NewExplorerModel(db, w, h)
		innerMsg := explorer.Init()()
		return ExplorerInitMsg{Explorer: explorer, InnerMsg: innerMsg}
	}
}

func (m *BrowserModel) updateComponentSizes() {
	ew, eh, qw, qh, rw, rh := m.paneDimensions()

	// Handle maximized pane - use full main height for the maximized pane
	if m.maximizedPane == PaneQuery {
		qh = m.mainHeight()
		qw = m.width
	} else if m.maximizedPane == PaneResults {
		rh = m.mainHeight()
		rw = m.width
	} else if m.maximizedPane == PaneExplorer {
		eh = m.mainHeight()
		ew = m.width
	}

	// Update explorer size
	if m.explorer != nil && m.showExplorer {
		m.explorer.SetSize(ew, eh)
	}

	// Update query editor size
	// Border takes 2 lines + header, so content height is qh - 3
	m.query.SetWidth(maxInt(10, qw-6))
	m.query.SetHeight(maxInt(3, qh-3))

	// Update results table size
	m.results.SetWidth(maxInt(10, rw-6))
	m.results.SetHeight(maxInt(3, rh-6))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
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

func (m *BrowserModel) routeKeyMsg(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.focusedPane {
	case PaneExplorer:
		if m.explorer != nil {
			_, cmd := m.explorer.Update(msg)
			return m, cmd
		}
	case PaneQuery:
		return m.handleQueryKey(msg)
	case PaneResults:
		return m.handleResultsKey(msg)
	}
	return m, nil
}

// handleQueryKey handles all key inputs for the query editor with vim-like modal editing
func (m *BrowserModel) handleQueryKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.queryMode {
	case QueryModeNormal:
		return m.handleQueryNormalMode(msg)
	case QueryModeInsert:
		// This should not happen - INSERT mode keys are handled in Update()
		// But handle ESC just in case
		if msg.String() == "esc" {
			m.queryMode = QueryModeNormal
			m.query.Blur()
			return m, nil
		}
		var cmd tea.Cmd
		m.query, cmd = m.query.Update(msg)
		return m, cmd
	case QueryModeVisual:
		return m.handleQueryVisualMode(msg)
	case QueryModeVisualLine:
		return m.handleQueryVisualMode(msg)
	}
	return m, nil
}

// handleQueryNormalMode handles keys in NORMAL mode
func (m *BrowserModel) handleQueryNormalMode(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	k := msg.String()

	// Handle pending multi-key commands (gg, dd, yy)
	if m.pendingNormalKey != "" {
		pending := m.pendingNormalKey
		m.pendingNormalKey = ""
		m.statusMsg = ""
		switch pending + k {
		case "gg":
			m.query.MoveToBegin()
			return m, nil
		case "dd":
			m.deleteCurrentLine()
			return m, nil
		case "yy":
			m.yankCurrentLine()
			return m, nil
		default:
			// Invalid combo, ignore
			return m, nil
		}
	}

	switch k {
	// === Mode switching ===
	case "i":
		m.queryMode = QueryModeInsert
		m.query.Focus()
		return m, nil
	case "a":
		// Append: move cursor one right, then insert
		col := m.query.Column()
		lines := strings.Split(m.query.Value(), "\n")
		row := m.query.Line()
		if row < len(lines) && col < len([]rune(lines[row])) {
			m.query.SetCursorColumn(col + 1)
		}
		m.queryMode = QueryModeInsert
		m.query.Focus()
		return m, nil
	case "I":
		// Insert at beginning of line
		m.query.CursorStart()
		m.queryMode = QueryModeInsert
		m.query.Focus()
		return m, nil
	case "A":
		// Append at end of line
		m.query.CursorEnd()
		m.queryMode = QueryModeInsert
		m.query.Focus()
		return m, nil
	case "o":
		// Open line below
		m.query.CursorEnd()
		m.query.InsertString("\n")
		m.queryMode = QueryModeInsert
		m.query.Focus()
		return m, nil
	case "O":
		// Open line above
		m.query.CursorStart()
		m.query.InsertString("\n")
		m.query.CursorUp()
		m.queryMode = QueryModeInsert
		m.query.Focus()
		return m, nil
	case "v":
		m.queryMode = QueryModeVisual
		m.visualStart.row = m.query.Line()
		m.visualStart.col = m.query.Column()
		m.visualEnd = m.visualStart
		m.statusMsg = "-- VISUAL --"
		return m, nil
	case "V":
		m.queryMode = QueryModeVisualLine
		m.visualStart.row = m.query.Line()
		m.visualStart.col = 0
		m.visualEnd = m.visualStart
		m.statusMsg = "-- VISUAL LINE --"
		return m, nil

	// === Navigation ===
	case "h", "left":
		col := m.query.Column()
		if col > 0 {
			m.query.SetCursorColumn(col - 1)
		}
		return m, nil
	case "j", "down":
		m.query.CursorDown()
		return m, nil
	case "k", "up":
		m.query.CursorUp()
		return m, nil
	case "l", "right":
		col := m.query.Column()
		lines := strings.Split(m.query.Value(), "\n")
		row := m.query.Line()
		if row < len(lines) && col < len([]rune(lines[row]))-1 {
			m.query.SetCursorColumn(col + 1)
		}
		return m, nil
	case "w":
		m.wordForward()
		return m, nil
	case "b":
		m.wordBackward()
		return m, nil
	case "0", "home":
		m.query.CursorStart()
		return m, nil
	case "$", "end":
		m.query.CursorEnd()
		return m, nil
	case "G":
		m.query.MoveToEnd()
		return m, nil
	case "pgup":
		m.query.PageUp()
		return m, nil
	case "pgdown":
		m.query.PageDown()
		return m, nil

	// === Multi-key command starters ===
	case "g":
		m.pendingNormalKey = "g"
		m.statusMsg = "g"
		return m, nil
	case "d":
		m.pendingNormalKey = "d"
		m.statusMsg = "d"
		return m, nil
	case "y":
		m.pendingNormalKey = "y"
		m.statusMsg = "y"
		return m, nil

	// === Editing ===
	case "x":
		m.deleteCharAtCursor()
		return m, nil
	case "p":
		if m.yankBuffer != "" {
			m.pasteAfter()
		}
		return m, nil
	case "P":
		if m.yankBuffer != "" {
			m.pasteBefore()
		}
		return m, nil
	case "u":
		// Undo: pass ctrl+z to textarea
		undoMsg := tea.KeyPressMsg(tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 'z'})
		m.query.Focus()
		m.query, _ = m.query.Update(undoMsg)
		m.query.Blur()
		return m, nil

	// === Execute query ===
	case "enter":
		return m, m.executeQuery()

	default:
		return m, nil
	}
}

// wordForward moves the cursor to the start of the next word.
func (m *BrowserModel) wordForward() {
	text := m.query.Value()
	if text == "" {
		return
	}
	runes := []rune(text)
	idx := lineColToIndex(text, m.query.Line(), m.query.Column())
	n := len(runes)

	// Skip current word characters
	for idx < n && !isWordBoundary(runes[idx]) {
		idx++
	}
	// Skip whitespace/punctuation
	for idx < n && isWordBoundary(runes[idx]) {
		idx++
	}

	row, col := indexToLineCol(text, idx)
	m.setQueryCursor(row, col)
}

// wordBackward moves the cursor to the start of the previous word.
func (m *BrowserModel) wordBackward() {
	text := m.query.Value()
	if text == "" {
		return
	}
	runes := []rune(text)
	idx := lineColToIndex(text, m.query.Line(), m.query.Column())

	if idx > len(runes) {
		idx = len(runes)
	}
	// Move back past any whitespace/punctuation
	for idx > 0 && isWordBoundary(runes[idx-1]) {
		idx--
	}
	// Move back through word characters
	for idx > 0 && !isWordBoundary(runes[idx-1]) {
		idx--
	}

	row, col := indexToLineCol(text, idx)
	m.setQueryCursor(row, col)
}

func isWordBoundary(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' ||
		r == ',' || r == ';' || r == '(' || r == ')' ||
		r == '.' || r == '=' || r == '<' || r == '>' ||
		r == '+' || r == '-' || r == '*' || r == '/' ||
		r == '\'' || r == '"' || r == '[' || r == ']'
}

// deleteCharAtCursor deletes the character at the cursor (like vim 'x').
func (m *BrowserModel) deleteCharAtCursor() {
	text := m.query.Value()
	if text == "" {
		return
	}
	runes := []rune(text)
	idx := lineColToIndex(text, m.query.Line(), m.query.Column())
	if idx >= len(runes) {
		return
	}
	newText := string(runes[:idx]) + string(runes[idx+1:])
	m.query.SetValue(newText)
	row, col := indexToLineCol(newText, idx)
	m.setQueryCursor(row, col)
}

// deleteCurrentLine deletes the current line (like vim 'dd').
func (m *BrowserModel) deleteCurrentLine() {
	text := m.query.Value()
	if text == "" {
		return
	}
	lines := strings.Split(text, "\n")
	row := m.query.Line()
	if row >= len(lines) {
		return
	}

	m.yankBuffer = lines[row] + "\n"

	newLines := make([]string, 0, len(lines)-1)
	newLines = append(newLines, lines[:row]...)
	if row+1 < len(lines) {
		newLines = append(newLines, lines[row+1:]...)
	}

	if len(newLines) == 0 {
		m.query.SetValue("")
		m.query.MoveToBegin()
		return
	}
	m.query.SetValue(strings.Join(newLines, "\n"))
	if row >= len(newLines) {
		row = len(newLines) - 1
	}
	m.setQueryCursor(row, 0)
}

// yankCurrentLine copies the current line to the yank buffer (like vim 'yy').
func (m *BrowserModel) yankCurrentLine() {
	text := m.query.Value()
	if text == "" {
		return
	}
	lines := strings.Split(text, "\n")
	row := m.query.Line()
	if row >= len(lines) {
		return
	}
	m.yankBuffer = lines[row] + "\n"
	m.statusMsg = "1 line yanked"
}

// pasteAfter pastes the yank buffer after the cursor (like vim 'p').
func (m *BrowserModel) pasteAfter() {
	if m.yankBuffer == "" {
		return
	}
	// If yank buffer ends with \n, it's a line yank — paste on line below
	if strings.HasSuffix(m.yankBuffer, "\n") {
		text := m.query.Value()
		lines := strings.Split(text, "\n")
		row := m.query.Line()
		content := strings.TrimSuffix(m.yankBuffer, "\n")

		newLines := make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[:row+1]...)
		newLines = append(newLines, content)
		if row+1 < len(lines) {
			newLines = append(newLines, lines[row+1:]...)
		}
		m.query.SetValue(strings.Join(newLines, "\n"))
		m.setQueryCursor(row+1, 0)
	} else {
		// Character-wise paste after cursor
		text := m.query.Value()
		idx := lineColToIndex(text, m.query.Line(), m.query.Column())
		if idx < len(text) {
			idx++ // paste after
		}
		newText := text[:idx] + m.yankBuffer + text[idx:]
		m.query.SetValue(newText)
		row, col := indexToLineCol(newText, idx+len(m.yankBuffer)-1)
		m.setQueryCursor(row, col)
	}
}

// pasteBefore pastes the yank buffer before the cursor (like vim 'P').
func (m *BrowserModel) pasteBefore() {
	if m.yankBuffer == "" {
		return
	}
	if strings.HasSuffix(m.yankBuffer, "\n") {
		text := m.query.Value()
		lines := strings.Split(text, "\n")
		row := m.query.Line()
		content := strings.TrimSuffix(m.yankBuffer, "\n")

		newLines := make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[:row]...)
		newLines = append(newLines, content)
		newLines = append(newLines, lines[row:]...)
		m.query.SetValue(strings.Join(newLines, "\n"))
		m.setQueryCursor(row, 0)
	} else {
		text := m.query.Value()
		idx := lineColToIndex(text, m.query.Line(), m.query.Column())
		newText := text[:idx] + m.yankBuffer + text[idx:]
		m.query.SetValue(newText)
		row, col := indexToLineCol(newText, idx+len(m.yankBuffer)-1)
		m.setQueryCursor(row, col)
	}
}

// handleQueryVisualMode handles keys in VISUAL mode
func (m *BrowserModel) handleQueryVisualMode(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Helper to update visual end after cursor movement
	updateEnd := func() {
		m.visualEnd.row = m.query.Line()
		m.visualEnd.col = m.query.Column()
	}

	switch msg.String() {
	case "esc":
		m.queryMode = QueryModeNormal
		m.statusMsg = ""
		return m, nil
	case "y":
		selected := m.getSelectedQueryText()
		if selected != "" {
			m.yankBuffer = selected
		}
		m.queryMode = QueryModeNormal
		m.statusMsg = "yanked"
		return m, nil
	case "d", "x":
		selected := m.getSelectedQueryText()
		if selected != "" {
			m.yankBuffer = selected
		}
		m.deleteSelectedQueryText()
		m.queryMode = QueryModeNormal
		m.statusMsg = "deleted"
		return m, nil
	case ">":
		m.indentSelectedLines(true)
		m.queryMode = QueryModeNormal
		return m, nil
	case "<":
		m.indentSelectedLines(false)
		m.queryMode = QueryModeNormal
		return m, nil
	case "p":
		if m.yankBuffer != "" {
			m.deleteSelectedQueryText()
			m.query.InsertString(m.yankBuffer)
		}
		m.queryMode = QueryModeNormal
		return m, nil
	case "c":
		m.deleteSelectedQueryText()
		m.queryMode = QueryModeInsert
		m.query.Focus()
		return m, nil

	// Navigation — move cursor explicitly and extend selection
	case "h", "left":
		col := m.query.Column()
		if col > 0 {
			m.query.SetCursorColumn(col - 1)
		}
		updateEnd()
		return m, nil
	case "j", "down":
		m.query.CursorDown()
		updateEnd()
		return m, nil
	case "k", "up":
		m.query.CursorUp()
		updateEnd()
		return m, nil
	case "l", "right":
		col := m.query.Column()
		lines := strings.Split(m.query.Value(), "\n")
		row := m.query.Line()
		if row < len(lines) && col < len([]rune(lines[row]))-1 {
			m.query.SetCursorColumn(col + 1)
		}
		updateEnd()
		return m, nil
	case "w":
		m.wordForward()
		updateEnd()
		return m, nil
	case "b":
		m.wordBackward()
		updateEnd()
		return m, nil
	case "0", "home":
		m.query.CursorStart()
		updateEnd()
		return m, nil
	case "$", "end":
		m.query.CursorEnd()
		updateEnd()
		return m, nil
	case "G":
		m.query.MoveToEnd()
		updateEnd()
		return m, nil
	case "g":
		// gg in visual mode — move to beginning
		m.query.MoveToBegin()
		updateEnd()
		return m, nil
	default:
		return m, nil
	}
}

func (m *BrowserModel) selectedQueryRange() (start, end int, ok bool) {
	text := m.query.Value()
	if text == "" {
		return 0, 0, false
	}
	start = lineColToIndex(text, m.visualStart.row, m.visualStart.col)
	end = lineColToIndex(text, m.visualEnd.row, m.visualEnd.col)
	if m.queryMode == QueryModeVisualLine {
		start = lineColToIndex(text, minInt(m.visualStart.row, m.visualEnd.row), 0)
		endLine := maxInt(m.visualStart.row, m.visualEnd.row)
		lines := strings.Split(text, "\n")
		if endLine >= len(lines) {
			endLine = len(lines) - 1
		}
		end = lineColToIndex(text, endLine, len(lines[endLine]))
	}
	if start > end {
		start, end = end, start
	}
	// Visual selection is inclusive of the character under the cursor
	if m.queryMode == QueryModeVisual && end < len([]rune(text)) {
		end++
	}
	if start == end {
		return 0, 0, false
	}
	return start, end, true
}

func (m *BrowserModel) getSelectedQueryText() string {
	text := m.query.Value()
	start, end, ok := m.selectedQueryRange()
	if !ok || start < 0 || end > len(text) {
		return ""
	}
	return text[start:end]
}

func (m *BrowserModel) deleteSelectedQueryText() {
	text := m.query.Value()
	start, end, ok := m.selectedQueryRange()
	if !ok || start < 0 || end > len(text) {
		return
	}
	newText := text[:start] + text[end:]
	m.query.SetValue(newText)
	row, col := indexToLineCol(newText, start)
	m.setQueryCursor(row, col)
}

func (m *BrowserModel) indentSelectedLines(indent bool) {
	text := m.query.Value()
	if text == "" {
		return
	}
	startLine := minInt(m.visualStart.row, m.visualEnd.row)
	endLine := maxInt(m.visualStart.row, m.visualEnd.row)
	lines := strings.Split(text, "\n")
	if startLine < 0 {
		startLine = 0
	}
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}
	for i := startLine; i <= endLine; i++ {
		if indent {
			lines[i] = "  " + lines[i]
		} else if strings.HasPrefix(lines[i], "  ") {
			lines[i] = lines[i][2:]
		} else if strings.HasPrefix(lines[i], " ") {
			lines[i] = lines[i][1:]
		}
	}
	m.query.SetValue(strings.Join(lines, "\n"))
	m.setQueryCursor(startLine, 0)
}

func (m *BrowserModel) setQueryCursor(line, col int) {
	if line < 0 {
		line = 0
	}
	if col < 0 {
		col = 0
	}
	m.query.MoveToBegin()
	for i := 0; i < line; i++ {
		m.query.CursorDown()
	}
	m.query.SetCursorColumn(col)
}

func lineColToIndex(text string, row, col int) int {
	if row < 0 {
		row = 0
	}
	if col < 0 {
		col = 0
	}
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return 0
	}
	if row >= len(lines) {
		row = len(lines) - 1
	}
	idx := 0
	for i := 0; i < row; i++ {
		idx += len(lines[i]) + 1
	}
	if col > len(lines[row]) {
		col = len(lines[row])
	}
	return idx + col
}

func indexToLineCol(text string, idx int) (row, col int) {
	if idx < 0 {
		idx = 0
	}
	if idx > len(text) {
		idx = len(text)
	}
	row, col = 0, 0
	for i := 0; i < idx; i++ {
		if text[i] == '\n' {
			row++
			col = 0
		} else {
			col++
		}
	}
	return
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// handleResultsKey handles keys when Results pane is focused
func (m *BrowserModel) handleResultsKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.resultsFilterActive {
		switch msg.String() {
		case "esc":
			m.resultsFilterActive = false
			m.resultsFilter = ""
			m.applyFilter()
			m.statusMsg = ""
			return m, nil
		case "enter":
			m.resultsFilterActive = false
			return m, nil
		case "backspace", "ctrl+h":
			r := []rune(m.resultsFilter)
			if len(r) > 0 {
				m.resultsFilter = string(r[:len(r)-1])
			}
		default:
			if len(msg.Text) > 0 {
				m.resultsFilter += msg.Text
			}
		}
		m.applyFilter()
		return m, nil
	}

	switch msg.String() {
	case "v":
		// Preview: show selected cell value
		return m.showPreviewDialog()
	case "d":
		// Delete: create DELETE SQL query
		return m.createDeleteQuery()
	case "y":
		// Copy: open copy menu
		m.showCopyMenu = true
		m.statusMsg = "COPY: c Cell, y Row, a All, e Export, esc Cancel"
		return m, nil
	case "h", "left":
		// Move left (previous column)
		if m.resultsCursorCol > 0 {
			m.resultsCursorCol--
			// Auto-scroll if needed
			if m.resultsCursorCol < m.resultsScrollCol {
				m.resultsScrollCol = m.resultsCursorCol
			}
		}
		return m, nil
	case "l", "right":
		// Move right (next column)
		cols := m.results.Columns()
		if m.resultsCursorCol < len(cols)-1 {
			m.resultsCursorCol++
		}
		return m, nil
	case "x":
		// Clear: clear results
		m.clearResults()
		return m, nil
	case "/":
		// Filter: start filtering
		m.resultsFilter = ""
		m.resultsFilterActive = true
		m.statusMsg = "Filter: type to search, esc to clear"
		return m, nil
	case "esc":
		if m.resultsFilter != "" {
			m.resultsFilter = ""
			m.applyFilter()
			m.statusMsg = ""
			return m, nil
		}
		// Navigation keys
	case "j", "down":
		m.results.MoveDown(1)
		return m, nil
	case "k", "up":
		m.results.MoveUp(1)
		return m, nil
	case "g", "home":
		m.results.GotoTop()
		return m, nil
	case "G", "end":
		m.results.GotoBottom()
		return m, nil
	case "ctrl+d", "pgdown":
		m.results.MoveDown(10)
		return m, nil
	case "ctrl+u", "pgup":
		m.results.MoveUp(10)
		return m, nil
	default:
		// Pass other keys to table for default navigation
		var cmd tea.Cmd
		m.results, cmd = m.results.Update(msg)
		return m, cmd
	}
	return m, nil
}

// showPreviewDialog shows a preview of the selected cell value
func (m *BrowserModel) showPreviewDialog() (tea.Model, tea.Cmd) {
	active := m.activeResultSet()
	row := m.results.SelectedRow()
	if row == nil || active == nil {
		return m, nil
	}

	cursor := m.results.Cursor()
	if cursor < 0 || cursor >= len(active.Rows) {
		return m, nil
	}

	// Use the tracked column index
	colIdx := m.resultsCursorCol
	if colIdx >= len(row) {
		colIdx = 0
	}

	value := fmt.Sprintf("%v", row[colIdx])
	colName := ""
	cols := m.results.Columns()
	if colIdx < len(cols) {
		colName = cols[colIdx].Title
	}

	m.showPreview = true
	m.previewTitle = fmt.Sprintf("Preview: %s", colName)
	m.previewContent = value
	m.statusMsg = "Preview: esc to close"

	return m, nil
}

// createDeleteQuery creates a DELETE SQL query for the selected row
func (m *BrowserModel) createDeleteQuery() (tea.Model, tea.Cmd) {
	active := m.activeResultSet()
	if active == nil || m.db == nil {
		return m, nil
	}

	// Get selected row
	rowIdx := m.results.Cursor()
	if rowIdx < 0 || rowIdx >= len(active.Rows) {
		return m, nil
	}

	row := active.Rows[rowIdx]
	columns := active.Columns

	// Try to extract table name from the query
	tableName := m.extractTableNameFromQuery(active.Query)
	if tableName == "" {
		m.statusMsg = "Cannot determine table name from query"
		return m, nil
	}

	// Build WHERE clause
	var conditions []string
	for i, col := range columns {
		if i >= len(row) {
			continue
		}
		val := row[i]
		if val == nil {
			conditions = append(conditions, fmt.Sprintf("%s IS NULL", col))
		} else {
			switch v := val.(type) {
			case string:
				conditions = append(conditions, fmt.Sprintf("%s = '%s'", col, strings.ReplaceAll(v, "'", "''")))
			default:
				conditions = append(conditions, fmt.Sprintf("%s = %v", col, v))
			}
		}
	}

	if len(conditions) == 0 {
		m.statusMsg = "Cannot create DELETE: no columns found"
		return m, nil
	}

	whereClause := strings.Join(conditions, " AND ")
	deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE %s;", tableName, whereClause)

	// Set the query in the query editor
	m.query.SetValue(deleteQuery)
	m.focusedPane = PaneQuery
	m.updateFocus()
	m.statusMsg = "DELETE query created - review and execute"

	return m, nil
}

// extractTableNameFromQuery attempts to extract the table name from a SELECT query
func (m *BrowserModel) extractTableNameFromQuery(query string) string {
	// Find FROM in uppercased copy but extract name from original string
	upper := strings.ToUpper(query)
	fromIdx := strings.Index(upper, "FROM ")
	if fromIdx == -1 {
		return ""
	}

	// Get everything after FROM from the original query
	afterFrom := query[fromIdx+5:]
	// Split on whitespace or punctuation to get just the table name
	parts := strings.FieldsFunc(afterFrom, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == ',' || r == ';' || r == '(' || r == ')'
	})

	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// clearResults clears the results section
func (m *BrowserModel) clearResults() {
	m.currentResults = nil
	m.filteredResults = nil
	m.resultsFilter = ""
	m.resultsFilterActive = false
	m.resultsCursorCol = 0
	m.resultsScrollCol = 0
	m.results.SetRows([]table.Row{})
	m.results.SetColumns([]table.Column{})
	m.statusMsg = "Results cleared"
}

// applyFilter applies fuzzy filter to results
func (m *BrowserModel) applyFilter() {
	if m.currentResults == nil {
		return
	}

	if m.resultsFilter == "" {
		m.filteredResults = nil
		m.updateResultsTable()
		m.results.GotoTop()
		return
	}

	filter := strings.ToLower(m.resultsFilter)
	var filteredRows [][]any

	for _, row := range m.currentResults.Rows {
		// Check if any cell contains the filter text
		for _, cell := range row {
			cellStr := strings.ToLower(fmt.Sprintf("%v", cell))
			if strings.Contains(cellStr, filter) {
				filteredRows = append(filteredRows, row)
				break
			}
		}
	}

	// Build filtered result
	if m.filteredResults == nil {
		m.filteredResults = &models.QueryResult{}
		*m.filteredResults = *m.currentResults
	}
	m.filteredResults.Rows = filteredRows
	m.filteredResults.RowCount = len(filteredRows)

	// Update table with filtered rows
	rows := make([]table.Row, len(filteredRows))
	for i, row := range filteredRows {
		rowData := make([]string, len(row))
		for j, val := range row {
			if val == nil {
				rowData[j] = "NULL"
			} else {
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

	m.results.SetRows(rows)
	m.results.GotoTop()
}

func highlightFilterMatch(text, filter string) string {
	if filter == "" {
		return text
	}

	lowerText := strings.ToLower(text)
	lowerFilter := strings.ToLower(filter)
	if lowerFilter == "" {
		return text
	}

	highlight := lipgloss.NewStyle().Foreground(styles.BgDefault).Background(styles.Primary).Bold(true)

	var out strings.Builder
	start := 0
	for {
		idx := strings.Index(lowerText[start:], lowerFilter)
		if idx == -1 {
			out.WriteString(text[start:])
			break
		}
		idx += start
		out.WriteString(text[start:idx])
		out.WriteString(highlight.Render(text[idx : idx+len(filter)]))
		start = idx + len(filter)
	}

	return out.String()
}

// applyAutocompleteSuggestion applies the selected autocomplete suggestion
func (m *BrowserModel) applyAutocompleteSuggestion(suggestion string) {
	query := m.query.Value()
	triggerPos := m.autocomplete.TriggerPos

	// Find the start of the current word
	wordStart := triggerPos
	for wordStart > 0 {
		ch := query[wordStart-1]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' ||
			ch == ',' || ch == ';' || ch == '(' || ch == ')' ||
			ch == '=' || ch == '<' || ch == '>' || ch == '+' ||
			ch == '-' || ch == '*' || ch == '/' || ch == '%' {
			break
		}
		wordStart--
	}

	// Replace the current word with the suggestion
	newQuery := query[:wordStart] + suggestion + query[triggerPos:]
	m.query.SetValue(newQuery)

	// Hide autocomplete
	m.autocomplete.Visible = false
}

// Cleanup cancels background operations and prepares for shutdown
func (m *BrowserModel) Cleanup() {
	if m.cleanedUp {
		return
	}
	m.cleanedUp = true
	if m.cancel != nil {
		m.cancel()
	}
}

func (m *BrowserModel) executeQuery() tea.Cmd {
	query := m.query.Value()
	if query == "" {
		return nil
	}

	return func() tea.Msg {
		// Save query to history (async)
		if cfg := config.Get(); cfg != nil {
			go cfg.AddQuery(query)
		}

		startTime := time.Now()

		// Try to determine if it's a query or exec
		upperQuery := strings.ToUpper(query)

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
			if result != nil {
				result.ExecutionTime = time.Since(startTime)
			}
			return QueryExecutedMsg{Result: result, Err: err}
		} else {
			_, err := m.db.Exec(query)
			if err != nil {
				return QueryExecutedMsg{Err: err}
			}
			// For exec statements, return empty result
			return QueryExecutedMsg{
				Result: &models.QueryResult{
					Columns:       []string{"Result"},
					RowCount:      0,
					ExecutionTime: time.Since(startTime),
					Query:         query,
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

	// Clear rows first to prevent index mismatch when columns change
	m.results.SetRows([]table.Row{})
	m.results.SetColumns(columns)
	m.results.SetRows(rows)
}

func (m *BrowserModel) renderResults() string {
	bg := styles.BgDefault
	// Create a filler style that fills the available space
	fillerStyle := lipgloss.NewStyle().Background(bg)

	if m.queryError != "" {
		errContent := fillerStyle.Render(lipgloss.NewStyle().Foreground(styles.Error).Render("Error: " + m.queryError))
		return errContent
	}

	active := m.activeResultSet()
	if active == nil {
		emptyContent := fillerStyle.Render(lipgloss.NewStyle().Foreground(styles.TextMuted).Render("Results will appear here..."))
		return emptyContent
	}

	// Show result info with execution time - fill entire width with background
	var timeStr string
	if m.currentResults.ExecutionTime < time.Millisecond {
		timeStr = fmt.Sprintf("%d µs", m.currentResults.ExecutionTime.Microseconds())
	} else {
		timeStr = fmt.Sprintf("%d ms", m.currentResults.ExecutionTime.Milliseconds())
	}
	infoText := fmt.Sprintf("Query returned %d rows in %s", m.currentResults.RowCount, timeStr)
	if m.resultsFilter != "" {
		infoText = fmt.Sprintf("Showing %d/%d rows in %s (filter: %s)", len(active.Rows), m.currentResults.RowCount, timeStr, m.resultsFilter)
	}
	// Make info line fill width with proper background
	info := lipgloss.NewStyle().
		Background(bg).
		Foreground(styles.TextMuted).
		Width(m.width - 4).
		Render(infoText)

	// Render table with column highlighting
	tableView := m.renderTableWithColumnHighlight(bg)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		info,
		tableView,
	)

	// Ensure the entire content area has background color
	return lipgloss.NewStyle().Background(bg).Render(content)
}

// renderTableWithColumnHighlight renders the table with the current column highlighted
func (m *BrowserModel) renderTableWithColumnHighlight(bg color.Color) string {
	active := m.activeResultSet()
	if active == nil || len(active.Columns) == 0 {
		return ""
	}

	// Get current cursor position
	cursorRow := m.results.Cursor()
	cursorCol := m.resultsCursorCol

	// Ensure cursorCol is valid
	if cursorCol < 0 {
		cursorCol = 0
	}
	if cursorCol >= len(active.Columns) {
		cursorCol = len(active.Columns) - 1
	}

	// Define styles
	headerStyle := lipgloss.NewStyle().
		Foreground(styles.Primary).
		Bold(true).
		Background(bg).
		Padding(0, 1)

	cellStyle := lipgloss.NewStyle().
		Background(bg).
		Foreground(styles.Text).
		Padding(0, 1)

	selectedRowStyle := lipgloss.NewStyle().
		Foreground(styles.TextBold).
		Bold(true).
		Background(lipgloss.Color("240")).
		Padding(0, 1)

	selectedCellStyle := lipgloss.NewStyle().
		Foreground(styles.Primary).
		Bold(true).
		Background(lipgloss.Color("238")).
		Padding(0, 1)

	// Calculate column widths
	colWidths := make([]int, len(active.Columns))
	for i, col := range active.Columns {
		colWidths[i] = lipgloss.Width(col) + 2 // +2 for padding
		if colWidths[i] < 12 {
			colWidths[i] = 12 // Minimum width
		}
	}

	// Calculate total table width
	totalWidth := 0
	for _, w := range colWidths {
		totalWidth += w
	}

	// Get available width from pane dimensions
	_, _, _, _, rw, _ := m.paneDimensions()
	// Use full pane width minus minimal borders (2 chars for box borders)
	availableWidth := rw - 2
	if availableWidth < 20 {
		availableWidth = 20
	}

	// Auto-scroll horizontally to keep cursor in view
	if cursorCol >= m.resultsScrollCol {
		// Check if cursor is beyond visible area
		visibleWidth := 0
		for i := m.resultsScrollCol; i <= cursorCol && i < len(colWidths); i++ {
			visibleWidth += colWidths[i]
		}
		// If cursor column exceeds available width, scroll right
		for visibleWidth > availableWidth && m.resultsScrollCol < cursorCol {
			visibleWidth -= colWidths[m.resultsScrollCol]
			m.resultsScrollCol++
		}
	}

	// Find end column for visible area
	endCol := m.resultsScrollCol
	visibleWidth := 0
	for i := m.resultsScrollCol; i < len(colWidths); i++ {
		if visibleWidth+colWidths[i] > availableWidth && i > m.resultsScrollCol {
			break
		}
		visibleWidth += colWidths[i]
		endCol = i + 1
	}

	// Build header - only show visible columns
	var headerParts []string
	for i := m.resultsScrollCol; i < endCol && i < len(active.Columns); i++ {
		colName := truncateString(active.Columns[i], colWidths[i]-2)
		headerParts = append(headerParts, headerStyle.Width(colWidths[i]).Render(colName))
	}
	header := lipgloss.NewStyle().Background(bg).Render(strings.Join(headerParts, ""))

	// Build rows
	var rows []string
	startIdx := 0
	endIdx := len(active.Rows)

	// Only show visible rows (respect viewport)
	visibleHeight := m.results.Height()
	if visibleHeight > 0 && endIdx > visibleHeight {
		startIdx = cursorRow - visibleHeight/2
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + visibleHeight
		if endIdx > len(active.Rows) {
			endIdx = len(active.Rows)
			startIdx = endIdx - visibleHeight
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	for rowIdx := startIdx; rowIdx < endIdx && rowIdx < len(active.Rows); rowIdx++ {
		row := active.Rows[rowIdx]
		var cellParts []string

		// Only render visible columns
		for colIdx := m.resultsScrollCol; colIdx < endCol && colIdx < len(active.Columns); colIdx++ {
			if colIdx >= len(row) {
				continue
			}
			val := row[colIdx]
			if val == nil {
				val = "NULL"
			}
			cellStr := fmt.Sprintf("%v", val)
			cellStr = truncateString(cellStr, colWidths[colIdx]-2)
			cellStr = highlightFilterMatch(cellStr, m.resultsFilter)

			// Apply appropriate style
			if rowIdx == cursorRow {
				if colIdx == cursorCol {
					// Selected cell (intersection of row and column)
					cellParts = append(cellParts, selectedCellStyle.Width(colWidths[colIdx]).Render(cellStr))
				} else {
					// Other cells in selected row
					cellParts = append(cellParts, selectedRowStyle.Width(colWidths[colIdx]).Render(cellStr))
				}
			} else {
				// Normal cell
				cellParts = append(cellParts, cellStyle.Width(colWidths[colIdx]).Render(cellStr))
			}
		}

		rowStr := lipgloss.NewStyle().Background(bg).Render(strings.Join(cellParts, ""))
		rows = append(rows, rowStr)
	}

	// Combine everything
	var tableParts []string
	tableParts = append(tableParts, header)
	tableParts = append(tableParts, rows...)

	// Calculate total visible width
	visibleWidth = 0
	for i := m.resultsScrollCol; i < endCol && i < len(colWidths); i++ {
		visibleWidth += colWidths[i]
	}

	// Ensure each row fills the full available width
	rowStyle := lipgloss.NewStyle().Background(bg)
	for i := 0; i < len(tableParts); i++ {
		// Extend each row to fill available width
		tableParts[i] = rowStyle.Width(availableWidth).Render(tableParts[i])
	}

	// Wrap entire table in background to prevent any terminal bleed
	return lipgloss.NewStyle().
		Background(bg).
		Width(availableWidth).
		Render(strings.Join(tableParts, "\n"))
}

func applyTextAreaStyles(ta *textarea.Model) {
	s := ta.Styles()

	bg := styles.BgDark
	cursorBg := styles.BgLight

	// Focused styles with visible cursor - ALL components must have theme background
	// Note: Base style should NOT have explicit Width/Height - let textarea handle sizing
	s.Focused.Base = lipgloss.NewStyle().Background(bg)
	s.Focused.Text = lipgloss.NewStyle().Foreground(styles.Text).Background(bg)
	s.Focused.Placeholder = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(bg)
	s.Focused.LineNumber = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(bg)
	s.Focused.CursorLine = lipgloss.NewStyle().Background(cursorBg)
	s.Focused.CursorLineNumber = lipgloss.NewStyle().Foreground(styles.Primary).Background(cursorBg)
	s.Focused.EndOfBuffer = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(bg)
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(styles.Primary).Background(bg)

	// Blurred styles - ALL components must have theme background
	s.Blurred.Base = lipgloss.NewStyle().Background(bg)
	s.Blurred.Text = lipgloss.NewStyle().Foreground(styles.Text).Background(bg)
	s.Blurred.Placeholder = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(bg)
	s.Blurred.LineNumber = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(bg)
	s.Blurred.CursorLine = lipgloss.NewStyle().Background(bg)
	s.Blurred.CursorLineNumber = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(bg)
	s.Blurred.EndOfBuffer = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(bg)
	s.Blurred.Prompt = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(bg)

	ta.SetStyles(s)
}

func applyTableStyles(t *table.Model) {
	bg := styles.BgDefault
	// Use a much more contrasting background for selected row
	selectedBg := lipgloss.Color("240") // Lighter gray for visibility
	selectedFg := lipgloss.Color("255") // White text

	t.SetStyles(table.Styles{
		Header:   lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Background(bg).Padding(0, 1),
		Cell:     lipgloss.NewStyle().Padding(0, 1).Background(bg).Foreground(styles.Text),
		Selected: lipgloss.NewStyle().Foreground(selectedFg).Bold(true).Background(selectedBg).Padding(0, 1),
	})
}

// updateThemeListStyles updates the theme list styles to match current theme
func (m *BrowserModel) updateThemeListStyles() {
	bg := styles.BgDark

	// Update delegate styles
	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(styles.Text).
		Background(bg)
	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(bg)
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(styles.Primary).
		Background(bg).
		Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(bg)
	delegate.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(bg)
	delegate.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(bg)
	delegate.Styles.FilterMatch = lipgloss.NewStyle().
		Foreground(styles.Primary).
		Background(bg).
		Bold(true)

	// Hide help by setting empty help functions and spacing
	delegate.ShortHelpFunc = func() []key.Binding { return nil }
	delegate.FullHelpFunc = func() [][]key.Binding { return nil }
	delegate.SetSpacing(0)
	delegate.SetHeight(1)

	m.themeList.SetDelegate(delegate)

	// Update list styles
	m.themeList.Styles.TitleBar = lipgloss.NewStyle().Background(bg)
	m.themeList.Styles.Title = lipgloss.NewStyle().Background(bg)
	m.themeList.Styles.Spinner = lipgloss.NewStyle().Background(bg)
	m.themeList.Styles.Filter = textinput.Styles{
		Focused: textinput.StyleState{
			Text:        lipgloss.NewStyle().Foreground(styles.Text).Background(bg),
			Placeholder: lipgloss.NewStyle().Foreground(styles.TextMuted).Background(bg),
			Prompt:      lipgloss.NewStyle().Foreground(styles.Primary).Background(bg),
		},
		Blurred: textinput.StyleState{
			Text:        lipgloss.NewStyle().Foreground(styles.Text).Background(bg),
			Placeholder: lipgloss.NewStyle().Foreground(styles.TextMuted).Background(bg),
			Prompt:      lipgloss.NewStyle().Foreground(styles.Primary).Background(bg),
		},
	}
	m.themeList.Styles.DefaultFilterCharacterMatch = lipgloss.NewStyle().
		Foreground(styles.Primary).
		Background(bg).
		Bold(true)
	m.themeList.Styles.StatusBar = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(bg)
	m.themeList.Styles.StatusEmpty = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(bg)
	m.themeList.Styles.StatusBarActiveFilter = lipgloss.NewStyle().
		Foreground(styles.Text).
		Background(bg)
	m.themeList.Styles.StatusBarFilterCount = lipgloss.NewStyle().
		Foreground(styles.Primary).
		Background(bg)
	m.themeList.Styles.NoItems = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(bg)
	m.themeList.Styles.PaginationStyle = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(bg)
	m.themeList.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(bg)
	m.themeList.Styles.ActivePaginationDot = lipgloss.NewStyle().
		Foreground(styles.Primary).
		Background(bg)
	m.themeList.Styles.InactivePaginationDot = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(bg)
	m.themeList.Styles.ArabicPagination = lipgloss.NewStyle().
		Foreground(styles.Text).
		Background(bg)
	m.themeList.Styles.DividerDot = lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Background(bg)
}
