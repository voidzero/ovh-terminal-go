// internal/commands/command.go
package commands

// Command defines the interface for all commands
type Command interface {
	Execute() (string, error)
}

