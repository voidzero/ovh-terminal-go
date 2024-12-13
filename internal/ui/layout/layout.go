// internal/ui/layout/layout.go

// Package layout handles UI component positioning and sizing
package layout

import (
	"ovh-terminal/internal/logger"
	"ovh-terminal/internal/ui/common"
)

const (
	// Minimum window dimensions
	MinWidth  = 80
	MinHeight = 20

	// Fixed UI element dimensions
	StatusBarSpace  = 3  // Space needed for status bar including borders
	UIElementsSpace = 2  // Space needed for title and borders
	HorizontalSpace = 9  // Total horizontal space for margins/padding
	MenuBaseWidth   = 32 // Base menu width
)

// Manager handles layout calculations and updates
type Manager struct {
	model common.UIModel
}

// NewManager creates a new layout manager
func NewManager(model common.UIModel) *Manager {
	return &Manager{
		model: model,
	}
}

// Update recalculates and applies layout dimensions
func (m *Manager) Update() {
	logger.Log.Debug("Starting layout update",
		"ready", m.model.IsReady())

	if !m.model.IsReady() {
		return
	}

	// Calculate and update content sizes
	contentWidth := m.model.GetWidth() - MenuBaseWidth - HorizontalSpace
	contentHeight := m.model.GetHeight() - StatusBarSpace - UIElementsSpace

	logger.Log.Debug("Layout calculations",
		"content_width", contentWidth,
		"content_height", contentHeight)

	if contentWidth > 0 && contentHeight > 0 {
		list := m.model.GetList()
		list.SetSize(MenuBaseWidth, contentHeight)

		viewport := m.model.GetViewport()
		viewport.Width = contentWidth
		viewport.Height = contentHeight
	}
}

// ValidateWindowSize checks if the window size is sufficient
func (m *Manager) ValidateWindowSize(width, height int) bool {
	return width >= MinWidth && height >= MinHeight
}

// CalculateStatusBarWidth calculates width for the status bar
func (m *Manager) CalculateStatusBarWidth() int {
	return m.model.GetWidth() - 2
}
