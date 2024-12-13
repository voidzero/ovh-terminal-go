// Package handlers provides UI event handling
package handlers

import (
	"fmt"

	"ovh-terminal/internal/logger"
	"ovh-terminal/internal/ui/common"
	"ovh-terminal/internal/ui/layout"
	"ovh-terminal/internal/ui/styles"

	tea "github.com/charmbracelet/bubbletea"
)

var layoutManager *layout.Manager

// HandleKeyMsg processes keyboard input messages
func HandleKeyMsg(model common.UIModel, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !model.IsReady() {
		return model, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return model, tea.Quit

	case "tab":
		model.ToggleActivePane()
		styles.UpdateBorderStyles(model.GetActivePane())
		return model, nil

	case "enter":
		return handleEnterKey(model)

	// Handle navigation keys
	case "up", "k", "down", "j", "g", "G":
		if model.GetActivePane() == "content" {
			return handleContentNavigation(model, msg)
		}
	}

	return model, nil
}

// HandleWindowSizeMsg processes window resize messages
func HandleWindowSizeMsg(model common.UIModel, msg tea.WindowSizeMsg) tea.Model {
	logger.Log.Debug("Window size message received",
		"width", msg.Width,
		"height", msg.Height)

	if layoutManager == nil {
		layoutManager = layout.NewManager(model)
	}

	model.SetSize(msg.Width, msg.Height)

	if !layoutManager.ValidateWindowSize(msg.Width, msg.Height) {
		model.SetStatusMessage("Window too small - please resize")
		return model
	}

	layoutManager.Update()
	return model
}

// handleContentNavigation handles navigation in the content pane
func handleContentNavigation(model common.UIModel, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	vp := model.GetViewport()

	switch msg.String() {
	case "up", "k":
		return model, model.UpdateViewport(tea.KeyMsg{Type: tea.KeyUp})
	case "down", "j":
		return model, model.UpdateViewport(tea.KeyMsg{Type: tea.KeyDown})
	case "g":
		// Simulate home key for top navigation
		vp.GotoTop()
		model.SetViewport(vp)
		model.SetStatusMessage("Navigated to top of content")
	case "G":
		// Simulate end key for bottom navigation
		vp.GotoBottom()
		model.SetViewport(vp)
		model.SetStatusMessage("Navigated to bottom of content")
	}

	return model, nil
}

// handleEnterKey processes enter key presses
func handleEnterKey(model common.UIModel) (tea.Model, tea.Cmd) {
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
		if layoutManager != nil {
			layoutManager.Update()
		}

		return model, nil
	}

	logger.Log.Debug("Selected item is not a MenuItem",
		"type", fmt.Sprintf("%T", selectedItem))

	return model, nil
}
