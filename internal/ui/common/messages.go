// internal/ui/common/messages.go

// Package common provides shared functionality for the UI
package common

// MessageType represents different types of UI messages
type MessageType int

const (
	// Message types for different UI events
	MessageNavigate MessageType = iota
	MessageResize
	MessageQuit
	MessageTogglePane
)

// Message represents a UI message with its type and data
type Message struct {
	Type MessageType
	Data interface{}
}

// NavigationDirection represents different navigation commands
type NavigationDirection int

const (
	NavUp NavigationDirection = iota
	NavDown
	NavTop
	NavBottom
)

// NavigationMessage contains navigation-specific data
type NavigationMessage struct {
	Direction NavigationDirection
	Pane      string
}
