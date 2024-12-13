// internal/ui/handlers/commands.go
package handlers

import (
	"fmt"

	"ovh-terminal/internal/api"
	"ovh-terminal/internal/commands"
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
	// Get current list
	list := model.GetList()

	// First toggle the expanded state
	currentIndex := list.Index()
	model.ToggleItemExpanded(currentIndex)

	// Update menu items to show/hide children
	model.UpdateMenuItems()

	// Update status message
	model.SetStatusMessage(fmt.Sprintf("Menu %s %s", item.Title(),
		map[bool]string{true: "expanded", false: "collapsed"}[item.IsExpanded()]))

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
