// internal/ui/types/model.go
package types

import (
	"sort"

	"ovh-terminal/internal/api"
	"ovh-terminal/internal/commands"
	"ovh-terminal/internal/logger"
	"ovh-terminal/internal/ui/common"
	"ovh-terminal/internal/ui/handlers"
	"ovh-terminal/internal/ui/help"
	"ovh-terminal/internal/ui/styles"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the application UI state
type Model struct {
	// Core components
	List          list.Model
	Viewport      viewport.Model
	apiClient     *api.Client
	ActiveCommand commands.Command

	// Content state
	Content       string
	StatusMessage string
	ServerList    []string

	// UI state
	Ready      bool
	ActivePane string
	Width      int
	Height     int

	ShowHelp bool
}

// Ensure Model implements common.UIModel
var _ common.UIModel = (*Model)(nil)

// UIModel interface implementation
func (m *Model) GetAPIClient() *api.Client {
	return m.apiClient
}

func (m *Model) SetAPIClient(client *api.Client) {
	m.apiClient = client
}

func (m *Model) GetActivePane() string {
	return m.ActivePane
}

func (m *Model) ToggleActivePane() {
	if m.ActivePane == "menu" {
		m.ActivePane = "content"
	} else {
		m.ActivePane = "menu"
	}
}

func (m *Model) SetSize(width, height int) {
	m.Width = width
	m.Height = height
	m.Ready = width >= 80 && height >= 20
}

func (m *Model) SetContent(content string) {
	m.Content = content
	if m.Viewport.Width > 0 {
		m.Viewport.SetContent(content)
	}
}

func (m *Model) SetStatusMessage(msg string) {
	m.StatusMessage = msg
}

func (m *Model) IsReady() bool {
	return m.Ready
}

func (m *Model) GetWidth() int {
	return m.Width
}

func (m *Model) GetHeight() int {
	return m.Height
}

func (m *Model) GetList() *list.Model {
	return &m.List
}

func (m *Model) UpdateList(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return cmd
}

func (m *Model) SetList(list *list.Model) {
	m.List = *list
}

func (m *Model) GetViewport() *viewport.Model {
	return &m.Viewport
}

func (m *Model) UpdateViewport(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.Viewport, cmd = m.Viewport.Update(msg)
	return cmd
}

func (m *Model) SetViewport(vp *viewport.Model) {
	m.Viewport = *vp
}

// UpdateMenuItems refreshes all menu items while preserving states
func (m *Model) UpdateMenuItems() {
	var updatedItems []list.Item
	currentItems := m.List.Items()

	// Helper to add child items for a header
	addChildItems := func(items []*ListItem) {
		for i, item := range items {
			itemType := common.TypeTreeItem
			if i == len(items)-1 {
				itemType = common.TypeTreeLastItem
			}

			newItem := NewListItem(
				item.Title(),
				itemType,
				WithDesc(item.Description()),
				WithIndent(item.GetIndent()),
				WithSelectable(item.IsSelectable()),
			)
			updatedItems = append(updatedItems, newItem)
		}
	}

	// Build new list preserving expanded states
	for _, item := range currentItems {
		curr, ok := item.(*ListItem)
		if !ok {
			continue
		}

		if curr.GetIndent() == 0 {
			updatedItems = append(updatedItems, curr)

			if curr.GetType() == common.TypeHeader && curr.IsExpanded() {
				switch curr.Title() {
				case "Account Information":
					addChildItems([]*ListItem{
						NewListItem("My information", common.TypeTreeItem,
							WithDesc("View and manage my current information"),
							WithIndent(1)),
						NewListItem("API information", common.TypeTreeLastItem,
							WithDesc("Information about applications and credentials"),
							WithIndent(1)),
					})

				case "Bare Metal Cloud":
					// Find current states
					var isDedServersExpanded bool
					var dedServersItem *ListItem
					for _, oldItem := range currentItems {
						if old, ok := oldItem.(*ListItem); ok {
							if old.GetIndent() == 1 && old.Title() == "Dedicated Servers" {
								isDedServersExpanded = old.IsExpanded()
								dedServersItem = old
								break
							}
						}
					}

					// Add Dedicated Servers header
					if dedServersItem == nil {
						dedServersItem = NewListItem("Dedicated Servers", common.TypeHeader,
							WithDesc("View and manage servers"),
							WithIndent(1),
							WithExpanded(isDedServersExpanded))
					}
					updatedItems = append(updatedItems, dedServersItem)

					// If Dedicated Servers is expanded, add servers
					if isDedServersExpanded {
						// Get server list via command
						cmd := commands.NewServerCommand(m.apiClient)
						servers, err := cmd.ListServers()
						if err != nil {
							updatedItems = append(updatedItems,
								NewListItem("Error loading servers", common.TypeTreeLastItem,
									WithDesc(err.Error()),
									WithIndent(2)))
						} else {
							// Convert map to sorted slice
							type serverInfo struct {
								name string
								id   string
							}
							serverList := make([]serverInfo, 0, len(servers))
							for id, name := range servers {
								serverList = append(serverList, serverInfo{name, id})
							}
							// Sort servers by name
							sort.Slice(serverList, func(i, j int) bool {
								return serverList[i].name < serverList[j].name
							})

							// Add servers as menu items
							for i, server := range serverList {
								itemType := common.TypeTreeItem
								if i == len(serverList)-1 {
									itemType = common.TypeTreeLastItem
								}
								updatedItems = append(updatedItems,
									NewListItem(server.name, itemType,
										WithDesc(server.id),
										WithIndent(2)))
							}
						}
					}

					// Add Virtual Private Servers with same expansion logic as Dedicated Servers
					var isVPSExpanded bool
					var vpsItem *ListItem
					for _, oldItem := range currentItems {
						if old, ok := oldItem.(*ListItem); ok {
							if old.GetIndent() == 1 && old.Title() == "Virtual Private Servers" {
								isVPSExpanded = old.IsExpanded()
								vpsItem = old
								break
							}
						}
					}

					// Add VPS header
					if vpsItem == nil {
						vpsItem = NewListItem("Virtual Private Servers", common.TypeHeader,
							WithDesc("Virtual Private Servers"),
							WithIndent(1),
							WithExpanded(isVPSExpanded))
					}
					updatedItems = append(updatedItems, vpsItem)

					// If VPS section is expanded, add VPS instances
					if isVPSExpanded {
						// Get VPS list via API
						vpsServers, err := m.apiClient.ListVPS()
						if err != nil {
							updatedItems = append(updatedItems,
								NewListItem("Error loading VPS instances", common.TypeTreeLastItem,
									WithDesc(err.Error()),
									WithIndent(2)))
						} else {
							// Convert to sorted slice with display names
							type vpsInfo struct {
								name string
								id   string
							}
							vpsList := make([]vpsInfo, 0, len(vpsServers))
							for _, id := range vpsServers {
								info, err := m.apiClient.GetVPSInfo(id)
								if err != nil {
									logger.Log.Error("Failed to get VPS info",
										"id", id,
										"error", err)
									continue
								}
								vpsList = append(vpsList, vpsInfo{
									name: info.GetDisplayTitle(),
									id:   id,
								})
							}
							// Sort VPS instances by name
							sort.Slice(vpsList, func(i, j int) bool {
								return vpsList[i].name < vpsList[j].name
							})

							// Add VPS instances as menu items
							for i, vps := range vpsList {
								itemType := common.TypeTreeItem
								if i == len(vpsList)-1 {
									itemType = common.TypeTreeLastItem
								}
								updatedItems = append(updatedItems,
									NewListItem(vps.name, itemType,
										WithDesc(vps.id),
										WithIndent(2)))
							}
						}
					}

				case "Web Cloud":
					addChildItems([]*ListItem{
						NewListItem("Domain names", common.TypeTreeItem,
							WithDesc("View and manage domain names"),
							WithIndent(1)),
						NewListItem("Hosting plans", common.TypeTreeLastItem,
							WithDesc(""),
							WithIndent(1)),
					})
				}
			}
		}
	}

	// Preserve current selection if possible
	currentIndex := m.List.Index()
	m.List.SetItems(updatedItems)
	if currentIndex < len(updatedItems) {
		m.List.Select(currentIndex)
	}
}

// Tea.Model implementation
func (m *Model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		_, cmd := handlers.HandleKeyMsg(m, msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		handlers.HandleWindowSizeMsg(m, msg)
	}

	// Update active component
	var cmd tea.Cmd
	if m.ActivePane == "menu" {
		m.List, cmd = m.List.Update(msg)
	} else {
		m.Viewport, cmd = m.Viewport.Update(msg)
	}
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if !m.IsReady() {
		return "\n  Initializing... (resize window if needed)"
	}

	// Disable filtering which we don't use
	m.List.SetFilteringEnabled(false)

	// Render menu and content
	menuView := styles.MenuStyle.Render(m.List.View())
	contentView := styles.ContentStyle.Render(m.Viewport.View())

	// Combine menu and content horizontally
	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		menuView,
		contentView,
	)

	// Get status text based on current state
	statusText := m.StatusMessage
	if statusText == "" {
		if m.GetActivePane() == "menu" {
			statusText = "↑/k up • ↓/j down • g/G top/bottom • ? help"
		} else {
			statusText = "↑/k up • ↓/j down • g/G top/bottom • Tab to menu"
		}
	}

	// Calculate status bar width
	mainViewWidth := lipgloss.Width(mainView)
	statusBarWidth := mainViewWidth - 2
	statusStyle := styles.StatusStyle.Width(statusBarWidth)

	// Render final view
	finalView := styles.DocStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			mainView,
			statusStyle.Render(statusText),
		),
	)

	// If help is enabled, overlay the help content
	if m.ShowHelp {
		return help.GetHelpContent(m.Width, m.Height)
	}

	return finalView
}

func (m *Model) ToggleHelp() {
	m.ShowHelp = !m.ShowHelp
	if m.ShowHelp {
		m.SetStatusMessage("Showing help (press F1 to close)")
	} else {
		m.SetStatusMessage("Help closed")
	}
}

// ToggleItemExpanded toggles the expanded state of a menu item
func (m *Model) ToggleItemExpanded(index int) {
	items := m.List.Items()
	if index < 0 || index >= len(items) {
		return
	}

	if listItem, ok := items[index].(*ListItem); ok {
		// Create a new item with toggled expanded state
		newItem := listItem.WithExpanded(!listItem.IsExpanded())

		// Update the item in the list
		newItems := make([]list.Item, len(items))
		copy(newItems, items)
		newItems[index] = newItem
		m.List.SetItems(newItems)
	}
}

// NewModel creates a new Model instance
func NewModel() *Model {
	return &Model{
		ActivePane: "menu",
		ShowHelp:   false,
	}
}
