// internal/commands/command.go
package commands

import (
	"context"
	"errors"
	"time"
)

// CommandType represents different types of commands
type CommandType int

const (
	// TypeInfo represents informational commands (read-only)
	TypeInfo CommandType = iota

	// TypeAction represents commands that modify resources
	TypeAction

	// TypeBulk represents commands that operate on multiple resources
	TypeBulk
)

// CommandState represents the current state of a command
type CommandState int

const (
	// StateNew indicates a newly created command
	StateNew CommandState = iota

	// StateRunning indicates the command is executing
	StateRunning

	// StateCompleted indicates successful completion
	StateCompleted

	// StateFailed indicates command failure
	StateFailed
)

// CommandResult contains the result of a command execution
type CommandResult struct {
	Output   string
	Error    error
	Duration time.Duration
	State    CommandState
}

// CommandOption defines a function type for configuring commands
type CommandOption func(*CommandConfig)

// CommandConfig holds configuration for command execution
type CommandConfig struct {
	Timeout     time.Duration
	RetryCount  int
	RetryDelay  time.Duration
	Interactive bool
}

var defaultConfig = CommandConfig{
	Timeout:     30 * time.Second,
	RetryCount:  3,
	RetryDelay:  time.Second,
	Interactive: false,
}

// WithTimeout sets a command timeout
func WithTimeout(d time.Duration) CommandOption {
	return func(c *CommandConfig) {
		c.Timeout = d
	}
}

// WithRetry configures retry behavior
func WithRetry(count int, delay time.Duration) CommandOption {
	return func(c *CommandConfig) {
		c.RetryCount = count
		c.RetryDelay = delay
	}
}

// WithInteractive enables interactive mode
func WithInteractive(interactive bool) CommandOption {
	return func(c *CommandConfig) {
		c.Interactive = interactive
	}
}

// Command defines the interface for all commands
type Command interface {
	// Execute runs the command with default configuration
	Execute() (string, error)

	// ExecuteWithOptions runs the command with specific options
	ExecuteWithOptions(opts ...CommandOption) (string, error)

	// GetType returns the command type
	GetType() CommandType

	// ExecuteAsync runs the command asynchronously
	ExecuteAsync(ctx context.Context) (<-chan CommandResult, error)
}

// BaseCommand provides common functionality for commands
type BaseCommand struct {
	cmdType CommandType
	config  CommandConfig
	state   CommandState
}

// NewBaseCommand creates a new base command
func NewBaseCommand(cmdType CommandType, opts ...CommandOption) BaseCommand {
	config := defaultConfig

	for _, opt := range opts {
		opt(&config)
	}

	return BaseCommand{
		cmdType: cmdType,
		config:  config,
		state:   StateNew,
	}
}

// GetType implements Command interface
func (b *BaseCommand) GetType() CommandType {
	return b.cmdType
}

// executeWithTimeout wraps command execution with timeout
func (b *BaseCommand) executeWithTimeout(
	ctx context.Context,
	fn func() (string, error),
) (string, error) {
	if b.config.Timeout <= 0 {
		return fn()
	}

	ctx, cancel := context.WithTimeout(ctx, b.config.Timeout)
	defer cancel()

	resultCh := make(chan struct {
		output string
		err    error
	}, 1)

	go func() {
		output, err := fn()
		resultCh <- struct {
			output string
			err    error
		}{output, err}
	}()

	select {
	case result := <-resultCh:
		return result.output, result.err
	case <-ctx.Done():
		return "", errors.New("command execution timed out")
	}
}

// executeWithRetry wraps command execution with retry logic
func (b *BaseCommand) executeWithRetry(
	ctx context.Context,
	fn func() (string, error),
) (string, error) {
	var lastError error

	for attempt := 0; attempt <= b.config.RetryCount; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(b.config.RetryDelay):
			}
		}

		output, err := fn()
		if err == nil {
			return output, nil
		}

		lastError = err
	}

	return "", lastError
}

// ErrCommandCanceled indicates command cancellation
var ErrCommandCanceled = errors.New("command canceled")

// IsRetryableError determines if an error should trigger a retry
func IsRetryableError(err error) bool {
	// Add specific error type checks here
	return true
}

// CommandProgress represents command execution progress
type CommandProgress struct {
	Step       int
	TotalSteps int
	Message    string
	Error      error
}

// ProgressReporter defines the interface for reporting command progress
type ProgressReporter interface {
	// ReportProgress reports command execution progress
	ReportProgress(CommandProgress)
}
