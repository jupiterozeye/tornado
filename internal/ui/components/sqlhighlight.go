package components

import (
	"strings"
	"unicode"

	"charm.land/lipgloss/v2"

	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

// SQL token types for syntax highlighting.
type sqlTokenType int

const (
	tokenIdentifier sqlTokenType = iota
	tokenKeyword
	tokenFunction
	tokenString
	tokenNumber
	tokenComment
	tokenOperator
)

// sqlKeywords is the set of SQL keywords to highlight.
var sqlKeywords = map[string]bool{
	"SELECT": true, "FROM": true, "WHERE": true, "AND": true, "OR": true,
	"NOT": true, "IN": true, "IS": true, "NULL": true, "AS": true,
	"ON": true, "JOIN": true, "LEFT": true, "RIGHT": true, "INNER": true,
	"OUTER": true, "CROSS": true, "FULL": true, "NATURAL": true,
	"INSERT": true, "INTO": true, "VALUES": true, "UPDATE": true, "SET": true,
	"DELETE": true, "CREATE": true, "ALTER": true, "DROP": true, "TABLE": true,
	"INDEX": true, "VIEW": true, "TRIGGER": true, "DATABASE": true,
	"IF": true, "EXISTS": true, "BEGIN": true, "END": true, "COMMIT": true,
	"ROLLBACK": true, "TRANSACTION": true, "SAVEPOINT": true,
	"ORDER": true, "BY": true, "GROUP": true, "HAVING": true, "LIMIT": true,
	"OFFSET": true, "UNION": true, "ALL": true, "DISTINCT": true, "TOP": true,
	"ASC": true, "DESC": true, "BETWEEN": true, "LIKE": true, "CASE": true,
	"WHEN": true, "THEN": true, "ELSE": true, "WITH": true, "RECURSIVE": true,
	"PRIMARY": true, "KEY": true, "FOREIGN": true, "REFERENCES": true,
	"CONSTRAINT": true, "UNIQUE": true, "CHECK": true, "DEFAULT": true,
	"INTEGER": true, "TEXT": true, "REAL": true, "BLOB": true, "VARCHAR": true,
	"CHAR": true, "BOOLEAN": true, "DATE": true, "TIMESTAMP": true, "FLOAT": true,
	"DOUBLE": true, "DECIMAL": true, "NUMERIC": true, "BIGINT": true, "SMALLINT": true,
	"TRUE": true, "FALSE": true, "EXPLAIN": true, "ANALYZE": true, "PRAGMA": true,
	"REPLACE": true, "EXCEPT": true, "INTERSECT": true, "WINDOW": true, "OVER": true,
	"PARTITION": true, "ROWS": true, "RANGE": true, "UNBOUNDED": true, "PRECEDING": true,
	"FOLLOWING": true, "CURRENT": true, "ROW": true,
}

// sqlFunctions is the set of SQL functions to highlight.
var sqlFunctions = map[string]bool{
	"COUNT": true, "SUM": true, "AVG": true, "MIN": true, "MAX": true,
	"COALESCE": true, "NULLIF": true, "CAST": true, "IFNULL": true,
	"LENGTH": true, "LOWER": true, "UPPER": true, "TRIM": true, "SUBSTR": true,
	"SUBSTRING": true, "REPLACE": true, "INSTR": true, "PRINTF": true,
	"ABS": true, "ROUND": true, "RANDOM": true, "TYPEOF": true,
	"DATE": true, "TIME": true, "DATETIME": true, "STRFTIME": true, "JULIANDAY": true,
	"GROUP_CONCAT": true, "TOTAL": true, "HEX": true, "QUOTE": true, "ZEROBLOB": true,
	"ROW_NUMBER": true, "RANK": true, "DENSE_RANK": true, "NTILE": true,
	"LAG": true, "LEAD": true, "FIRST_VALUE": true, "LAST_VALUE": true, "NTH_VALUE": true,
	"CONCAT": true, "LEFT": true, "RIGHT": true, "LPAD": true, "RPAD": true,
	"NOW": true, "CURRENT_TIMESTAMP": true, "CURRENT_DATE": true, "CURRENT_TIME": true,
	"EXTRACT": true, "FLOOR": true, "CEIL": true, "CEILING": true, "MOD": true,
	"POWER": true, "SQRT": true, "LOG": true, "LN": true, "EXP": true,
	"SIGN": true, "GREATEST": true, "LEAST": true,
}

type sqlToken struct {
	Text string
	Type sqlTokenType
}

// tokenizeSQL splits a single line of SQL into tokens for highlighting.
// It handles mid-line positions for block comments that span lines.
func tokenizeSQL(line string, inBlockComment bool) ([]sqlToken, bool) {
	var tokens []sqlToken
	i := 0
	runes := []rune(line)
	n := len(runes)

	for i < n {
		// If we're inside a block comment, scan for */
		if inBlockComment {
			start := i
			for i < n-1 {
				if runes[i] == '*' && runes[i+1] == '/' {
					i += 2
					inBlockComment = false
					break
				}
				i++
			}
			if inBlockComment {
				// Reached end of line still in comment
				i = n
			}
			tokens = append(tokens, sqlToken{string(runes[start:i]), tokenComment})
			continue
		}

		ch := runes[i]

		// Line comment: -- to end of line
		if ch == '-' && i+1 < n && runes[i+1] == '-' {
			tokens = append(tokens, sqlToken{string(runes[i:]), tokenComment})
			i = n
			continue
		}

		// Block comment start: /*
		if ch == '/' && i+1 < n && runes[i+1] == '*' {
			start := i
			i += 2
			inBlockComment = true
			for i < n-1 {
				if runes[i] == '*' && runes[i+1] == '/' {
					i += 2
					inBlockComment = false
					break
				}
				i++
			}
			if inBlockComment {
				i = n
			}
			tokens = append(tokens, sqlToken{string(runes[start:i]), tokenComment})
			continue
		}

		// Single-quoted string
		if ch == '\'' {
			start := i
			i++
			for i < n {
				if runes[i] == '\'' {
					if i+1 < n && runes[i+1] == '\'' {
						i += 2 // escaped quote
						continue
					}
					i++
					break
				}
				i++
			}
			tokens = append(tokens, sqlToken{string(runes[start:i]), tokenString})
			continue
		}

		// Numbers
		if unicode.IsDigit(ch) || (ch == '.' && i+1 < n && unicode.IsDigit(runes[i+1])) {
			start := i
			for i < n && (unicode.IsDigit(runes[i]) || runes[i] == '.') {
				i++
			}
			tokens = append(tokens, sqlToken{string(runes[start:i]), tokenNumber})
			continue
		}

		// Identifiers and keywords
		if unicode.IsLetter(ch) || ch == '_' {
			start := i
			for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
				i++
			}
			word := string(runes[start:i])
			upper := strings.ToUpper(word)

			// Check if followed by '(' to detect function calls
			isFunc := false
			for j := i; j < n; j++ {
				if runes[j] == '(' {
					isFunc = true
					break
				} else if runes[j] != ' ' && runes[j] != '\t' {
					break
				}
			}

			if isFunc && sqlFunctions[upper] {
				tokens = append(tokens, sqlToken{word, tokenFunction})
			} else if sqlKeywords[upper] {
				tokens = append(tokens, sqlToken{word, tokenKeyword})
			} else {
				tokens = append(tokens, sqlToken{word, tokenIdentifier})
			}
			continue
		}

		// Whitespace: group consecutive whitespace
		if unicode.IsSpace(ch) {
			start := i
			for i < n && unicode.IsSpace(runes[i]) {
				i++
			}
			tokens = append(tokens, sqlToken{string(runes[start:i]), tokenIdentifier})
			continue
		}

		// Operators and punctuation
		tokens = append(tokens, sqlToken{string(ch), tokenOperator})
		i++
	}

	return tokens, inBlockComment
}

// styleForToken returns the lipgloss style for a given token type.
func styleForToken(tt sqlTokenType) lipgloss.Style {
	switch tt {
	case tokenKeyword:
		return lipgloss.NewStyle().Foreground(styles.Primary).Bold(true)
	case tokenFunction:
		return lipgloss.NewStyle().Foreground(styles.Info)
	case tokenString:
		return lipgloss.NewStyle().Foreground(styles.Success)
	case tokenNumber:
		return lipgloss.NewStyle().Foreground(styles.Warning)
	case tokenComment:
		return lipgloss.NewStyle().Foreground(styles.TextMuted).Italic(true)
	case tokenOperator:
		return lipgloss.NewStyle().Foreground(styles.Secondary)
	default:
		return lipgloss.NewStyle().Foreground(styles.Text)
	}
}

// HighlightSQL highlights a single line of SQL, returning a styled string.
// inBlockComment indicates whether we are inside a /* */ block comment from a
// previous line. It returns the rendered string and whether a block comment is
// still open at the end of the line.
func HighlightSQL(line string, inBlockComment bool) (string, bool) {
	tokens, stillInComment := tokenizeSQL(line, inBlockComment)
	var b strings.Builder
	for _, tok := range tokens {
		b.WriteString(styleForToken(tok.Type).Render(tok.Text))
	}
	return b.String(), stillInComment
}
