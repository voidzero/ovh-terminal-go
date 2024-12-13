// internal/ui/styles/components.go
// Package styles provides UI styling definitions
package styles

import "github.com/charmbracelet/lipgloss"

// Common style objects for UI components
var (
	// DocStyle defines the main document container style
	DocStyle = lipgloss.NewStyle().
			MarginLeft(1).
			MarginRight(1)

	// TitleStyle defines the style for section titles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			Align(lipgloss.Center).
			Width(28)

	// MenuStyle defines the style for the menu pane
	MenuStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ActiveBorderColor).
			PaddingLeft(1).
			PaddingRight(1).
			Width(32)

	// ContentStyle defines the style for the content pane
	ContentStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(InactiveBorderColor).
			PaddingLeft(1).
			PaddingRight(1).
			MarginLeft(2)

	// SelectedItemStyle defines the style for selected menu items
	SelectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(SelectedFg).
				Background(SelectedBg)

	// NormalItemStyle defines the style for normal (unselected) menu items
	NormalItemStyle = lipgloss.NewStyle().
			Foreground(NormalTextColor)

	// DimmedStyle defines the style for less prominent text
	DimmedStyle = lipgloss.NewStyle().
			Foreground(DimmedTextColor)

	// StatusStyle defines the style for the status bar
	StatusStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(InactiveBorderColor).
			Padding(0, 0).
			MarginTop(0)
)

// UpdateBorderStyles updates the border styles based on the active pane
func UpdateBorderStyles(activePane string) {
	if activePane == "menu" {
		MenuStyle = MenuStyle.BorderForeground(ActiveBorderColor)
		ContentStyle = ContentStyle.BorderForeground(InactiveBorderColor)
	} else {
		MenuStyle = MenuStyle.BorderForeground(InactiveBorderColor)
		ContentStyle = ContentStyle.BorderForeground(ActiveBorderColor)
	}
}

// WithWidth returns a new style with the specified width
func WithWidth(style lipgloss.Style, width int) lipgloss.Style {
	return style.Width(width)
}
