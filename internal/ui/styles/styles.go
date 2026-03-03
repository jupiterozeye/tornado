// This file centralizes all style definitions, making it easy to:
//   - Maintain consistent look and feel across the application
//   - Implement themes (dark/light mode)
//   - Change colors/sizing in one place
//
// References:
//   - https://charm.land/lipgloss/v2
//   - https://charm.land/lipgloss/v2#adaptive-colors
package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type ThemePalette struct {
	Name                                  string
	Primary, PrimaryBg                    color.Color
	Secondary, Accent                     color.Color
	Success, Warning, Error, Info         color.Color
	Text, TextMuted, TextBold, TextAccent color.Color
	BgDefault, BgDark, BgLight            color.Color
	Border, BorderFocus                   color.Color
}

var themeOrder = []string{
	"nord", "gruvbox", "gruvbox-light", "tokyo-night", "solarized-dark", "solarized-light", "catppuccin-mocha", "catppuccin-latte", "rose-pine", "rose-pine-dawn", "dracula", "everforest", "kanagawa", "hackerman", "matte-black", "ristretto", "osaka-jade",
}

var palettes = map[string]ThemePalette{
	"nord":             {"nord", lipgloss.Color("110"), lipgloss.Color("67"), lipgloss.Color("109"), lipgloss.Color("181"), lipgloss.Color("108"), lipgloss.Color("179"), lipgloss.Color("174"), lipgloss.Color("67"), lipgloss.Color("252"), lipgloss.Color("245"), lipgloss.Color("255"), lipgloss.Color("110"), lipgloss.Color("237"), lipgloss.Color("236"), lipgloss.Color("239"), lipgloss.Color("240"), lipgloss.Color("110")},
	"gruvbox":          {"gruvbox", lipgloss.Color("108"), lipgloss.Color("130"), lipgloss.Color("172"), lipgloss.Color("108"), lipgloss.Color("142"), lipgloss.Color("214"), lipgloss.Color("167"), lipgloss.Color("109"), lipgloss.Color("223"), lipgloss.Color("246"), lipgloss.Color("230"), lipgloss.Color("108"), lipgloss.Color("235"), lipgloss.Color("234"), lipgloss.Color("237"), lipgloss.Color("239"), lipgloss.Color("108")},
	"gruvbox-light":    {"gruvbox-light", lipgloss.Color("65"), lipgloss.Color("100"), lipgloss.Color("136"), lipgloss.Color("65"), lipgloss.Color("100"), lipgloss.Color("130"), lipgloss.Color("124"), lipgloss.Color("67"), lipgloss.Color("237"), lipgloss.Color("244"), lipgloss.Color("234"), lipgloss.Color("65"), lipgloss.Color("230"), lipgloss.Color("223"), lipgloss.Color("187"), lipgloss.Color("180"), lipgloss.Color("65")},
	"tokyo-night":      {"tokyo-night", lipgloss.Color("111"), lipgloss.Color("68"), lipgloss.Color("68"), lipgloss.Color("111"), lipgloss.Color("114"), lipgloss.Color("179"), lipgloss.Color("167"), lipgloss.Color("110"), lipgloss.Color("252"), lipgloss.Color("245"), lipgloss.Color("255"), lipgloss.Color("111"), lipgloss.Color("235"), lipgloss.Color("234"), lipgloss.Color("237"), lipgloss.Color("239"), lipgloss.Color("111")},
	"solarized-dark":   {"solarized-dark", lipgloss.Color("136"), lipgloss.Color("37"), lipgloss.Color("244"), lipgloss.Color("136"), lipgloss.Color("64"), lipgloss.Color("166"), lipgloss.Color("160"), lipgloss.Color("33"), lipgloss.Color("250"), lipgloss.Color("244"), lipgloss.Color("255"), lipgloss.Color("136"), lipgloss.Color("236"), lipgloss.Color("235"), lipgloss.Color("238"), lipgloss.Color("240"), lipgloss.Color("136")},
	"solarized-light":  {"solarized-light", lipgloss.Color("64"), lipgloss.Color("37"), lipgloss.Color("37"), lipgloss.Color("64"), lipgloss.Color("64"), lipgloss.Color("166"), lipgloss.Color("160"), lipgloss.Color("33"), lipgloss.Color("238"), lipgloss.Color("244"), lipgloss.Color("234"), lipgloss.Color("64"), lipgloss.Color("230"), lipgloss.Color("254"), lipgloss.Color("187"), lipgloss.Color("244"), lipgloss.Color("64")},
	"catppuccin-mocha": {"catppuccin-mocha", lipgloss.Color("111"), lipgloss.Color("68"), lipgloss.Color("104"), lipgloss.Color("111"), lipgloss.Color("114"), lipgloss.Color("179"), lipgloss.Color("167"), lipgloss.Color("68"), lipgloss.Color("252"), lipgloss.Color("245"), lipgloss.Color("255"), lipgloss.Color("111"), lipgloss.Color("236"), lipgloss.Color("235"), lipgloss.Color("239"), lipgloss.Color("240"), lipgloss.Color("111")},
	"catppuccin-latte": {"catppuccin-latte", lipgloss.Color("68"), lipgloss.Color("24"), lipgloss.Color("110"), lipgloss.Color("68"), lipgloss.Color("64"), lipgloss.Color("172"), lipgloss.Color("167"), lipgloss.Color("24"), lipgloss.Color("238"), lipgloss.Color("245"), lipgloss.Color("16"), lipgloss.Color("68"), lipgloss.Color("254"), lipgloss.Color("255"), lipgloss.Color("252"), lipgloss.Color("249"), lipgloss.Color("68")},
	"rose-pine":        {"rose-pine", lipgloss.Color("175"), lipgloss.Color("103"), lipgloss.Color("145"), lipgloss.Color("175"), lipgloss.Color("108"), lipgloss.Color("180"), lipgloss.Color("174"), lipgloss.Color("110"), lipgloss.Color("252"), lipgloss.Color("245"), lipgloss.Color("255"), lipgloss.Color("175"), lipgloss.Color("236"), lipgloss.Color("235"), lipgloss.Color("239"), lipgloss.Color("240"), lipgloss.Color("175")},
	"rose-pine-dawn":   {"rose-pine-dawn", lipgloss.Color("103"), lipgloss.Color("67"), lipgloss.Color("145"), lipgloss.Color("103"), lipgloss.Color("101"), lipgloss.Color("180"), lipgloss.Color("167"), lipgloss.Color("67"), lipgloss.Color("238"), lipgloss.Color("245"), lipgloss.Color("16"), lipgloss.Color("103"), lipgloss.Color("254"), lipgloss.Color("255"), lipgloss.Color("252"), lipgloss.Color("249"), lipgloss.Color("103")},
	"dracula":          {"dracula", lipgloss.Color("141"), lipgloss.Color("63"), lipgloss.Color("212"), lipgloss.Color("141"), lipgloss.Color("84"), lipgloss.Color("227"), lipgloss.Color("203"), lipgloss.Color("75"), lipgloss.Color("255"), lipgloss.Color("246"), lipgloss.Color("255"), lipgloss.Color("141"), lipgloss.Color("236"), lipgloss.Color("235"), lipgloss.Color("238"), lipgloss.Color("240"), lipgloss.Color("141")},
	"everforest":       {"everforest", lipgloss.Color("150"), lipgloss.Color("71"), lipgloss.Color("108"), lipgloss.Color("109"), lipgloss.Color("150"), lipgloss.Color("179"), lipgloss.Color("174"), lipgloss.Color("108"), lipgloss.Color("223"), lipgloss.Color("245"), lipgloss.Color("255"), lipgloss.Color("150"), lipgloss.Color("235"), lipgloss.Color("236"), lipgloss.Color("239"), lipgloss.Color("240"), lipgloss.Color("150")},
	"kanagawa":         {"kanagawa", lipgloss.Color("110"), lipgloss.Color("67"), lipgloss.Color("110"), lipgloss.Color("175"), lipgloss.Color("113"), lipgloss.Color("215"), lipgloss.Color("174"), lipgloss.Color("110"), lipgloss.Color("230"), lipgloss.Color("245"), lipgloss.Color("255"), lipgloss.Color("110"), lipgloss.Color("235"), lipgloss.Color("234"), lipgloss.Color("237"), lipgloss.Color("239"), lipgloss.Color("110")},
	"hackerman":        {"hackerman", lipgloss.Color("46"), lipgloss.Color("22"), lipgloss.Color("34"), lipgloss.Color("46"), lipgloss.Color("46"), lipgloss.Color("226"), lipgloss.Color("196"), lipgloss.Color("46"), lipgloss.Color("46"), lipgloss.Color("34"), lipgloss.Color("46"), lipgloss.Color("46"), lipgloss.Color("233"), lipgloss.Color("232"), lipgloss.Color("234"), lipgloss.Color("28"), lipgloss.Color("46")},
	"matte-black":      {"matte-black", lipgloss.Color("255"), lipgloss.Color("245"), lipgloss.Color("245"), lipgloss.Color("255"), lipgloss.Color("46"), lipgloss.Color("214"), lipgloss.Color("203"), lipgloss.Color("75"), lipgloss.Color("252"), lipgloss.Color("245"), lipgloss.Color("255"), lipgloss.Color("255"), lipgloss.Color("233"), lipgloss.Color("232"), lipgloss.Color("234"), lipgloss.Color("238"), lipgloss.Color("255")},
	"ristretto":        {"ristretto", lipgloss.Color("224"), lipgloss.Color("138"), lipgloss.Color("181"), lipgloss.Color("224"), lipgloss.Color("114"), lipgloss.Color("180"), lipgloss.Color("174"), lipgloss.Color("109"), lipgloss.Color("223"), lipgloss.Color("246"), lipgloss.Color("255"), lipgloss.Color("224"), lipgloss.Color("235"), lipgloss.Color("234"), lipgloss.Color("237"), lipgloss.Color("239"), lipgloss.Color("224")},
	"osaka-jade":       {"osaka-jade", lipgloss.Color("115"), lipgloss.Color("65"), lipgloss.Color("72"), lipgloss.Color("115"), lipgloss.Color("115"), lipgloss.Color("179"), lipgloss.Color("174"), lipgloss.Color("72"), lipgloss.Color("152"), lipgloss.Color("245"), lipgloss.Color("255"), lipgloss.Color("115"), lipgloss.Color("233"), lipgloss.Color("234"), lipgloss.Color("236"), lipgloss.Color("238"), lipgloss.Color("115")},
}

var currentThemeIndex = 0

func init() {
	if p, ok := palettes[themeOrder[currentThemeIndex]]; ok {
		applyPalette(p)
	}
}

func applyPalette(p ThemePalette) {
	Primary = p.Primary
	PrimaryBg = p.PrimaryBg
	Secondary = p.Secondary
	Accent = p.Accent
	Success = p.Success
	Warning = p.Warning
	Error = p.Error
	Info = p.Info
	Text = p.Text
	TextMuted = p.TextMuted
	TextBold = p.TextBold
	TextAccent = p.TextAccent
	BgDefault = p.BgDefault
	BgDark = p.BgDark
	BgLight = p.BgLight
	Border = p.Border
	BorderFocus = p.BorderFocus
}

func CycleTheme() string {
	currentThemeIndex = (currentThemeIndex + 1) % len(themeOrder)
	name := themeOrder[currentThemeIndex]
	if p, ok := palettes[name]; ok {
		applyPalette(p)
	}
	return name
}

func AvailableThemes() []string {
	out := make([]string, len(themeOrder))
	copy(out, themeOrder)
	return out
}

func SetTheme(name string) bool {
	for i, n := range themeOrder {
		if n == name {
			currentThemeIndex = i
			if p, ok := palettes[name]; ok {
				applyPalette(p)
				return true
			}
			return false
		}
	}
	return false
}

func CurrentTheme() string {
	return themeOrder[currentThemeIndex]
}

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

// DialogBox creates a consistent popup style.
func DialogBox() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderFocus).
		Background(BgDark)
}

// FieldContainer creates a focused/unfocused input container style.
func FieldContainer(focused bool) lipgloss.Style {
	borderColor := Border
	if focused {
		borderColor = BorderFocus
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Background(BgDark)
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
