// internal/ui/common/types.go
package common

import (
	"ovh-terminal/internal/api"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// ItemType represents different types of menu items
type ItemType int

const (
	// TypeNormal represents a regular menu item
	TypeNormal ItemType = iota

	// TypeHeader represents a collapsible section header
	TypeHeader

	// TypeSubHeader represents a sub-section header
	TypeSubHeader

	// TypeServerItem represents a server in the list
	TypeServerItem

	// TypeTreeItem represents a tree node with siblings
	TypeTreeItem

	// TypeTreeLastItem represents the last item in a tree branch
	TypeTreeLastItem
)

// String provides human-readable names for ItemTypes
func (it ItemType) String() string {
	return map[ItemType]string{
		TypeNormal:       "Normal Item",
		TypeHeader:       "Header",
		TypeSubHeader:    "SubHeader",
		TypeServerItem:   "Server Item",
		TypeTreeItem:     "Tree Item",
		TypeTreeLastItem: "Last Tree Item",
	}[it]
}

// MenuItem defines the interface for menu items
type MenuItem interface {
	// Basic list.Item interface requirements
	Title() string
	Description() string
	FilterValue() string

	// Additional menu item functionality
	GetType() ItemType
	IsExpanded() bool
	GetIndent() int
	IsSelectable() bool
	WithExpanded(bool) list.Item
}

// UIState represents the current state of the UI
type UIState struct {
	ActivePane    string
	Width         int
	Height        int
	Ready         bool
	ShowHelp      bool
	StatusMessage string
}

// UIComponents holds the main UI components
type UIComponents struct {
	List     *list.Model
	Viewport *viewport.Model
}

// UIModel defines the interface that handlers can use to interact with the model
type UIModel interface {
	// Tea.Model implementation
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string

	// Core functionality
	GetAPIClient() *api.Client
	SetAPIClient(*api.Client)
	GetActivePane() string
	ToggleActivePane()

	// Size and state
	GetWidth() int
	GetHeight() int
	SetSize(width, height int)
	IsReady() bool

	// Content management
	SetContent(content string)
	SetStatusMessage(msg string)

	// List functionality
	GetList() *list.Model
	UpdateList(msg tea.Msg) tea.Cmd
	SetList(*list.Model)
	UpdateMenuItems()
	ToggleItemExpanded(index int)

	// Viewport functionality
	GetViewport() *viewport.Model
	UpdateViewport(msg tea.Msg) tea.Cmd
	SetViewport(*viewport.Model)

	// Help functionality
	ToggleHelp()
}

// UpdateType represents different types of UI updates
type UpdateType int

const (
	// UpdateContent indicates content should be updated
	UpdateContent UpdateType = iota

	// UpdateMenu indicates the menu should be updated
	UpdateMenu

	// UpdateStatus indicates the status bar should be updated
	UpdateStatus

	// UpdateLayout indicates the layout should be recalculated
	UpdateLayout
)

// UpdateEvent represents a UI update event
type UpdateEvent struct {
	Type    UpdateType
	Content string
	Error   error
}

// NewUpdateEvent creates a new update event
func NewUpdateEvent(updateType UpdateType, content string) UpdateEvent {
	return UpdateEvent{
		Type:    updateType,
		Content: content,
	}
}

// NewErrorEvent creates a new error update event
func NewErrorEvent(err error) UpdateEvent {
	return UpdateEvent{
		Type:  UpdateStatus,
		Error: err,
	}
}

// LayoutManager represents the layout management functionality
type LayoutManager interface {
	Update()
	ValidateWindowSize(width, height int) bool
	CalculateStatusBarWidth() int
	CalculateDimensions() (int, int)
}

