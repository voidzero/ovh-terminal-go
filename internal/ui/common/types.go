// internal/ui/common/types.go

// Package common provides shared types and interfaces for the UI
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

// MenuItem defines the interface for menu items that handlers can interact with
type MenuItem interface {
	Title() string
	Description() string
	GetType() ItemType
	IsExpanded() bool
	GetIndent() int
	IsSelectable() bool
	WithExpanded(bool) list.Item
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

// LayoutManager represents the layout management functionality
type LayoutManager interface {
	Update()
	ValidateWindowSize(width, height int) bool
	CalculateStatusBarWidth() int
}
