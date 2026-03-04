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
	"fmt"
	"image/color"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	statusMsg     string
	maximizedPane Pane

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
	themeList := list.New(themeItems, list.NewDefaultDelegate(), 36, 12)
	themeList.Title = "Themes"
	themeList.SetShowStatusBar(false)
	themeList.SetFilteringEnabled(false)

	return &BrowserModel{
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
		usableHeight := msg.Height - 1 // reserve one line for contextual footer
		if usableHeight < 3 {
			usableHeight = 3
		}
		m.layoutManager.Update(msg.Width, usableHeight)
		m.updateComponentSizes()

	case tea.KeyPressMsg:
		if m.themeMenu {
			return m.handleThemeMenuKey(msg)
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

		// Global key bindings
		switch msg.String() {
		case "space":
			if m.focusedPane == PaneQuery && m.queryMode == QueryModeInsert {
				return m.routeKeyMsg(msg)
			}
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

	// Render query editor
	queryTitle := "Query [NORMAL]"
	if m.queryMode == QueryModeInsert {
		queryTitle = "Query [INSERT]"
	}
	queryPane := m.renderPane(queryTitle, "q", m.query.View(), qw, qh, m.focusedPane == PaneQuery, styles.BgDark)

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
			main = m.renderPane(queryTitle, "q", m.query.View(), m.width, m.mainHeight(), m.focusedPane == PaneQuery, styles.BgDark)
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

	return view
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
	for len(bodyLines) < bodyHeight {
		bodyLines = append(bodyLines, "")
	}

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
		if m.queryMode == QueryModeInsert {
			text = "Query INSERT: Esc Normal  Enter Newline"
		} else {
			text = "Query NORMAL: i Insert  Enter Execute  e Explorer  r Results"
		}
	case PaneResults:
		text = "Results: j/k Move  q Query  e Explorer"
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

	// Dismiss menu and execute the command
	m.leaderActive = false
	return m.executeLeaderCommand(msg.String())
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
		if m.db != nil {
			_ = m.db.Disconnect()
		}
		return m, func() tea.Msg { return RequestConnectMsg{} }
	case "t":
		m.themeMenu = true
		m.leaderActive = false
		m.statusMsg = "Select theme and press Enter"
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

func (m *BrowserModel) handleThemeMenuKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.themeMenu = false
		m.statusMsg = ""
		return m, nil
	case "enter":
		if it, ok := m.themeList.SelectedItem().(browserThemeItem); ok {
			if styles.SetTheme(it.name) {
				m.styles = styles.Default()
				applyTextAreaStyles(&m.query)
				applyTableStyles(&m.results)
				m.statusMsg = "Theme: " + it.name
				// Save theme preference (async)
				if cfg := config.Get(); cfg != nil {
					go cfg.SetTheme(it.name)
				}
			}
		}
		m.themeMenu = false
		return m, nil
	}

	var cmd tea.Cmd
	m.themeList, cmd = m.themeList.Update(msg)
	return m, cmd
}

func (m *BrowserModel) handleExplorerActionKey(msg tea.KeyPressMsg) (bool, tea.Cmd) {
	node := (*components.TreeNode)(nil)
	if m.explorer != nil {
		node = m.explorer.CurrentNode()
	}

	switch msg.String() {
	case "x":
		if m.db != nil {
			_ = m.db.Disconnect()
		}
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
	lines := strings.Split(m.themeList.View(), "\n")
	menu := renderDialogBox("Themes", lines, "enter Select · esc Cancel", 44)

	// Center position
	boxH := len(strings.Split(menu, "\n"))
	boxW := 44
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

func buildLeaderMenuLines() []string {
	keyStyle := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true)
	headStyle := lipgloss.NewStyle().Foreground(styles.Secondary).Bold(true)
	return []string{
		headStyle.Render("Navigation"),
		"  " + keyStyle.Render("e") + "  Toggle Explorer",
		"  " + keyStyle.Render("f") + "  Toggle Maximize",
		"",
		headStyle.Render("Connection"),
		"  " + keyStyle.Render("c") + "  Connect",
		"  " + keyStyle.Render("x") + "  Disconnect",
		"",
		headStyle.Render("Other"),
		"  " + keyStyle.Render("t") + "  Change Theme",
		"  " + keyStyle.Render("h") + "  Help",
		"  " + keyStyle.Render("/") + "  Search",
		"  " + keyStyle.Render("q") + "  Quit",
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
	borderStyle := lipgloss.NewStyle().
		Foreground(styles.BorderFocus).
		Background(styles.BgDark)
	bodyStyle := lipgloss.NewStyle().
		Background(styles.BgDark).
		Width(innerWidth)

	out := make([]string, 0, len(body)+2)
	out = append(out, borderStyle.Render("╭"+makeMenuTopBorder(title, innerWidth)+"╮"))

	for _, line := range body {
		line = truncateToWidth(line, innerWidth)
		out = append(out, borderStyle.Render("│")+bodyStyle.Render(line)+borderStyle.Render("│"))
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
	if width == 1 {
		return xansi.Truncate(s, width, "")
	}
	return xansi.Truncate(s, width, "…")
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
	ew, eh, qw, qh, rw, rh := m.paneDimensions()

	// Update explorer size
	if m.explorer != nil && m.showExplorer {
		m.explorer.SetSize(ew, eh)
	}

	// Update query editor size
	m.query.SetWidth(maxInt(10, qw-6))
	m.query.SetHeight(maxInt(3, qh-6))

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
		// Save query to history (async)
		if cfg := config.Get(); cfg != nil {
			go cfg.AddQuery(query)
		}
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

	// Clear rows first to prevent index mismatch when columns change
	m.results.SetRows([]table.Row{})
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

func applyTextAreaStyles(ta *textarea.Model) {
	s := ta.Styles()
	bgStyle := lipgloss.NewStyle().Background(styles.BgDark)

	s.Focused.Base = bgStyle
	s.Focused.Text = lipgloss.NewStyle().Foreground(styles.Text).Background(styles.BgDark)
	s.Focused.Placeholder = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(styles.BgDark)
	s.Focused.LineNumber = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(styles.BgDark)
	s.Focused.CursorLine = lipgloss.NewStyle().Background(styles.BgLight)
	s.Focused.CursorLineNumber = lipgloss.NewStyle().Foreground(styles.Primary).Background(styles.BgLight)
	s.Focused.EndOfBuffer = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(styles.BgDark)
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(styles.Primary).Background(styles.BgDark)

	s.Blurred.Base = bgStyle
	s.Blurred.Text = lipgloss.NewStyle().Foreground(styles.Text).Background(styles.BgDark)
	s.Blurred.Placeholder = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(styles.BgDark)
	s.Blurred.LineNumber = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(styles.BgDark)
	s.Blurred.CursorLine = lipgloss.NewStyle().Background(styles.BgDark)
	s.Blurred.CursorLineNumber = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(styles.BgDark)
	s.Blurred.EndOfBuffer = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(styles.BgDark)
	s.Blurred.Prompt = lipgloss.NewStyle().Foreground(styles.TextMuted).Background(styles.BgDark)

	ta.SetStyles(s)
}

func applyTableStyles(t *table.Model) {
	t.SetStyles(table.Styles{
		Header:   styles.TableHeader().Background(styles.BgDefault),
		Cell:     lipgloss.NewStyle().Padding(0, 1).Background(styles.BgDefault),
		Selected: lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Background(styles.BgLight),
	})
}
