// internal/ui/styles/colors.go
package styles

import "github.com/charmbracelet/lipgloss"

// ColorScheme defines a set of colors for the application
type ColorScheme struct {
	Primary      lipgloss.Color
	Secondary    lipgloss.Color
	Background   lipgloss.Color
	Foreground   lipgloss.Color
	BorderActive lipgloss.Color
	BorderNormal lipgloss.Color
	Success      lipgloss.Color
	Warning      lipgloss.Color
	Error        lipgloss.Color
	Selection    SelectionColors
	Text         TextColors
}

// SelectionColors defines colors for selected items
type SelectionColors struct {
	Background lipgloss.Color
	Foreground lipgloss.Color
}

// TextColors defines different text color variants
type TextColors struct {
	Normal lipgloss.Color
	Dimmed lipgloss.Color
	Bright lipgloss.Color
}

// Predefined color schemes
var (
	// DefaultScheme is the default color scheme
	DefaultScheme = ColorScheme{
		Primary:      lipgloss.Color("#7CE38B"),
		Secondary:    lipgloss.Color("#5A5A5A"),
		Background:   lipgloss.Color("#222222"),
		Foreground:   lipgloss.Color("#FFFFFF"),
		BorderActive: lipgloss.Color("#888888"),
		BorderNormal: lipgloss.Color("#444444"),
		Success:      lipgloss.Color("#00FF00"),
		Warning:      lipgloss.Color("#FFAA00"),
		Error:        lipgloss.Color("#FF0000"),
		Selection: SelectionColors{
			Background: lipgloss.Color("#2D79C7"),
			Foreground: lipgloss.Color("#FFFF22"),
		},
		Text: TextColors{
			Normal: lipgloss.Color("245"),
			Dimmed: lipgloss.Color("241"),
			Bright: lipgloss.Color("255"),
		},
	}

	// LightScheme is a light theme variant
	LightScheme = ColorScheme{
		Primary:      lipgloss.Color("#2E8B57"),
		Secondary:    lipgloss.Color("#708090"),
		Background:   lipgloss.Color("#FFFFFF"),
		Foreground:   lipgloss.Color("#000000"),
		BorderActive: lipgloss.Color("#000000"),
		BorderNormal: lipgloss.Color("#CCCCCC"),
		Success:      lipgloss.Color("#008000"),
		Warning:      lipgloss.Color("#FFA500"),
		Error:        lipgloss.Color("#FF0000"),
		Selection: SelectionColors{
			Background: lipgloss.Color("#ADD8E6"),
			Foreground: lipgloss.Color("#000000"),
		},
		Text: TextColors{
			Normal: lipgloss.Color("#000000"),
			Dimmed: lipgloss.Color("#666666"),
			Bright: lipgloss.Color("#000000"),
		},
	}

	// Current active color scheme
	ActiveScheme = DefaultScheme
)

// SetColorScheme updates the active color scheme
func SetColorScheme(scheme ColorScheme) {
	ActiveScheme = scheme
	updateAllStyles()
}

// GetActiveScheme returns the current color scheme
func GetActiveScheme() ColorScheme {
	return ActiveScheme
}

// UpdateTheme sets a predefined theme
func UpdateTheme(theme string) {
	switch theme {
	case "light":
		SetColorScheme(LightScheme)
	default:
		SetColorScheme(DefaultScheme)
	}
}

// Color getters for convenience
func GetPrimaryColor() lipgloss.Color      { return ActiveScheme.Primary }
func GetSecondaryColor() lipgloss.Color    { return ActiveScheme.Secondary }
func GetBorderActiveColor() lipgloss.Color { return ActiveScheme.BorderActive }
func GetBorderNormalColor() lipgloss.Color { return ActiveScheme.BorderNormal }
func GetSelectionFg() lipgloss.Color       { return ActiveScheme.Selection.Foreground }
func GetSelectionBg() lipgloss.Color       { return ActiveScheme.Selection.Background }
func GetNormalTextColor() lipgloss.Color   { return ActiveScheme.Text.Normal }
func GetDimmedTextColor() lipgloss.Color   { return ActiveScheme.Text.Dimmed }
func GetBrightTextColor() lipgloss.Color   { return ActiveScheme.Text.Bright }

// State-based color functions
func GetBorderColor(isActive bool) lipgloss.Color {
	if isActive {
		return ActiveScheme.BorderActive
	}
	return ActiveScheme.BorderNormal
}

// Status colors
func GetStatusColor(status string) lipgloss.Color {
	switch status {
	case "success", "active", "enabled":
		return ActiveScheme.Success
	case "warning", "pending":
		return ActiveScheme.Warning
	case "error", "failed", "disabled":
		return ActiveScheme.Error
	default:
		return ActiveScheme.Text.Normal
	}
}

// updateAllStyles refreshes all component styles with new colors
func updateAllStyles() {
	// Update component styles in components.go
	UpdateComponentStyles()

	// Update any other style-dependent components
	MenuStyle = MenuStyle.BorderForeground(GetBorderColor(true))
	ContentStyle = ContentStyle.BorderForeground(GetBorderColor(false))
}
