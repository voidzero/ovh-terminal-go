// internal/ui/types/model.go
// Package types provides concrete implementations of UI types
package types

import (
	"ovh-terminal/internal/api"
	"ovh-terminal/internal/commands"
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

// GetWidth returns the model's width
func (m *Model) GetWidth() int {
	return m.Width
}

// GetHeight returns the model's height
func (m *Model) GetHeight() int {
	return m.Height
}

// Additional UIModel interface methods
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

func (m *Model) UpdateMenuItems() {
	var updatedItems []list.Item
	currentItems := m.List.Items()

	// Build new list preserving expanded states
	for _, item := range currentItems {
		curr := item.(ListItem)

		if curr.GetIndent() == 0 {
			updatedItems = append(updatedItems, curr)

			// If this is an expanded header, add its children
			if curr.GetType() == common.TypeHeader && curr.IsExpanded() {
				switch curr.Title() {
				case "Account Information":
					updatedItems = append(updatedItems,
						NewListItem("My information", common.TypeTreeItem,
							WithDesc("View and manage my current information"),
							WithIndent(1)),
						NewListItem("API information", common.TypeTreeLastItem,
							WithDesc("Information about applications and credentials"),
							WithIndent(1)))
				case "Bare Metal Cloud":
					updatedItems = append(updatedItems,
						NewListItem("Dedicated Servers", common.TypeTreeItem,
							WithDesc("View and manage servers"),
							WithIndent(1)),
						NewListItem("Virtual Private Servers", common.TypeTreeLastItem,
							WithDesc(""),
							WithIndent(1)))
				case "Web Cloud":
					updatedItems = append(updatedItems,
						NewListItem("Domain names", common.TypeTreeItem,
							WithDesc("View and manage domain names"),
							WithIndent(1)),
						NewListItem("Hosting plans", common.TypeTreeLastItem,
							WithDesc(""),
							WithIndent(1)))
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

// Tea.Model implementation

func (m *Model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

// Update implements tea.Model
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

// ToggleHelp toggles the help overlay visibility
func (m *Model) ToggleHelp() {
	m.ShowHelp = !m.ShowHelp
	if m.ShowHelp {
		m.SetStatusMessage("Showing help (press F1 to close)")
	} else {
		m.SetStatusMessage("Help closed")
	}
}

// NewModel creates a new Model instance
func NewModel() *Model {
	return &Model{
		ActivePane: "menu",
		ShowHelp:   false,
	}
}

// ToggleItemExpanded toggles the expanded state of a menu item
func (m *Model) ToggleItemExpanded(index int) {
	items := m.List.Items()
	if index < 0 || index >= len(items) {
		return
	}

	if listItem, ok := items[index].(ListItem); ok {
		// Create a new item with toggled expanded state
		newItem := listItem.WithExpanded(!listItem.IsExpanded())

		// Update the item in the list
		newItems := make([]list.Item, len(items))
		copy(newItems, items)
		newItems[index] = newItem
		m.List.SetItems(newItems)
	}
}
