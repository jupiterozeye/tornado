// Package components provides UI components for the Tornado application.
//
// This file implements the Explorer component - a tree browser for database objects
// similar to lazygit's file explorer. It displays:
//   - Tables (expandable to show columns)
//   - Views
//   - Indexes
//   - Triggers
//   - Sequences
//
// Key bindings:
//   - j/k: Navigate up/down
//   - h: Collapse node or go to parent
//   - l/Enter: Expand node or go to children
//   - s: Select table (SELECT TOP 100)
//   - r: Refresh tree
package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jupiterozeye/tornado/internal/db"
	"github.com/jupiterozeye/tornado/internal/models"
	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

// NodeType represents the type of a tree node
type NodeType int

const (
	NodeRoot NodeType = iota
	NodeCategory
	NodeTable
	NodeView
	NodeIndex
	NodeTrigger
	NodeSequence
	NodeColumn
)

// TreeNode represents a node in the explorer tree
type TreeNode struct {
	Name       string
	Type       NodeType
	Expanded   bool
	Children   []*TreeNode
	Parent     *TreeNode
	TableName  string // For columns, which table they belong to
	ColumnInfo *models.Column
}

// ExplorerModel is the model for the database explorer component
type ExplorerModel struct {
	db       db.Database
	root     *TreeNode
	cursor   int // Current position in flattened tree
	flatList []*TreeNode
	width    int
	height   int
	styles   *styles.Styles
	focused  bool
}

// NewExplorerModel creates a new explorer component
func NewExplorerModel(database db.Database, width, height int) *ExplorerModel {
	s := styles.Default()

	m := &ExplorerModel{
		db:      database,
		width:   width,
		height:  height,
		styles:  s,
		focused: false,
		root: &TreeNode{
			Name:     "Database",
			Type:     NodeRoot,
			Expanded: true,
			Children: []*TreeNode{
				{Name: "Tables", Type: NodeCategory, Expanded: false},
				{Name: "Views", Type: NodeCategory, Expanded: false},
				{Name: "Indexes", Type: NodeCategory, Expanded: false},
				{Name: "Triggers", Type: NodeCategory, Expanded: false},
				{Name: "Sequences", Type: NodeCategory, Expanded: false},
			},
		},
	}

	// Set up parent pointers
	for _, child := range m.root.Children {
		child.Parent = m.root
	}

	m.flattenTree()
	return m
}

// Init returns the initial command
func (m *ExplorerModel) Init() tea.Cmd {
	return m.loadDatabaseObjects()
}

// Update handles messages
func (m *ExplorerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.moveDown()
		case "k", "up":
			m.moveUp()
		case "h":
			m.collapseOrGoToParent()
		case "l", "enter":
			return m, m.expandOrLoad()
		case "s":
			return m, m.selectCurrent()
		case "r":
			return m, m.loadDatabaseObjects()
		}

	case TablesLoadedMsg:
		m.updateTables(msg.Tables)

	case ViewsLoadedMsg:
		m.updateViews(msg.Views)

	case IndexesLoadedMsg:
		m.updateIndexes(msg.Indexes)

	case TriggersLoadedMsg:
		m.updateTriggers(msg.Triggers)

	case SequencesLoadedMsg:
		m.updateSequences(msg.Sequences)

	case ColumnsLoadedMsg:
		m.updateColumns(msg.TableName, msg.Columns)
	}

	return m, nil
}

// View renders the explorer
func (m *ExplorerModel) View() string {
	if len(m.flatList) == 0 {
		return "Loading..."
	}

	var lines []string
	visibleCount := 0

	for i, node := range m.flatList {
		// Only show visible nodes (respect parent expanded state)
		if !m.isNodeVisible(node) {
			continue
		}

		// Skip if scrolled out of view
		if visibleCount < m.getScrollOffset() {
			visibleCount++
			continue
		}

		// Stop if beyond visible area
		if visibleCount >= m.getScrollOffset()+m.height-2 {
			break
		}

		line := m.renderNode(node, i == m.cursor)
		lines = append(lines, line)
		visibleCount++
	}

	return strings.Join(lines, "\n")
}

// Helper methods

func (m *ExplorerModel) flattenTree() {
	m.flatList = nil
	m.flattenNode(m.root, 0)
}

func (m *ExplorerModel) flattenNode(node *TreeNode, depth int) {
	m.flatList = append(m.flatList, node)
	if node.Expanded {
		for _, child := range node.Children {
			m.flattenNode(child, depth+1)
		}
	}
}

func (m *ExplorerModel) isNodeVisible(node *TreeNode) bool {
	if node.Parent == nil {
		return true
	}
	return node.Parent.Expanded && m.isNodeVisible(node.Parent)
}

func (m *ExplorerModel) renderNode(node *TreeNode, selected bool) string {
	var prefix string
	indent := m.getIndent(node)

	// Determine prefix based on node state
	switch {
	case len(node.Children) > 0 || node.Type == NodeCategory || node.Type == NodeTable:
		if node.Expanded {
			prefix = "▼ "
		} else {
			prefix = "▶ "
		}
	default:
		prefix = "  "
	}

	// Format node name
	name := node.Name
	switch node.Type {
	case NodeColumn:
		if node.ColumnInfo != nil {
			name = fmt.Sprintf("%s (%s)", node.Name, node.ColumnInfo.Type)
		}
	}

	line := indent + prefix + name
	maxWidth := m.width - 4
	if maxWidth < 1 {
		maxWidth = 1
	}
	line = truncateLine(line, maxWidth)

	// Apply selection styling
	if selected {
		return lipgloss.NewStyle().Bold(true).Foreground(styles.TextBold).Render(line)
	}

	// Apply type-specific styling
	switch node.Type {
	case NodeCategory:
		return lipgloss.NewStyle().Bold(true).Foreground(styles.Secondary).Render(line)
	case NodeTable:
		return lipgloss.NewStyle().Foreground(styles.Primary).Render(line)
	default:
		return lipgloss.NewStyle().Foreground(styles.Text).Render(line)
	}
}

func truncateLine(s string, width int) string {
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

func (m *ExplorerModel) getIndent(node *TreeNode) string {
	depth := 0
	parent := node.Parent
	for parent != nil && parent != m.root {
		depth++
		parent = parent.Parent
	}
	return strings.Repeat("  ", depth)
}

func (m *ExplorerModel) getScrollOffset() int {
	// Keep cursor centered when possible
	visibleHeight := m.height - 2
	halfHeight := visibleHeight / 2

	if m.cursor < halfHeight {
		return 0
	}

	return m.cursor - halfHeight
}

func (m *ExplorerModel) moveDown() {
	if m.cursor < len(m.flatList)-1 {
		m.cursor++
	}
}

func (m *ExplorerModel) moveUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

func (m *ExplorerModel) collapseOrGoToParent() {
	if m.cursor >= len(m.flatList) {
		return
	}

	node := m.flatList[m.cursor]
	if node.Expanded && len(node.Children) > 0 {
		node.Expanded = false
		m.flattenTree()
	} else if node.Parent != nil && node.Parent != m.root {
		// Find parent in flat list and move cursor there
		for i, n := range m.flatList {
			if n == node.Parent {
				m.cursor = i
				break
			}
		}
	}
}

func (m *ExplorerModel) expandOrLoad() tea.Cmd {
	if m.cursor >= len(m.flatList) {
		return nil
	}

	node := m.flatList[m.cursor]

	// Toggle collapse if already expanded.
	if node.Expanded {
		node.Expanded = false
		m.flattenTree()
		return nil
	}

	node.Expanded = true

	// Load children if needed
	switch node.Type {
	case NodeCategory:
		switch node.Name {
		case "Tables":
			return m.loadTables()
		case "Views":
			return m.loadViews()
		case "Triggers":
			return m.loadTriggers()
		case "Sequences":
			return m.loadSequences()
		}
	case NodeTable:
		return m.loadColumns(node.Name)
	}

	m.flattenTree()

	return nil
}

func (m *ExplorerModel) selectCurrent() tea.Cmd {
	if m.cursor >= len(m.flatList) {
		return nil
	}

	node := m.flatList[m.cursor]
	if node.Type == NodeTable {
		return func() tea.Msg {
			return TableSelectedMsg{Name: node.Name}
		}
	}

	return nil
}

func (m *ExplorerModel) SetFocused(focused bool) {
	m.focused = focused
}

func (m *ExplorerModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// CurrentNode returns the currently highlighted node.
func (m *ExplorerModel) CurrentNode() *TreeNode {
	if m.cursor < 0 || m.cursor >= len(m.flatList) {
		return nil
	}
	return m.flatList[m.cursor]
}

// Async loading commands

func (m *ExplorerModel) loadDatabaseObjects() tea.Cmd {
	return func() tea.Msg {
		tables, err := m.db.ListTables()
		if err != nil {
			return TablesLoadedMsg{Err: err}
		}
		return TablesLoadedMsg{Tables: tables}
	}
}

func (m *ExplorerModel) loadTables() tea.Cmd {
	return func() tea.Msg {
		tables, err := m.db.ListTables()
		if err != nil {
			return TablesLoadedMsg{Err: err}
		}
		return TablesLoadedMsg{Tables: tables}
	}
}

func (m *ExplorerModel) loadViews() tea.Cmd {
	return func() tea.Msg {
		views, err := m.db.ListViews()
		if err != nil {
			return ViewsLoadedMsg{Err: err}
		}
		return ViewsLoadedMsg{Views: views}
	}
}

func (m *ExplorerModel) loadTriggers() tea.Cmd {
	return func() tea.Msg {
		triggers, err := m.db.ListTriggers()
		if err != nil {
			return TriggersLoadedMsg{Err: err}
		}
		return TriggersLoadedMsg{Triggers: triggers}
	}
}

func (m *ExplorerModel) loadSequences() tea.Cmd {
	return func() tea.Msg {
		sequences, err := m.db.ListSequences()
		if err != nil {
			return SequencesLoadedMsg{Err: err}
		}
		return SequencesLoadedMsg{Sequences: sequences}
	}
}

func (m *ExplorerModel) loadColumns(tableName string) tea.Cmd {
	return func() tea.Msg {
		schema, err := m.db.DescribeTable(tableName)
		if err != nil {
			return ColumnsLoadedMsg{TableName: tableName, Err: err}
		}
		return ColumnsLoadedMsg{TableName: tableName, Columns: schema.Columns}
	}
}

// Update methods

func (m *ExplorerModel) updateTables(tables []string) {
	for _, category := range m.root.Children {
		if category.Name == "Tables" {
			category.Children = nil
			for _, tableName := range tables {
				category.Children = append(category.Children, &TreeNode{
					Name:     tableName,
					Type:     NodeTable,
					Expanded: false,
					Parent:   category,
				})
			}
			category.Expanded = true
			break
		}
	}
	m.flattenTree()
}

func (m *ExplorerModel) updateViews(views []string) {
	for _, category := range m.root.Children {
		if category.Name == "Views" {
			category.Children = nil
			for _, viewName := range views {
				category.Children = append(category.Children, &TreeNode{
					Name:   viewName,
					Type:   NodeView,
					Parent: category,
				})
			}
			category.Expanded = true
			break
		}
	}
	m.flattenTree()
}

func (m *ExplorerModel) updateIndexes(indexes []string) {
	// Indexes are loaded per-table, not globally
	m.flattenTree()
}

func (m *ExplorerModel) updateTriggers(triggers []string) {
	for _, category := range m.root.Children {
		if category.Name == "Triggers" {
			category.Children = nil
			for _, triggerName := range triggers {
				category.Children = append(category.Children, &TreeNode{
					Name:   triggerName,
					Type:   NodeTrigger,
					Parent: category,
				})
			}
			category.Expanded = true
			break
		}
	}
	m.flattenTree()
}

func (m *ExplorerModel) updateSequences(sequences []string) {
	for _, category := range m.root.Children {
		if category.Name == "Sequences" {
			category.Children = nil
			for _, seqName := range sequences {
				category.Children = append(category.Children, &TreeNode{
					Name:   seqName,
					Type:   NodeSequence,
					Parent: category,
				})
			}
			category.Expanded = true
			break
		}
	}
	m.flattenTree()
}

func (m *ExplorerModel) updateColumns(tableName string, columns []models.Column) {
	for _, category := range m.root.Children {
		if category.Name == "Tables" {
			for _, table := range category.Children {
				if table.Name == tableName {
					table.Children = nil
					for _, col := range columns {
						c := col
						table.Children = append(table.Children, &TreeNode{
							Name:       c.Name,
							Type:       NodeColumn,
							Parent:     table,
							TableName:  tableName,
							ColumnInfo: &c,
						})
					}
					table.Expanded = true
					break
				}
			}
			break
		}
	}
	m.flattenTree()
}

// Message types

type TablesLoadedMsg struct {
	Tables []string
	Err    error
}

type ViewsLoadedMsg struct {
	Views []string
	Err   error
}

type IndexesLoadedMsg struct {
	TableName string
	Indexes   []string
	Err       error
}

type TriggersLoadedMsg struct {
	Triggers []string
	Err      error
}

type SequencesLoadedMsg struct {
	Sequences []string
	Err       error
}

type ColumnsLoadedMsg struct {
	TableName string
	Columns   []models.Column
	Err       error
}
