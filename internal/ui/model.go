// internal/ui/model.go
// Package ui provides the terminal user interface
package ui

import (
	"fmt"

	"ovh-terminal/internal/api"
	"ovh-terminal/internal/logger"
	"ovh-terminal/internal/ui/layout"
	"ovh-terminal/internal/ui/styles"
	"ovh-terminal/internal/ui/types"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// Initialize creates a new model with initial state
func Initialize(client *api.Client) *types.Model {
	// Configure logger
	if err := logger.Log.Configure("debug", "logs/ovh-terminal.log", false); err != nil {
		// Since we're in Initialize, we can only log to stdout
		fmt.Printf("Failed to configure logger: %v\n", err)
	}

	logger.Log.Debug("Initializing model")

	// Create initial model
	model := types.NewModel()
	model.SetAPIClient(client)

	// Create initial menu list
	items := types.CreateBaseMenuItems()
	logger.Log.Debug("Created initial menu items", "count", len(items))

	// Create custom delegate
	delegate := types.NewItemDelegate()

	// Create and configure the list
	list := list.New(items, delegate, 0, 0)
	list.SetShowTitle(true)
	list.Title = "OVH Terminal Client"
	list.SetShowStatusBar(false)
	list.SetFilteringEnabled(false)
	list.Styles.Title = styles.TitleStyle
	list.DisableQuitKeybindings()

	model.List = list

	// Initialize viewport
	vp := viewport.New(0, 0)
	model.Viewport = vp

	// Set initial content
	welcomeMsg := "Welcome to OVH Terminal Client!\n\n" +
		"Use arrow keys to navigate and Enter to select an option."
	model.SetContent(lipgloss.NewStyle().Render(welcomeMsg))

	// Create layout manager and do initial layout
	layoutMgr := layout.NewManager(model)
	layoutMgr.Update()

	logger.Log.Debug("Model initialization complete",
		"active_pane", model.GetActivePane(),
		"ready", model.IsReady())

	return model
}
