// internal/ui/styles/colors.go
package styles

import "github.com/charmbracelet/lipgloss"

// Color definitions for the UI
var (
	// PrimaryColor is used for main UI elements like titles
	PrimaryColor = lipgloss.Color("#7CE38B")

	// SecondaryColor is used for less prominent UI elements
	SecondaryColor = lipgloss.Color("#5A5A5A")

	// SelectedFg is the text color for selected items
	SelectedFg = lipgloss.Color("#FFFF22")

	// SelectedBg is the background color for selected items
	SelectedBg = lipgloss.Color("#2D79C7")

	// ActiveBorderColor is used for borders of the active pane
	ActiveBorderColor = lipgloss.Color("#888888")

	// InactiveBorderColor is used for borders of inactive panes
	InactiveBorderColor = lipgloss.Color("#444444")

	// NormalTextColor is used for regular text
	NormalTextColor = lipgloss.Color("245")

	// DimmedTextColor is used for less prominent text
	DimmedTextColor = lipgloss.Color("241")
)
