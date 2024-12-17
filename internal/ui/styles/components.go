// internal/ui/styles/components.go
package styles

import "github.com/charmbracelet/lipgloss"

// BaseStyle provides common styling options
var BaseStyle = lipgloss.NewStyle()

// BorderStyle provides common border styling
var BorderStyle = BaseStyle.
	BorderStyle(lipgloss.RoundedBorder())

// Common style objects for UI components
var (
	// DocStyle defines the main document container style
	DocStyle = BaseStyle.
			MarginLeft(1).
			MarginRight(1)

	// TitleStyle defines the style for section titles
	TitleStyle = BaseStyle.
			Bold(true).
			Foreground(GetPrimaryColor()).
			Align(lipgloss.Center).
			Width(28)

	// MenuStyle defines the style for the menu pane
	MenuStyle = BorderStyle.
			BorderForeground(GetBorderActiveColor()).
			PaddingLeft(1).
			PaddingRight(1).
			Width(32)

	// ContentStyle defines the style for the content pane
	ContentStyle = BorderStyle.
			BorderForeground(GetBorderNormalColor()).
			PaddingLeft(1).
			PaddingRight(1).
			MarginLeft(2)

	// SelectedItemStyle defines the style for selected menu items
	SelectedItemStyle = BaseStyle.
				Bold(true).
				Foreground(GetSelectionFg()).
				Background(GetSelectionBg())

	// NormalItemStyle defines the style for normal (unselected) menu items
	NormalItemStyle = BaseStyle.
			Foreground(GetNormalTextColor())

	// DimmedStyle defines the style for less prominent text
	DimmedStyle = BaseStyle.
			Foreground(GetDimmedTextColor())

	// StatusStyle defines the style for the status bar
	StatusStyle = BorderStyle.
			BorderForeground(GetBorderNormalColor()).
			Padding(0, 0).
			MarginTop(0)
)

// UpdateComponentStyles refreshes all component styles with current colors
func UpdateComponentStyles() {
	TitleStyle = TitleStyle.
		Bold(true).
		Foreground(GetPrimaryColor()).
		Align(lipgloss.Center).
		Width(28)

	SelectedItemStyle = BaseStyle.
		Bold(true).
		Foreground(GetSelectionFg()).
		Background(GetSelectionBg())

	NormalItemStyle = BaseStyle.
		Foreground(GetNormalTextColor())

	DimmedStyle = BaseStyle.
		Foreground(GetDimmedTextColor())
}

// UpdateBorderStyles updates the border styles based on the active pane
func UpdateBorderStyles(activePane string) {
	if activePane == "menu" {
		MenuStyle = MenuStyle.BorderForeground(GetBorderActiveColor())
		ContentStyle = ContentStyle.BorderForeground(GetBorderNormalColor())
	} else {
		MenuStyle = MenuStyle.BorderForeground(GetBorderNormalColor())
		ContentStyle = ContentStyle.BorderForeground(GetBorderActiveColor())
	}
}

