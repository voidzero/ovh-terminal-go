// internal/ui/handlers/commands.go
package handlers

import (
	"fmt"

	"ovh-terminal/internal/api"
	"ovh-terminal/internal/commands"
	"ovh-terminal/internal/logger"
	"ovh-terminal/internal/ui/common"
	"ovh-terminal/internal/ui/styles"
)

// CommandHandler is a function type that creates commands
type CommandHandler func(*api.Client) commands.Command

// commandRegistry maps menu items to their command handlers
var commandRegistry = map[string]CommandHandler{
	"My information": func(client *api.Client) commands.Command {
		return commands.NewMeCommand(client)
	},
	"API information": func(client *api.Client) commands.Command {
		return commands.NewAPIInfoCommand(client)
	},
}

// HandleCommand processes a selected menu item and executes any associated command
func HandleCommand(model common.UIModel, item common.MenuItem) error {
	// Handle commands based on item type
	switch item.GetType() {
	case common.TypeHeader:
		return handleHeaderCommand(model, item)
	case common.TypeTreeItem, common.TypeTreeLastItem:
		return handleTreeCommand(model, item)
	case common.TypeNormal:
		if item.Title() == "Exit" {
			return nil
		}
	}
	return nil
}

// handleHeaderCommand handles collapsible header items
func handleHeaderCommand(model common.UIModel, item common.MenuItem) error {
	logger.Log.Debug("Starting handleHeaderCommand",
		"item", item.Title(),
		"type", item.GetType(),
		"expanded", item.IsExpanded())

	// Get current list
	list := model.GetList()
	currentIndex := list.Index()
	headerTitle := item.Title() // Remember which header we're working with

	// Toggle current item
	model.ToggleItemExpanded(currentIndex)

	// Get fresh items after toggle
	items := list.Items()
	clickedExpanded := items[currentIndex].(common.MenuItem).IsExpanded()

	// If we're expanding this header, collapse all others
	if clickedExpanded {
		for i, item := range items {
			if menuItem, ok := item.(common.MenuItem); ok {
				// Skip current item and non-headers
				if i != currentIndex && menuItem.GetType() == common.TypeHeader &&
					menuItem.IsExpanded() {
					logger.Log.Debug("Collapsing other header",
						"index", i,
						"title", menuItem.Title())
					// Collapse this header
					model.ToggleItemExpanded(i)
				}
			}
		}
	}

	// Update the menu structure
	model.UpdateMenuItems()

	// Find our header in the new menu structure
	items = list.Items()
	for i, menuItem := range items {
		if mi, ok := menuItem.(common.MenuItem); ok {
			if mi.Title() == headerTitle {
				logger.Log.Debug("Found header in new structure",
					"title", headerTitle,
					"newIndex", i)
				list.Select(i)
				break
			}
		}
	}

	// Update status message
	model.SetStatusMessage(fmt.Sprintf("Menu %s %s", item.Title(),
		map[bool]string{true: "expanded", false: "collapsed"}[clickedExpanded]))

	return nil
}

// handleTreeCommand handles tree item commands
func handleTreeCommand(model common.UIModel, item common.MenuItem) error {
	handler, exists := commandRegistry[item.Title()]
	if !exists {
		model.SetStatusMessage(fmt.Sprintf("Selected: %s", item.Title()))
		return nil
	}

	// Create and execute command
	cmd := handler(model.GetAPIClient())
	output, err := cmd.Execute()
	if err != nil {
		model.SetStatusMessage(fmt.Sprintf("Error: %v", err))
		model.SetContent(fmt.Sprintf("Failed to execute command: %v", err))
		return err
	}

	// Update UI with command output
	model.SetStatusMessage(fmt.Sprintf("Executed: %s", item.Title()))
	model.SetContent(output)

	// Switch to content pane to show output
	model.ToggleActivePane()

	// Update border colors to reflect the active pane
	styles.UpdateBorderStyles(model.GetActivePane())

	return nil
}
