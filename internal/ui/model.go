package ui

import (
	"fmt"
	"io"
	"strings"

	"ovh-terminal/internal/api"
	"ovh-terminal/internal/commands"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base colors
	primaryColor        = lipgloss.Color("#7CE38B")
	secondaryColor      = lipgloss.Color("#5A5A5A")
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
			PaddingLeft(2).
			PaddingRight(2).
			MarginLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("231")).
				Background(selectedBg)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	dimmedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// Status style
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingLeft(1).
			MarginTop(1)
)

// types and functions related to the menu pane
type itemType int

const (
	typeNormal itemType = iota
	typeHeader
	typeSubHeader
	typeServerItem
)

// List item implementation
type listItem struct {
	title    string
	desc     string
	itemType itemType
	expanded bool
	indent   int
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

	indent := strings.Repeat("  ", li.indent)
	title := indent + li.Title()

	var style lipgloss.Style
	if index == m.Index() {
		style = selectedItemStyle
	} else {
		style = normalItemStyle
	}

	if li.itemType == typeHeader {
		style = style.Bold(true)
	}

	fmt.Fprint(w, style.Render(title))
	if li.desc != "" {
		fmt.Fprint(w, "\n"+indent+"  "+dimmedStyle.Render(li.Description()))
	}
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
	items := []list.Item{
		listItem{
			title:    "Account Information",
			desc:     "",
			itemType: typeHeader,
			indent:   0,
		},
		listItem{
			title:    "Me",
			desc:     "View account details",
			itemType: typeNormal,
			indent:   1,
		},
		listItem{
			title:    "Bare Metal Cloud",
			desc:     "",
			itemType: typeHeader,
			indent:   0,
		},
		listItem{
			title:    "Dedicated Servers",
			desc:     "View and manage servers",
			itemType: typeSubHeader,
			expanded: false,
			indent:   1,
		},
		listItem{
			title:    "Exit",
			desc:     "Exit the application",
			itemType: typeNormal,
		},
	}

	// Create custom delegate
	delegate := itemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
	}

	// Style the list items
	delegate.Styles.SelectedTitle = selectedItemStyle
	delegate.Styles.SelectedDesc = dimmedStyle
	delegate.Styles.NormalTitle = normalItemStyle
	delegate.Styles.NormalDesc = dimmedStyle.Copy()

	// Create the list
	l := list.New(items, delegate, 0, 0)
	l.SetShowTitle(true)
	l.Title = "OVH Terminal Client"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.DisableQuitKeybindings()

	// Create the viewport
	vp := viewport.New(0, 0)
	vp.Style = contentStyle

	return Model{
		list:       l,
		viewport:   vp,
		content:    "Welcome to OVH Terminal Client!\nUse arrow keys to navigate and Enter to select an option.",
		apiClient:  client,
		activePane: "menu", // Start with menu active
	}
}

// Map of available commands
var commandHandlers = map[string]CommandHandler{
	"Me": func(client *api.Client) commands.Command {
		return commands.NewMeCommand(client)
	},
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

// Voeg een helper functie toe om de layout te updaten
func (m *Model) updateLayout() {
	if !m.ready {
		return
	}

	menuWidth := 34 // Fixed menu width including borders
	contentWidth := m.width - menuWidth - 14

	// Calculate vertical space
	verticalSpace := 4
	if m.statusMessage != "" {
		verticalSpace += 2
	}

	// Update dimensions
	m.list.SetSize(menuWidth-2, m.height-verticalSpace)
	m.viewport.Width = contentWidth
	m.viewport.Height = m.height - verticalSpace

	// Update border styles
	m.updateBorderStyles()
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

		case "up", "k", "down", "j":
			// Only handle navigation keys for active pane
			if m.activePane == "menu" {
				var cmd tea.Cmd
				m.list, cmd = m.list.Update(msg)
				return m, cmd
			} else if m.activePane == "content" {
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}

		case "enter":
			// Only handle enter when menu is active
			if m.activePane == "menu" {
				if i := m.list.SelectedItem(); i != nil {
					item := i.(listItem)

					switch item.itemType {
					case typeNormal:
						if item.title == "Exit" {
							return m, tea.Quit
						}

						// Execute command if available
						if handler, exists := commandHandlers[item.title]; exists {
							m.activeCommand = handler(m.apiClient)
							if output, err := m.activeCommand.Execute(); err != nil {
								m.statusMessage = fmt.Sprintf("Error: %v", err)
								m.viewport.SetContent(fmt.Sprintf("Failed to execute command: %v", err))
							} else {
								m.statusMessage = fmt.Sprintf("Executed: %s", item.title)
								m.viewport.SetContent(output)
							}
						} else {
							m.statusMessage = fmt.Sprintf("Command not implemented: %s", item.title)
							m.viewport.SetContent("This command is not implemented yet.")
						}
						// Switch to content pane after successful command
						m.activePane = "content"
						m.updateLayout()

					case typeSubHeader:
						// Handle expandable items
						if item.title == "Dedicated Servers" {
							return m, tea.Batch(
								m.expandServers(),
								func() tea.Msg { return statusMsg(fmt.Sprintf("Loading servers...")) },
							)
						}

					case typeServerItem:
						// When clicking a server, show its details
						m.statusMessage = fmt.Sprintf("Selected server: %s", item.title)
						// TODO: Implement server details command
						m.activePane = "content"
						m.updateLayout()
					}
				}
			}
		}

	// Update the content style to have a fixed width based on calculations
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if m.width < 80 || m.height < 20 {
			m.statusMessage = "Window too small - please resize"
			m.ready = false
			return m, nil
		}

		// Calculate dimensions with borders and margins in mind
		menuWidth := 34 // Fixed menu width including borders

		// Total horizontal space needed for borders and margins:
		// - 2 for left document margin
		// - 2 for menu borders
		// - 2 for menu padding
		// - 2 for content margin between menu and content
		// - 2 for content borders
		// - 4 for content padding (2 left, 2 right)
		// Total: 14 characters

		// Available width for content
		contentWidth := m.width - menuWidth - 14

		// Calculate vertical space needed:
		// - 4 for top/bottom borders
		// - 1 for status bar (if present)
		// - 1 for status bar margin
		verticalSpace := 4
		if m.statusMessage != "" {
			verticalSpace += 2
		}

		// Update menu dimensions (subtract borders)
		m.list.SetSize(menuWidth-2, m.height-verticalSpace)

		// Update content dimensions
		m.viewport.Width = contentWidth
		m.viewport.Height = m.height - verticalSpace

		if !m.ready {
			m.viewport.SetContent(m.content)
		}

		m.ready = true
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
		return "\n  Initializing... (resize window if needed)"
	}

	// Render menu and content
	menuView := menuStyle.Render(m.list.View())
	contentView := contentStyle.Render(m.viewport.View())

	// Combine horizontally
	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		menuView,
		contentView,
	)

	// Add status if present
	if m.statusMessage != "" {
		return docStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				mainView,
				statusStyle.Render(m.statusMessage),
			),
		)
	}

	return docStyle.Render(mainView)
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
