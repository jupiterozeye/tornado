package screens

import (
	"fmt"
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
			line = truncateString(line, innerWidth)
			// Ensure line has background by wrapping it
			lineWithBg := lipgloss.NewStyle().Background(bg).Render(line)
			out = append(out, borderStyle.Render("│")+bodyStyle.Render(lineWithBg)+borderStyle.Render("│"))
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

// overlayBoxAt uses ANSI cursor positioning to paint box lines on top of base.
func overlayBoxAt(base, box string, x, y, termHeight int) string {
	boxLines := strings.Split(box, "\n")
	out := base
	for i, line := range boxLines {
		out += fmt.Sprintf("\x1b[%d;%dH%s", y+i+1, x+1, line)
	}
	out += fmt.Sprintf("\x1b[%d;%dH", termHeight, 1)
	return out
}

// compositeOverlay places the box on top of the base by replacing lines directly.
// This is more robust than ANSI escape positioning for alt-screen content.
func compositeOverlay(base, box string, baseWidth, baseHeight int) string {
	baseLines := strings.Split(base, "\n")
	// Pad base to full height
	for len(baseLines) < baseHeight {
		baseLines = append(baseLines, strings.Repeat(" ", baseWidth))
	}

	boxLines := strings.Split(box, "\n")
	boxH := len(boxLines)
	boxW := 0
	for _, line := range boxLines {
		if w := lipgloss.Width(line); w > boxW {
			boxW = w
		}
	}

	// Center the box
	startY := (baseHeight - boxH) / 2
	startX := (baseWidth - boxW) / 2
	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}

	for i, boxLine := range boxLines {
		row := startY + i
		if row >= len(baseLines) {
			break
		}
		baseLine := baseLines[row]
		baseLines[row] = spliceLineAt(baseLine, boxLine, startX, baseWidth)
	}

	return strings.Join(baseLines, "\n")
}

// spliceLineAt replaces characters in baseLine starting at column x with overlay content.
// Uses visual width accounting for ANSI sequences.
func spliceLineAt(baseLine, overlay string, x, totalWidth int) string {
	// Build the result: left portion of base + overlay + right portion of base
	// We need to work with visual columns, skipping ANSI escapes.

	// Simple approach: pad base to totalWidth, then replace columns x..x+overlayWidth
	_ = baseLine
	leftPad := strings.Repeat(" ", x)
	overlayW := lipgloss.Width(overlay)
	rightPadW := totalWidth - x - overlayW
	if rightPadW < 0 {
		rightPadW = 0
	}
	rightPad := strings.Repeat(" ", rightPadW)

	return leftPad + overlay + rightPad
}
