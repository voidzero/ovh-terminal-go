// internal/ui/handlers/updates.go
package handlers

import (
	"ovh-terminal/internal/logger"
	"ovh-terminal/internal/ui/common"
	"ovh-terminal/internal/ui/layout"
	"ovh-terminal/internal/ui/styles"

	tea "github.com/charmbracelet/bubbletea"
)

// KeyHandler defines a function type for handling key presses
type KeyHandler func(common.UIModel) (tea.Model, tea.Cmd)

// KeyMap defines keyboard mappings
var KeyMap = map[string]KeyHandler{
	"q":      handleQuit,
	"ctrl+c": handleQuit,
	"f1":     handleHelp,
	"tab":    handlePaneToggle,
	"enter":  handleEnter,
	// "up":     handleUpNav,
	// "k":      handleUpNav,
	// "down":   handleDownNav,
	// "j":      handleDownNav,
	"g": handleTopNav,
	"G": handleBottomNav,
}

// LayoutManager singleton
var layoutManager *layout.Manager

// HandleKeyMsg processes keyboard input messages
func HandleKeyMsg(model common.UIModel, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !model.IsReady() {
		return model, nil
	}

	// Check for registered key handler
	if handler, exists := KeyMap[msg.String()]; exists {
		return handler(model)
	}

	return model, nil
}

// Key handlers
func handleQuit(model common.UIModel) (tea.Model, tea.Cmd) {
	return model, tea.Quit
}

func handleHelp(model common.UIModel) (tea.Model, tea.Cmd) {
	model.ToggleHelp()
	return model, nil
}

func handlePaneToggle(model common.UIModel) (tea.Model, tea.Cmd) {
	model.ToggleActivePane()
	styles.UpdateBorderStyles(model.GetActivePane())
	return model, nil
}

func handleEnter(model common.UIModel) (tea.Model, tea.Cmd) {
	if model.GetActivePane() != "menu" {
		return model, nil
	}

	list := model.GetList()
	selectedItem := list.SelectedItem()

	logger.Log.Debug("Enter pressed",
		"activePane", model.GetActivePane(),
		"selectedItem", selectedItem)

	if selectedItem == nil {
		logger.Log.Debug("No item selected")
		return model, nil
	}

	if menuItem, ok := selectedItem.(common.MenuItem); ok {
		logger.Log.Debug("Item type check",
			"isMenuItem", ok,
			"type", menuItem.GetType())

		if err := HandleCommand(model, menuItem); err != nil {
			logger.Log.Error("Error handling command",
				"error", err,
				"item", menuItem.Title())
			return model, nil
		}

		// Update layout after command execution
		ensureLayoutManager(model).Update()
		return model, nil
	}

	return model, nil
}

// Navigation handlers
func handleUpNav(model common.UIModel) (tea.Model, tea.Cmd) {
	if model.GetActivePane() == "content" {
		return model, model.UpdateViewport(tea.KeyMsg{Type: tea.KeyUp})
	}
	return model, model.UpdateList(tea.KeyMsg{Type: tea.KeyUp})
}

func handleDownNav(model common.UIModel) (tea.Model, tea.Cmd) {
	if model.GetActivePane() == "content" {
		return model, model.UpdateViewport(tea.KeyMsg{Type: tea.KeyDown})
	}
	return model, model.UpdateList(tea.KeyMsg{Type: tea.KeyDown})
}

func handleTopNav(model common.UIModel) (tea.Model, tea.Cmd) {
	if model.GetActivePane() == "content" {
		vp := model.GetViewport()
		vp.GotoTop()
		model.SetViewport(vp)
		model.SetStatusMessage("Navigated to top of content")
		return model, nil
	}
	return model, model.UpdateList(tea.KeyMsg{Type: tea.KeyHome})
}

func handleBottomNav(model common.UIModel) (tea.Model, tea.Cmd) {
	if model.GetActivePane() == "content" {
		vp := model.GetViewport()
		vp.GotoBottom()
		model.SetViewport(vp)
		model.SetStatusMessage("Navigated to bottom of content")
		return model, nil
	}
	return model, model.UpdateList(tea.KeyMsg{Type: tea.KeyEnd})
}

// HandleWindowSizeMsg processes window resize messages
func HandleWindowSizeMsg(model common.UIModel, msg tea.WindowSizeMsg) tea.Model {
	logger.Log.Debug("Window size message received",
		"width", msg.Width,
		"height", msg.Height)

	mgr := ensureLayoutManager(model)
	model.SetSize(msg.Width, msg.Height)

	if !mgr.ValidateWindowSize(msg.Width, msg.Height) {
		model.SetStatusMessage("Window too small - please resize")
		return model
	}

	mgr.Update()
	return model
}

// ensureLayoutManager creates or returns the existing layout manager
func ensureLayoutManager(model common.UIModel) *layout.Manager {
	if layoutManager == nil {
		layoutManager = layout.NewManager(model)
	}
	return layoutManager
}
