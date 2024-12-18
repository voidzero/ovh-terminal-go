// internal/commands/server.go
package commands

import (
	"context"
	"fmt"
	"time"

	"ovh-terminal/internal/api"
	"ovh-terminal/internal/logger"
)

// ServerCommand handles dedicated server operations
type ServerCommand struct {
	BaseCommand
	client *api.Client
	log    *logger.Logger
}

// NewServerCommand creates a new server command instance
func NewServerCommand(client *api.Client) *ServerCommand {
	return &ServerCommand{
		BaseCommand: NewBaseCommand(TypeInfo),
		client:      client,
		log:         logger.Log.With(map[string]interface{}{"command": "server"}),
	}
}

// Execute implements the Command interface
func (c *ServerCommand) Execute() (string, error) {
	return c.ExecuteWithOptions()
}

// ExecuteWithOptions implements the Command interface
func (c *ServerCommand) ExecuteWithOptions(opts ...CommandOption) (string, error) {
	// Apply options to base command
	for _, opt := range opts {
		opt(&c.config)
	}

	return c.executeWithTimeout(context.Background(), func() (string, error) {
		return c.executeCommand()
	})
}

// ExecuteAsync implements the Command interface
func (c *ServerCommand) ExecuteAsync(ctx context.Context) (<-chan CommandResult, error) {
	resultCh := make(chan CommandResult, 1)

	go func() {
		defer close(resultCh)

		start := time.Now()
		output, err := c.executeCommand()
		duration := time.Since(start)

		state := StateCompleted
		if err != nil {
			state = StateFailed
		}

		resultCh <- CommandResult{
			Output:   output,
			Error:    err,
			Duration: duration,
			State:    state,
		}
	}()

	return resultCh, nil
}

// GetServerDisplayName returns the best available name for a server
func (c *ServerCommand) GetServerDisplayName(serverName string) (string, error) {
	info, err := c.client.GetDedicatedServerInfo(serverName)
	if err != nil {
		return serverName, fmt.Errorf("failed to get server info: %w", err)
	}

	// Try IAM display name first
	if info.IAM != nil && info.IAM.DisplayName != "" {
		return info.IAM.DisplayName, nil
	}

	// Then try reverse DNS
	if info.Reverse != "" {
		return info.Reverse, nil
	}

	// Fall back to server name
	return info.Name, nil
}

// ListServers returns a list of all dedicated servers with their display names
func (c *ServerCommand) ListServers() (map[string]string, error) {
	c.log.Debug("Fetching server list")

	// Get list of server IDs
	servers, err := c.client.ListDedicatedServers()
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	// Create a map of server ID to display name
	result := make(map[string]string)
	for _, server := range servers {
		displayName, err := c.GetServerDisplayName(server)
		if err != nil {
			c.log.Error("Failed to get display name for server",
				"server", server,
				"error", err)
			displayName = server // Fallback to server ID
		}
		result[server] = displayName
	}

	return result, nil
}

// executeCommand handles the actual command execution
func (c *ServerCommand) executeCommand() (string, error) {
	c.log.Debug("Executing server command")

	servers, err := c.ListServers()
	if err != nil {
		return "", err
	}

	if len(servers) == 0 {
		return "No dedicated servers found.", nil
	}

	// For now just return a simple list
	// Later we can format this nicely with the format package
	output := "Dedicated Servers:\n\n"
	for id, name := range servers {
		output += fmt.Sprintf("%s (%s)\n", name, id)
	}

	return output, nil
}
