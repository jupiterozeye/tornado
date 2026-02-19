// This file centralizes all style definitions, making it easy to:
//   - Maintain consistent look and feel across the application
//   - Implement themes (dark/light mode)
//   - Change colors/sizing in one place
//
// References:
//   - https://github.com/charmbracelet/lipgloss
//   - https://github.com/charmbracelet/lipgloss#adaptive-colors
package styles

import (
	"github.com/charmbracelet/lipgloss"
)

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
type Styles struct {
	// ===== Title and Headers =====
	Title     lipgloss.Style
	Header    lipgloss.Style
	Subheader lipgloss.Style

	// ===== Text Styles =====
	Body    lipgloss.Style
	Muted   lipgloss.Style
	Bold    lipgloss.Style
	Error   lipgloss.Style
	Success lipgloss.Style

	// ===== Component Styles =====
	Input       lipgloss.Style
	InputFocus  lipgloss.Style
	Button      lipgloss.Style
	ButtonFocus lipgloss.Style

	// ===== Layout Styles =====
	Box       lipgloss.Style
	BoxFocus  lipgloss.Style
	Sidebar   lipgloss.Style
	StatusBar lipgloss.Style
}

// Default returns the default style set.
func Default() *Styles {
	return &Styles{
		Title:       lipgloss.NewStyle().Foreground(Primary).Bold(true).Padding(0, 1),
		Header:      lipgloss.NewStyle().Foreground(Secondary).Bold(true).Padding(0, 1),
		Subheader:   lipgloss.NewStyle().Foreground(Accent).Bold(true).Padding(0, 1),
		Body:        lipgloss.NewStyle().Foreground(Text).Padding(0, 1),
		Muted:       lipgloss.NewStyle().Foreground(TextMuted).Padding(0, 1),
		Bold:        lipgloss.NewStyle().Foreground(TextBold).Bold(true).Padding(0, 1),
		Error:       lipgloss.NewStyle().Foreground(Error).Padding(0, 1),
		Success:     lipgloss.NewStyle().Foreground(Success).Padding(0, 1),
		Input:       lipgloss.NewStyle().Foreground(Text).Background(BgLight).Padding(0, 1),
		InputFocus:  lipgloss.NewStyle().Foreground(Text).Background(BgLight).Border(lipgloss.NormalBorder()).BorderForeground(BorderFocus).Padding(0, 1),
		Button:      lipgloss.NewStyle().Foreground(TextBold).Background(Secondary).Padding(0, 2),
		ButtonFocus: lipgloss.NewStyle().Foreground(TextBold).Background(Primary).Padding(0, 2).Bold(true),
		Box:         lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Border).Padding(0, 1),
		BoxFocus:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(BorderFocus).Padding(0, 1),
		Sidebar:     lipgloss.NewStyle().Background(BgDark).Foreground(Text).Padding(0, 1),
		StatusBar:   lipgloss.NewStyle().Background(BgLight).Foreground(TextMuted).Padding(0, 1),
	}
}

// Theme styles (pre-defined style sets)

// DarkTheme returns styles optimized for dark terminal backgrounds.
// TODO: Implement dark theme
func DarkTheme() *Styles {
	return Default()
}

// LightTheme returns styles optimized for light terminal backgrounds.
// TODO: Implement light theme
func LightTheme() *Styles {
	return Default()
}

// Component style helpers

// Box creates a bordered box style.
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
