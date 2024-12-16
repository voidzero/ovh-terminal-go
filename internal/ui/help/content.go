// internal/ui/help/content.go

// Package help provides help screen functionality
package help

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Help overlay styling
	helpStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#888888")).
			Padding(1, 2)

	// Section title styling
	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7CE38B"))

	// Keyboard shortcut styling
	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF22"))

	// Description styling
	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))
)

// section creates a formatted help section
func section(title string) string {
	return sectionStyle.Render(title)
}

// shortcut formats a keyboard shortcut with description
func shortcut(key, description string) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		keyStyle.Render(key),
		"  ",
		descStyle.Render(description),
	)
}

// GetHelpContent returns formatted help content
func GetHelpContent(width, height int) string {
	// Calculate available space for content
	availWidth := width - 6   // Account for borders and padding
	availHeight := height - 4 // Account for borders and padding

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		section("Navigation"),
		shortcut("↑/k, ↓/j", "Move up/down"),
		shortcut("g/G", "Go to top/bottom"),
		shortcut("Tab", "Switch between menu and content"),
		"",
		section("Menu Actions"),
		shortcut("Enter", "Select menu item / Toggle section"),
		shortcut("←/→", "Collapse/Expand section"),
		"",
		section("General"),
		shortcut("F1", "Toggle this help screen"),
		shortcut("q/Ctrl+c", "Quit application"),
	)

	return helpStyle.
		Width(availWidth).
		Height(availHeight).
		Render(content)
}
