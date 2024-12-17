// internal/ui/layout/layout.go
package layout

import (
	"ovh-terminal/internal/logger"
	"ovh-terminal/internal/ui/common"
)

// UI component dimensions
const (
	MinWidth  = 80 // Minimum window width
	MinHeight = 20 // Minimum window height
	MenuWidth = 32 // Base menu width
)

// UI spacing constants
const (
	StatusBarSpace  = 3 // Space for status bar including borders
	UIElementsSpace = 2 // Space for title and borders
	HorizontalSpace = 9 // Total horizontal space for margins/padding
)

// Dimensions represents UI component dimensions
type Dimensions struct {
	ContentWidth  int
	ContentHeight int
	MenuWidth     int
	StatusWidth   int
}

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

// calculateDimensions computes dimensions for all UI components
func (m *Manager) calculateDimensions() Dimensions {
	totalWidth := m.model.GetWidth()
	totalHeight := m.model.GetHeight()

	dims := Dimensions{
		MenuWidth:     MenuWidth,
		ContentWidth:  totalWidth - MenuWidth - HorizontalSpace,
		ContentHeight: totalHeight - StatusBarSpace - UIElementsSpace,
		StatusWidth:   totalWidth - 2,
	}

	logger.Log.Debug("Layout dimensions calculated",
		"total_width", totalWidth,
		"total_height", totalHeight,
		"content_width", dims.ContentWidth,
		"content_height", dims.ContentHeight)

	return dims
}

// Update recalculates and applies layout dimensions
func (m *Manager) Update() {
	if !m.model.IsReady() {
		logger.Log.Debug("Model not ready, skipping layout update")
		return
	}

	dims := m.calculateDimensions()
	if dims.ContentWidth <= 0 || dims.ContentHeight <= 0 {
		logger.Log.Debug("Invalid dimensions, skipping update",
			"content_width", dims.ContentWidth,
			"content_height", dims.ContentHeight)
		return
	}

	// Update list dimensions
	list := m.model.GetList()
	list.SetSize(dims.MenuWidth, dims.ContentHeight)

	// Update viewport dimensions
	viewport := m.model.GetViewport()
	viewport.Width = dims.ContentWidth
	viewport.Height = dims.ContentHeight

	logger.Log.Debug("Layout updated successfully")
}

// ValidateWindowSize checks if the window size is sufficient
func (m *Manager) ValidateWindowSize(width, height int) bool {
	isValid := width >= MinWidth && height >= MinHeight

	if !isValid {
		logger.Log.Debug("Window size validation failed",
			"width", width,
			"height", height,
			"min_width", MinWidth,
			"min_height", MinHeight)
	}

	return isValid
}

// CalculateStatusBarWidth returns the appropriate width for the status bar
func (m *Manager) CalculateStatusBarWidth() int {
	return m.calculateDimensions().StatusWidth
}
