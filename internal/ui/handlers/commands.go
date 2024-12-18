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
	logger.Log.Debug("Handling command",
		"title", item.Title(),
		"type", item.GetType(),
		"indent", item.GetIndent())

	switch item.GetType() {
	case common.TypeHeader:
		if item.GetIndent() == 0 {
			return handleTopLevelHeader(model, item)
		}
		return handleNestedHeader(model, item)
	case common.TypeTreeItem, common.TypeTreeLastItem:
		return handleTreeCommand(model, item)
	case common.TypeNormal:
		if item.Title() == "Exit" {
			return nil
		}
	}
	return nil
}

// handleTopLevelHeader handles main menu headers (indent level 0)
func handleTopLevelHeader(model common.UIModel, item common.MenuItem) error {
	logger.Log.Debug("Starting handleTopLevelHeader",
		"item", item.Title(),
		"expanded", item.IsExpanded())

	// Get current list
	list := model.GetList()
	currentIndex := list.Index()
	headerTitle := item.Title()

	// Toggle current item
	model.ToggleItemExpanded(currentIndex)

	// Get fresh items after toggle
	items := list.Items()
	clickedExpanded := items[currentIndex].(common.MenuItem).IsExpanded()

	// If we're expanding this header, collapse others
	if clickedExpanded {
		for i, item := range items {
			if menuItem, ok := item.(common.MenuItem); ok {
				if i != currentIndex && menuItem.GetType() == common.TypeHeader &&
					menuItem.GetIndent() == 0 && menuItem.IsExpanded() {
					logger.Log.Debug("Collapsing other header",
						"index", i,
						"title", menuItem.Title())
					model.ToggleItemExpanded(i)
				}
			}
		}
	}

	// Update menu structure
	model.UpdateMenuItems()

	// Find our header in the new menu structure and select it
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

	model.SetStatusMessage(fmt.Sprintf("Menu %s %s", item.Title(),
		map[bool]string{true: "expanded", false: "collapsed"}[clickedExpanded]))

	return nil
}

// handleNestedHeader handles nested headers (indent level > 0)
func handleNestedHeader(model common.UIModel, item common.MenuItem) error {
	logger.Log.Debug("Starting handleNestedHeader",
		"item", item.Title(),
		"expanded", item.IsExpanded())

	// Get current list
	list := model.GetList()
	currentIndex := list.Index()
	headerTitle := item.Title()

	// Toggle only this nested header
	model.ToggleItemExpanded(currentIndex)

	// Update menu structure
	model.UpdateMenuItems()

	// Find and select our header in the new structure
	items := list.Items()
	for i, menuItem := range items {
		if mi, ok := menuItem.(common.MenuItem); ok {
			if mi.Title() == headerTitle {
				logger.Log.Debug("Found nested header in new structure",
					"title", headerTitle,
					"newIndex", i)
				list.Select(i)
				break
			}
		}
	}

	model.SetStatusMessage(fmt.Sprintf("Section %s %s", item.Title(),
		map[bool]string{true: "expanded", false: "collapsed"}[!item.IsExpanded()]))

	return nil
}

// handleTreeCommand handles actions for regular tree items
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
