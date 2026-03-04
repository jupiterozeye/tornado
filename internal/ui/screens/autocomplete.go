package screens

import (
	"strings"
	"time"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

// Autocomplete suggestion types
type SuggestionType int

const (
	SuggestKeyword SuggestionType = iota
	SuggestTable
	SuggestColumn
	SuggestFunction
)

// Suggestion represents an autocomplete suggestion
type Suggestion struct {
	Text        string
	Type        SuggestionType
	Description string
}

// QueryContext represents the editing context for autocomplete
type QueryContext struct {
	Text        string
	CursorPos   int
	CurrentWord string
	PrevToken   string
	InString    bool
	InComment   bool
}

// AutocompleteModel manages autocomplete state
type AutocompleteModel struct {
	Visible     bool
	Suggestions []Suggestion
	Selected    int
	TriggerPos  int
	Width       int
	Height      int
}

// NewAutocompleteModel creates a new autocomplete model
func NewAutocompleteModel() *AutocompleteModel {
	return &AutocompleteModel{
		Visible:     false,
		Suggestions: []Suggestion{},
		Selected:    0,
		Width:       40,
		Height:      10,
	}
}

// getQueryContext analyzes the current query state
func getQueryContext(queryText string, cursorPos int) QueryContext {
	ctx := QueryContext{
		Text:      queryText,
		CursorPos: cursorPos,
	}

	if cursorPos > len(queryText) {
		cursorPos = len(queryText)
	}

	// Find current word
	start := cursorPos
	for start > 0 && !isWordBreak(queryText[start-1]) {
		start--
	}
	ctx.CurrentWord = queryText[start:cursorPos]

	// Find previous token (skipping current word)
	pos := start
	for pos > 0 && unicode.IsSpace(rune(queryText[pos-1])) {
		pos--
	}
	if pos > 0 {
		tokenStart := pos
		for tokenStart > 0 && !isWordBreak(queryText[tokenStart-1]) {
			tokenStart--
		}
		ctx.PrevToken = strings.ToUpper(queryText[tokenStart:pos])
	}

	// Check if in string
	quoteCount := 0
	for i := 0; i < cursorPos; i++ {
		if queryText[i] == '\'' {
			// Check for escaped quote
			if i == 0 || queryText[i-1] != '\'' {
				quoteCount++
			}
		}
	}
	ctx.InString = quoteCount%2 == 1

	// Check if in comment
	for i := 0; i < cursorPos-1; i++ {
		if queryText[i] == '-' && queryText[i+1] == '-' {
			// Line comment - check if we're still on this line
			for j := i + 2; j < cursorPos; j++ {
				if queryText[j] == '\n' {
					ctx.InComment = false
					break
				}
				if j == cursorPos-1 {
					ctx.InComment = true
				}
			}
		}
	}

	return ctx
}

func isWordBreak(ch byte) bool {
	return unicode.IsSpace(rune(ch)) || strings.ContainsRune(".,;()=<>!+-*/", rune(ch))
}

// getSuggestions returns autocomplete suggestions based on context
func getSuggestions(ctx QueryContext, tables []string, columns map[string][]string) []Suggestion {
	if ctx.InString || ctx.InComment {
		return nil
	}

	var suggestions []Suggestion
	prefix := strings.ToUpper(ctx.CurrentWord)

	// Context-aware suggestions
	switch ctx.PrevToken {
	case "FROM", "INTO", "UPDATE", "TABLE", "JOIN":
		// Suggest tables
		for _, table := range tables {
			if strings.HasPrefix(strings.ToUpper(table), prefix) {
				suggestions = append(suggestions, Suggestion{
					Text:        table,
					Type:        SuggestTable,
					Description: "table",
				})
			}
		}
	case "SELECT", "WHERE", "AND", "OR", "SET", "HAVING":
		// Suggest columns and keywords
		suggestions = append(suggestions, getColumnSuggestions(columns, prefix)...)
		suggestions = append(suggestions, getKeywordSuggestions(prefix)...)
	default:
		// General suggestions - keywords and functions
		suggestions = append(suggestions, getKeywordSuggestions(prefix)...)
		suggestions = append(suggestions, getFunctionSuggestions(prefix)...)
	}

	return suggestions
}

func getKeywordSuggestions(prefix string) []Suggestion {
	keywords := []string{
		"SELECT", "FROM", "WHERE", "INSERT", "UPDATE", "DELETE",
		"CREATE", "DROP", "ALTER", "TABLE", "INDEX", "VIEW",
		"JOIN", "INNER", "LEFT", "RIGHT", "OUTER", "ON",
		"GROUP", "BY", "ORDER", "HAVING", "LIMIT", "OFFSET",
		"AND", "OR", "NOT", "IN", "EXISTS", "BETWEEN",
		"LIKE", "IS", "NULL", "TRUE", "FALSE",
		"VALUES", "SET", "INTO", "AS", "DISTINCT", "ALL",
		"UNION", "INTERSECT", "EXCEPT", "WITH", "RECURSIVE",
		"CASE", "WHEN", "THEN", "ELSE", "END",
		"COMMIT", "ROLLBACK", "BEGIN", "TRANSACTION",
	}

	var suggestions []Suggestion
	for _, kw := range keywords {
		if strings.HasPrefix(kw, prefix) {
			suggestions = append(suggestions, Suggestion{
				Text:        kw,
				Type:        SuggestKeyword,
				Description: "keyword",
			})
		}
	}
	return suggestions
}

func getFunctionSuggestions(prefix string) []Suggestion {
	functions := []string{
		"COUNT", "SUM", "AVG", "MIN", "MAX",
		"LENGTH", "SUBSTR", "TRIM", "UPPER", "LOWER",
		"CONCAT", "REPLACE", "INSTR", "ABS", "ROUND",
		"COALESCE", "IFNULL", "NULLIF", "DATE", "TIME",
		"DATETIME", "STRFTIME", "RANDOM", "CHANGES",
		"ROW_NUMBER", "RANK", "DENSE_RANK",
		"LAG", "LEAD", "FIRST_VALUE", "LAST_VALUE",
	}

	var suggestions []Suggestion
	for _, fn := range functions {
		if strings.HasPrefix(fn, prefix) {
			suggestions = append(suggestions, Suggestion{
				Text:        fn,
				Type:        SuggestFunction,
				Description: "function",
			})
		}
	}
	return suggestions
}

func getColumnSuggestions(columns map[string][]string, prefix string) []Suggestion {
	var suggestions []Suggestion
	for table, cols := range columns {
		for _, col := range cols {
			if strings.HasPrefix(strings.ToUpper(col), prefix) {
				suggestions = append(suggestions, Suggestion{
					Text:        col,
					Type:        SuggestColumn,
					Description: table + " column",
				})
			}
		}
	}
	return suggestions
}

// Render renders the autocomplete dropdown
func (m *AutocompleteModel) Render() string {
	if !m.Visible || len(m.Suggestions) == 0 {
		return ""
	}

	// Limit displayed suggestions
	maxDisplay := m.Height - 2
	if maxDisplay < 3 {
		maxDisplay = 3
	}

	start := 0
	if m.Selected >= maxDisplay {
		start = m.Selected - maxDisplay + 1
	}
	end := start + maxDisplay
	if end > len(m.Suggestions) {
		end = len(m.Suggestions)
	}

	// Type color mapping
	typeStyles := map[SuggestionType]lipgloss.Style{
		SuggestKeyword:  lipgloss.NewStyle().Foreground(styles.Primary),
		SuggestTable:    lipgloss.NewStyle().Foreground(styles.Success),
		SuggestColumn:   lipgloss.NewStyle().Foreground(styles.Warning),
		SuggestFunction: lipgloss.NewStyle().Foreground(styles.Accent),
	}
	defaultStyle := lipgloss.NewStyle().Foreground(styles.Text)

	// Build content with explicit background on every element
	var lines []string
	for i := start; i < end; i++ {
		sugg := m.Suggestions[i]
		var line string
		if i == m.Selected {
			line = lipgloss.NewStyle().
				Background(styles.Primary).
				Foreground(styles.BgDark).
				Render(" " + sugg.Text + " ")
		} else {
			style, ok := typeStyles[sugg.Type]
			if !ok {
				style = defaultStyle
			}
			line = style.Background(styles.BgDark).Render(sugg.Text)
		}
		lines = append(lines, line)
	}

	// Add count indicator if truncated
	if len(m.Suggestions) > maxDisplay {
		countLine := lipgloss.NewStyle().
			Foreground(styles.TextMuted).
			Background(styles.BgDark).
			Render("... " + string(rune('0'+len(m.Suggestions)-maxDisplay)) + " more")
		lines = append(lines, countLine)
	}

	// Pad content to ensure minimum height for consistent rendering
	innerWidth := m.Width - 6 // Account for borders and padding
	minContentHeight := maxDisplay
	for len(lines) < minContentHeight {
		lines = append(lines, lipgloss.NewStyle().Background(styles.BgDark).Render(strings.Repeat(" ", innerWidth)))
	}

	content := strings.Join(lines, "\n")

	// Apply border with background - the key is applying background to the WHOLE style including border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderFocus).
		BorderBackground(styles.BgDark).
		Background(styles.BgDark).
		Padding(0, 1)

	return boxStyle.Render(content)
}

// HandleKey handles key presses when autocomplete is active
func (m *AutocompleteModel) HandleKey(msg tea.KeyPressMsg) (bool, string) {
	if !m.Visible {
		return false, ""
	}

	switch msg.String() {
	case "up":
		if m.Selected > 0 {
			m.Selected--
		} else {
			m.Selected = len(m.Suggestions) - 1
		}
		return true, ""
	case "down":
		if m.Selected < len(m.Suggestions)-1 {
			m.Selected++
		} else {
			m.Selected = 0
		}
		return true, ""
	case "enter", "tab":
		if len(m.Suggestions) > 0 && m.Selected < len(m.Suggestions) {
			return true, m.Suggestions[m.Selected].Text
		}
		return false, ""
	case "esc":
		m.Visible = false
		return true, ""
	default:
		return false, ""
	}
}

// TriggerAutocompleteMsg triggers autocomplete after a delay
type TriggerAutocompleteMsg struct {
	QueryText string
	CursorPos int
}

// TriggerAutocomplete creates a command that triggers autocomplete after debounce
func TriggerAutocomplete(queryText string, cursorPos int) tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TriggerAutocompleteMsg{
			QueryText: queryText,
			CursorPos: cursorPos,
		}
	})
}
