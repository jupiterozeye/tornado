// Package styles defines all visual styling for Tornado using Lip Gloss.
//
// This file centralizes all style definitions, making it easy to:
//   - Maintain consistent look and feel across the application
//   - Implement themes (dark/light mode)
//   - Change colors/sizing in one place
//
// Key Learning - Styling with Lip Gloss:
//   - lipgloss.NewStyle() creates a style
//   - Chain methods like .Foreground().Background().Padding()
//   - Call .Render("text") to apply the style
//   - Use lipgloss.JoinHorizontal/JoinVertical for layout
//
// TODO: Define all application styles:
//   - [ ] Color palette (primary, secondary, accent, error, success)
//   - [ ] Text styles (title, header, body, muted, error)
//   - [ ] Component styles (input, button, table, sidebar)
//   - [ ] Layout helpers (padding, margins, borders)
//   - [ ] Adaptive colors for light/dark terminal backgrounds
//
// References:
//   - https://github.com/charmbracelet/lipgloss
//   - https://github.com/charmbracelet/lipgloss#adaptive-colors
package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Color definitions using ANSI 256 color codes.
// These provide a good balance of compatibility and aesthetics.
// TODO: Adjust colors to match your preferred aesthetic
// TODO: Consider using lipgloss.AdaptiveColor for light/dark backgrounds
var (
	// Primary colors - main application accent
	Primary   = lipgloss.Color("99") // Purple
	PrimaryBg = lipgloss.Color("63") // Lighter purple for backgrounds

	// Secondary colors - supporting elements
	Secondary = lipgloss.Color("241") // Gray

	// Accent colors - highlights and emphasis
	Accent = lipgloss.Color("212") // Pink

	// Status colors
	Success = lipgloss.Color("10") // Green
	Warning = lipgloss.Color("11") // Yellow
	Error   = lipgloss.Color("9")  // Red
	Info    = lipgloss.Color("12") // Blue

	// Text colors
	Text       = lipgloss.Color("252") // Light gray
	TextMuted  = lipgloss.Color("241") // Dimmed gray
	TextBold   = lipgloss.Color("255") // Bright white
	TextAccent = lipgloss.Color("99")  // Purple tinted

	// Background colors
	BgDefault = lipgloss.Color("235") // Dark gray
	BgDark    = lipgloss.Color("234") // Darker gray
	BgLight   = lipgloss.Color("237") // Lighter gray

	// Border colors
	Border      = lipgloss.Color("238")
	BorderFocus = lipgloss.Color("99") // Purple when focused
)

// Styles holds all pre-defined styles for the application.
// TODO: Add more styles as needed for components
type Styles struct {
	// TODO: Add style fields
	//
	// ===== Title and Headers =====
	// Title      lipgloss.Style
	// Header     lipgloss.Style
	// Subheader  lipgloss.Style
	//
	// ===== Text Styles =====
	// Body       lipgloss.Style
	// Muted      lipgloss.Style
	// Bold       lipgloss.Style
	// Error      lipgloss.Style
	// Success    lipgloss.Style
	//
	// ===== Component Styles =====
	// Input      lipgloss.Style
	// InputFocus lipgloss.Style
	// Button     lipgloss.Style
	// ButtonFocus lipgloss.Style
	//
	// ===== Layout Styles =====
	// Box        lipgloss.Style
	// BoxFocus   lipgloss.Style
	// Sidebar    lipgloss.Style
	// StatusBar  lipgloss.Style
}

// Default returns the default style set.
// TODO: Initialize all styles with sensible defaults
func Default() *Styles {
	return &Styles{
		// TODO: Initialize styles
		// Example:
		// Title: lipgloss.NewStyle().
		//     Foreground(Primary).
		//     Bold(true).
		//     Padding(0, 1),
	}
}

// Theme styles (pre-defined style sets)
// TODO: Implement multiple themes if desired

// DarkTheme returns styles optimized for dark terminal backgrounds.
func DarkTheme() *Styles {
	return Default()
}

// LightTheme returns styles optimized for light terminal backgrounds.
// TODO: Implement light theme with appropriate color adjustments
func LightTheme() *Styles {
	return Default()
}

// Component style helpers
// TODO: Add helper functions for common styling patterns

// Box creates a bordered box style.
//
// TODO: Implement with customizable borders
func Box(focused bool) lipgloss.Style {
	borderColor := Border
	if focused {
		borderColor = BorderFocus
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)
}

// Input creates a style for text input fields.
//
// TODO: Implement with focus state
func Input(focused bool) lipgloss.Style {
	style := lipgloss.NewStyle().
		Padding(0, 1).
		Background(BgLight)

	if focused {
		style = style.Border(lipgloss.NormalBorder()).
			BorderForeground(BorderFocus)
	}

	return style
}

// Button creates a style for clickable buttons.
//
// TODO: Implement button styling
func Button(focused bool) lipgloss.Style {
	bg := Secondary
	if focused {
		bg = Primary
	}

	return lipgloss.NewStyle().
		Foreground(TextBold).
		Background(bg).
		Padding(0, 2).
		Bold(focused)
}

// Table creates styles for table components.
//
// TODO: Implement table header and cell styles
func TableHeader() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true).
		Padding(0, 1)
}

func TableRow(alt bool) lipgloss.Style {
	bg := BgDefault
	if alt {
		bg = BgLight
	}
	return lipgloss.NewStyle().
		Background(bg).
		Padding(0, 1)
}

// Layout helpers

// HorizontalPad adds horizontal padding to content.
func HorizontalPad(width int, content string) string {
	return lipgloss.NewStyle().
		PaddingLeft(width).
		Render(content)
}

// Center centers content within a given width.
func Center(width int, content string) string {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(content)
}

// Width helpers

// Clamp ensures a value is within min and max bounds.
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
