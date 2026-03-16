package screens

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

// renderDialogBox builds a bordered dialog with title in the top border
// and shortcuts in the bottom border. Body entries may be multi-line.
func renderDialogBox(title string, body []string, subtitle string, width int) string {
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

	out := make([]string, 0, len(body)*3+2)
	out = append(out, borderStyle.Render("╭"+makeDialogTopBorder(title, innerWidth)+"╮"))

	for _, entry := range body {
		subLines := strings.Split(entry, "\n")
		for _, line := range subLines {
			lineWidth := lipgloss.Width(line)
			// Pad line to fill inner width with spaces that have proper background
			if lineWidth < innerWidth {
				padding := lipgloss.NewStyle().Background(bg).Render(strings.Repeat(" ", innerWidth-lineWidth))
				line = line + padding
			}
			line = truncateString(line, innerWidth)
			// Render with background
			renderedLine := bodyStyle.Render(line)
			out = append(out, borderStyle.Render("│")+renderedLine+borderStyle.Render("│"))
		}
	}

	out = append(out, borderStyle.Render("╰"+makeDialogBottomBorder(subtitle, innerWidth)+"╯"))
	return strings.Join(out, "\n")
}

func makeDialogTopBorder(label string, width int) string {
	if width < 1 {
		return ""
	}
	segment := "─ " + label + " "
	segment = truncateString(segment, width)
	if lipgloss.Width(segment) < width {
		segment += strings.Repeat("─", width-lipgloss.Width(segment))
	}
	return segment
}

// truncateString truncates a string to fit within width without adding ellipsis
func truncateString(s string, width int) string {
	if width < 1 {
		return ""
	}
	w := lipgloss.Width(s)
	if w <= width {
		return s
	}
	// Simple truncation without ellipsis
	runes := []rune(s)
	result := ""
	for _, r := range runes {
		if lipgloss.Width(result+string(r)) > width {
			break
		}
		result += string(r)
	}
	return result
}

func makeDialogBottomBorder(label string, width int) string {
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

