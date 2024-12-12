package ui

import (
	"fmt"
	"io"
	"strings"

	"ovh-terminal/internal/api"
	"ovh-terminal/internal/commands"
	"ovh-terminal/internal/logger"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base colors
	primaryColor        = lipgloss.Color("#7CE38B")
	secondaryColor      = lipgloss.Color("#5A5A5A")
	selectedFg          = lipgloss.Color("#FFFF22")
	selectedBg          = lipgloss.Color("#2D79C7")
	activeBorderColor   = lipgloss.Color("#888888")
	inactiveBorderColor = lipgloss.Color("#444444")

	// Document style (main container)
	docStyle = lipgloss.NewStyle().
			MarginLeft(1).
			MarginRight(1)

	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			PaddingLeft(1)

	// Menu styles
	menuStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(activeBorderColor).
			PaddingLeft(1).
			PaddingRight(1).
			Width(32)

	// Content styles
	contentStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(inactiveBorderColor).
			PaddingLeft(1).
			PaddingRight(1).
			MarginLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(selectedFg).
				Background(selectedBg)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	dimmedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// Status style
	statusStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(inactiveBorderColor).
			Padding(0, 0).
			MarginTop(0)
)

// types and functions related to the menu pane
type itemType int

const (
	typeNormal itemType = iota
	typeHeader
	typeSubHeader
	typeServerItem
	typeTreeItem
	typeTreeLastItem
)

// List item implementation
type listItem struct {
	title      string
	desc       string
	itemType   itemType
	expanded   bool
	indent     int
	selectable bool
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.desc }
func (i listItem) FilterValue() string { return i.title }

type itemDelegate struct {
	list.DefaultDelegate
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	li, ok := item.(listItem)
	if !ok {
		return
	}

	indent := strings.Repeat(" ", li.indent*1)
	var symbol string

	// Tree structure symbols
	var prefix string
	switch li.itemType {
	case typeTreeItem:
		prefix = "├─ "
	case typeTreeLastItem:
		prefix = "└─ "
	case typeHeader:
		if li.expanded {
			symbol = "[-] "
		} else {
			symbol = "[+] "
		}
	}

	title := indent + prefix + symbol + li.Title()
	style := normalItemStyle
	if index == m.Index() {
		style = selectedItemStyle
	}

	fmt.Fprint(w, style.Render(title))
}

type Model struct {
	list          list.Model
	viewport      viewport.Model
	content       string
	ready         bool
	width, height int
	statusMessage string
	apiClient     *api.Client
	activeCommand commands.Command
	activePane    string   // "menu" or "content"
	serverList    []string // Cache for server hostnames
}

// CommandHandler is a function type that creates commands
type CommandHandler func(*api.Client) commands.Command

// Initialize creates a new model with initial state
func Initialize(client *api.Client) Model {
	logger.Log.Debug("Initializing model")

	items := []list.Item{
		listItem{
			title:      "Account Information",
			itemType:   typeHeader,
			expanded:   false,
			selectable: true,
		},
		listItem{
			title:      "Bare Metal Cloud",
			itemType:   typeHeader,
			expanded:   false,
			selectable: true,
		},
		listItem{
			title:      "Web Cloud",
			itemType:   typeHeader,
			expanded:   false,
			selectable: true,
		},
		listItem{
			title:      "Exit",
			desc:       "Exit the application",
			itemType:   typeNormal,
			selectable: true,
		},
	}

	logger.Log.Debug("Creating initial menu items",
		"count", len(items))

	// Create custom delegate
	delegate := itemDelegate{
		DefaultDelegate: list.DefaultDelegate{
			ShowDescription: false,
		},
	}

	// Style the list items
	delegate.Styles.SelectedTitle = selectedItemStyle
	delegate.Styles.SelectedDesc = dimmedStyle
	delegate.Styles.NormalTitle = normalItemStyle
	delegate.Styles.NormalDesc = dimmedStyle.Copy()

	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		UnsetPadding().
		UnsetMargins()

	delegate.Styles.NormalTitle = delegate.Styles.SelectedTitle.
		UnsetPadding().
		UnsetMargins()

	// Create the list
	l := list.New(items, delegate, 0, 0)
	l.SetShowTitle(true)
	l.Title = "OVH Terminal Client"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.DisableQuitKeybindings()

	logger.Log.Debug("List initialized",
		"title", l.Title,
		"initial_size", fmt.Sprintf("%dx%d", l.Width(), l.Height()))

	// Create the viewport
	vp := viewport.New(0, 0)
	vp.Style = contentStyle

	logger.Log.Debug("Viewport initialized",
		"initial_size", fmt.Sprintf("%dx%d", vp.Width, vp.Height))

	// Configure logger
	logger.Log.Configure("debug", "logs/ovh-terminal.log", false)

	model := Model{
		list:       l,
		viewport:   vp,
		content:    "Welcome to OVH Terminal Client!\nUse arrow keys to navigate and Enter to select an option.",
		apiClient:  client,
		activePane: "menu", // Start with menu active
	}

	logger.Log.Debug("Model initialization complete",
		"active_pane", model.activePane,
		"ready", model.ready)

	return model
}

// Map of available commands
var commandHandlers = map[string]CommandHandler{
	"Account Information": func(client *api.Client) commands.Command {
		return commands.NewMeCommand(client)
	},
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

// Add a helper function to update the layout
func (m *Model) updateLayout() {
	logger.Log.Debug("Starting layout update",
		"ready", m.ready,
		"window_size", fmt.Sprintf("%dx%d", m.width, m.height))

	if !m.ready {
		return
	}

	// Calculate space needed for status bar including borders
	statusBarSpace := 3

	// Calculate space needed for title and borders
	uiElementsSpace := 4 // title + borders + extra space

	// Calculate available height for content
	// Window height minus:
	// - statusBarSpace (3)
	// - uiElementsSpace (6)
	availableContentHeight := m.height - statusBarSpace - uiElementsSpace

	logger.Log.Debug("Height space calculations",
		"window_height", m.height,
		"status_space", statusBarSpace,
		"ui_elements", uiElementsSpace,
		"available_content", availableContentHeight)

	// Calculate minimum needed height for menu items
	var minimumContentHeight int
	for _, item := range m.list.Items() {
		if i, ok := item.(listItem); ok {
			minimumContentHeight++
			logger.Log.Debug("Counting menu item",
				"title", i.title,
				"type", i.itemType,
				"expanded", i.expanded)

			if i.expanded {
				switch i.title {
				case "Account Information", "Bare Metal Cloud", "Web Cloud":
					minimumContentHeight += 2
					logger.Log.Debug("Added space for children",
						"parent", i.title,
						"extra_height", 2)
				}
			}
		}
	}

	// Use the minimum required height if available height is not enough
	contentHeight := availableContentHeight
	if contentHeight < minimumContentHeight {
		contentHeight = minimumContentHeight
		logger.Log.Debug("Using minimum content height instead of available",
			"available", availableContentHeight,
			"minimum_needed", minimumContentHeight)
	}

	// Calculate widths
	// Total horizontal space:
	// 2 (doc margins)
	// + 2 (menu borders)
	// + 2 (menu padding)
	// + 2 (content margin)
	// + 2 (content borders)
	// + 2 (content padding)
	// - 1 (extra space, dunno)
	horizontalSpace := 11
	menuBaseWidth := 32
	contentWidth := m.width - menuBaseWidth - horizontalSpace

	logger.Log.Debug("Width calculations",
		"window_width", m.width,
		"menu_base", menuBaseWidth,
		"horizontal_space", horizontalSpace,
		"content_width", contentWidth)

	// Set component sizes
	m.list.SetSize(menuBaseWidth, contentHeight)
	m.viewport.Width = contentWidth
	m.viewport.Height = contentHeight

	// Update content if needed
	if m.content != "" {
		m.viewport.SetContent(m.content)
		logger.Log.Debug("Updated viewport content", "content_length", len(m.content))
	}

	logger.Log.Debug("Layout update complete",
		"menu_size", fmt.Sprintf("%dx%d", menuBaseWidth, contentHeight),
		"viewport_size", fmt.Sprintf("%dx%d", contentWidth, contentHeight))
}

func (m *Model) updateMenuItems() {
	var updatedItems []list.Item

	// Get current items to check their state
	currentItems := m.list.Items()

	// Build new list
	for _, item := range currentItems {
		curr := item.(listItem)

		if curr.indent == 0 {
			updatedItems = append(updatedItems, curr)

			// If this is an expanded header, add its children
			if curr.itemType == typeHeader && curr.expanded {
				switch curr.title {
				case "Account Information":
					updatedItems = append(updatedItems,
						listItem{
							title:      "My information",
							desc:       "View and manage my current information",
							itemType:   typeTreeItem,
							indent:     1,
							selectable: true,
						},
						listItem{
							title:      "API information",
							desc:       "Information about applications and credentials",
							itemType:   typeTreeLastItem,
							indent:     1,
							selectable: true,
						},
					)
				case "Bare Metal Cloud":
					updatedItems = append(updatedItems,
						listItem{
							title:      "Dedicated Servers",
							desc:       "View and manage servers",
							itemType:   typeTreeItem,
							indent:     1,
							selectable: true,
						},
						listItem{
							title:      "Virtual Private Servers",
							desc:       "",
							itemType:   typeTreeLastItem,
							indent:     1,
							selectable: true,
						},
					)
				case "Web Cloud":
					updatedItems = append(updatedItems,
						listItem{
							title:      "Domain names",
							desc:       "View and manage domain names",
							itemType:   typeTreeItem,
							indent:     1,
							selectable: true,
						},
						listItem{
							title:      "Hosting plans",
							desc:       "",
							itemType:   typeTreeLastItem,
							indent:     1,
							selectable: true,
						},
					)
				}
			}
		}
	}
	m.list.SetItems(updatedItems)
}

func (m *Model) updateBorderStyles() {
	if m.activePane == "menu" {
		menuStyle = menuStyle.BorderForeground(activeBorderColor)
		contentStyle = contentStyle.BorderForeground(inactiveBorderColor)
	} else {
		menuStyle = menuStyle.BorderForeground(inactiveBorderColor)
		contentStyle = contentStyle.BorderForeground(activeBorderColor)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.ready {
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			// Toggle between panes
			if m.activePane == "menu" {
				m.activePane = "content"
			} else {
				m.activePane = "menu"
			}
			m.updateBorderStyles()
			return m, nil

		case "up", "k", "down", "j", "g", "G":
			// Only handle navigation keys for active pane
			if m.activePane == "menu" {
				var cmd tea.Cmd
				switch msg.String() {
				case "g": // Go to top
					m.list.Select(0)
					m.statusMessage = "Navigated to top of menu"
					return m, nil
				case "G": // Go to bottom
					m.list.Select(len(m.list.Items()) - 1)
					m.statusMessage = "Navigated to bottom of menu"
					return m, nil
				default:
					m.list, cmd = m.list.Update(msg)
					return m, cmd
				}
			} else if m.activePane == "content" {
				var cmd tea.Cmd
				switch msg.String() {
				case "g": // Go to top
					m.viewport.GotoTop()
					m.statusMessage = "Navigated to top of content"
					return m, nil
				case "G": // Go to bottom
					m.viewport.GotoBottom()
					m.statusMessage = "Navigated to bottom of content"
					return m, nil
				default:
					m.viewport, cmd = m.viewport.Update(msg)
					return m, cmd
				}
			}

		case "enter":
			// Only handle enter when menu is active
			if m.activePane == "menu" {
				// Get the current list and selected index
				selectedItem := m.list.SelectedItem().(listItem)

				switch selectedItem.itemType {
				case typeHeader:
					// Toggle the expanded state
					newExpanded := !selectedItem.expanded

					// Create new base list with updated state
					var updatedItems []list.Item
					for _, item := range m.list.Items() {
						if curr, ok := item.(listItem); ok {
							if curr.title == selectedItem.title && curr.itemType == typeHeader {
								curr.expanded = newExpanded
							}
							// Only add if it's a header or non-child item
							if curr.indent == 0 {
								updatedItems = append(updatedItems, curr)
							}
						}
					}
					// Set the updated base list
					m.list.SetItems(updatedItems)

					// Update menu to show/hide children
					m.updateMenuItems()

					// Update layout to handle size changes
					m.updateLayout()

					m.statusMessage = fmt.Sprintf("Menu %s %s", selectedItem.title,
						map[bool]string{true: "expanded", false: "collapsed"}[newExpanded])

				case typeTreeItem, typeTreeLastItem:
					// Execute commands for specific items
					if selectedItem.title == "My information" {
						if handler, exists := commandHandlers["Account Information"]; exists {
							m.activeCommand = handler(m.apiClient)
							if output, err := m.activeCommand.Execute(); err != nil {
								m.statusMessage = fmt.Sprintf("Error: %v", err)
								m.viewport.SetContent(fmt.Sprintf("Failed to execute command: %v", err))
							} else {
								m.statusMessage = fmt.Sprintf("Executed: %s", selectedItem.title)
								m.viewport.SetContent(output)
							}
							m.activePane = "content"
							m.updateLayout()
						}
					} else {
						m.statusMessage = fmt.Sprintf("Selected: %s", selectedItem.title)
					}

				case typeNormal:
					if selectedItem.title == "Exit" {
						return m, tea.Quit
					}
				}
			}
		}
	case tea.WindowSizeMsg:
		logger.Log.Debug("Window size message received",
			"width", msg.Width,
			"height", msg.Height)

		m.width = msg.Width
		m.height = msg.Height

		if m.width < 80 || m.height < 20 {
			m.statusMessage = "Window too small - please resize"
			m.ready = false
			logger.Log.Debug("Window too small",
				"width", m.width,
				"height", m.height,
				"minimum_width", 80,
				"minimum_height", 20)
			return m, nil
		}

		// Set ready flag and update layout
		m.ready = true
		m.updateLayout()

		logger.Log.Debug("Window size handled",
			"ready", m.ready,
			"viewport_content", m.content != "")
	}

	// Update list and viewport
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		logger.Log.Debug("View called but not ready")
		return "\n  Initializing... (resize window if needed)"
	}

	logger.Log.Debug("Starting view render",
		"active_pane", m.activePane)

	m.list.SetFilteringEnabled(false)

	// Render menu and content
	menuView := menuStyle.Render(m.list.View())
	contentView := contentStyle.Render(m.viewport.View())

	logger.Log.Debug("Component dimensions",
		"menu_height", lipgloss.Height(menuView),
		"menu_width", lipgloss.Width(menuView),
		"content_height", lipgloss.Height(contentView),
		"content_width", lipgloss.Width(contentView))

	// Combine menu and content horizontally
	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		menuView,
		contentView,
	)

	logger.Log.Debug("Main view dimensions",
		"height", lipgloss.Height(mainView),
		"width", lipgloss.Width(mainView))

	// Always show status bar, with default text if no message
	statusText := m.statusMessage
	if statusText == "" {
		// Show help text in status bar when no message
		if m.activePane == "menu" {
			statusText = "↑/k up • ↓/j down • g/G top/bottom • ? help"
		} else {
			statusText = "↑/k up • ↓/j down • g/G top/bottom • Tab to menu"
		}
	}

	// Calculate status bar width:
	// - Take main view width
	// - Subtract 4 for the status bar borders and padding
	mainViewWidth := lipgloss.Width(mainView)
	statusBarWidth := mainViewWidth - 2
	statusStyle = statusStyle.Width(statusBarWidth)

	logger.Log.Debug("Status bar setup",
		"text_length", len(statusText),
		"main_view_width", mainViewWidth,
		"status_bar_width", statusBarWidth)

	// Render final view
	finalView := docStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			mainView,
			statusStyle.Render(statusText),
		),
	)

	logger.Log.Debug("Final view dimensions",
		"height", lipgloss.Height(finalView),
		"width", lipgloss.Width(finalView))

	return finalView
}

type statusMsg string

func (m Model) expandServers() tea.Cmd {
	return func() tea.Msg {
		currentItems := m.list.Items()
		var newItems []list.Item

		for _, item := range currentItems {
			li, ok := item.(listItem)
			if !ok {
				continue
			}
			newItems = append(newItems, li)

			// Als dit het Dedicated Servers item is
			if li.itemType == typeSubHeader && li.title == "Dedicated Servers" {
				// Toggle expanded state
				li.expanded = !li.expanded
				newItems[len(newItems)-1] = li

				if li.expanded {
					// Fetch servers if expanded
					servers, err := m.fetchServers()
					if err != nil {
						return statusMsg(fmt.Sprintf("Error loading servers: %v", err))
					}

					// Add server items as indented items
					for _, server := range servers {
						newItems = append(newItems, listItem{
							title:    server,
							desc:     "Dedicated server",
							itemType: typeServerItem,
							indent:   2,
						})
					}
				}
				// Als het item wordt ingeklapt, verwijderen we gewoon geen items
				// de nieuwe lijst bevat dan automatisch geen child items meer
			}
		}

		// Update the list
		m.list.SetItems(newItems)
		return statusMsg("Servers loaded")
	}
}

func (m *Model) fetchServers() ([]string, error) {
	if err := m.apiClient.Get("/dedicated/server", &m.serverList); err != nil {
		return nil, fmt.Errorf("failed to fetch servers: %w", err)
	}
	return m.serverList, nil
}
