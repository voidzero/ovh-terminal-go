package ui

import (
	"fmt"

	"ovh-terminal/internal/api"
	"ovh-terminal/internal/commands"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base colors
	primaryColor   = lipgloss.Color("#7CE38B")
	secondaryColor = lipgloss.Color("#5A5A5A")
	selectedBg     = lipgloss.Color("#2D79C7")
	borderColor    = lipgloss.Color("#444444")

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
			BorderForeground(borderColor).
			PaddingLeft(1).
			PaddingRight(1).
			Width(32)

	// Content styles
	contentStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
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

type Model struct {
	list          list.Model
	viewport      viewport.Model
	content       string
	ready         bool
	width, height int
	statusMessage string
	apiClient     *api.Client
	activeCommand commands.Command
}

// CommandHandler is a function type that creates commands
type CommandHandler func(*api.Client) commands.Command

// Initialize creates a new model with initial state
func Initialize(client *api.Client) Model {
	items := []list.Item{
		listItem{title: "Me", desc: "Account information"},
		listItem{title: "Servers", desc: "Manage dedicated servers"},
		listItem{title: "Domains", desc: "Domain management"},
		listItem{title: "Cloud Projects", desc: "Cloud project overview"},
		listItem{title: "IP Management", desc: "IP address management"},
		listItem{title: "Exit", desc: "Exit the application"},
	}

	// Create custom delegate
	delegate := list.NewDefaultDelegate()

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
		list:      l,
		viewport:  vp,
		content:   "Welcome to OVH Terminal Client!\nUse arrow keys to navigate and Enter to select an option.",
		apiClient: client,
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

		case "enter":
			if i := m.list.SelectedItem(); i != nil {
				item := i.(listItem)
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
					// Update layout after changing status
					m.updateLayout()
				} else {
					m.statusMessage = fmt.Sprintf("Command not implemented: %s", item.title)
					m.viewport.SetContent("This command is not implemented yet.")
				}
				// Update layout after changing status
				m.updateLayout()
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

// List item implementation
type listItem struct {
	title string
	desc  string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.desc }
func (i listItem) FilterValue() string { return i.title }
