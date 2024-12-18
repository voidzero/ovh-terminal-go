# Development Guide

This document provides a comprehensive overview of the OVH Terminal Client's architecture and development workflow. It's intended for developers who have a basic understanding of Go and want to contribute to or understand the project.

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Component Structure](#component-structure)
3. [Key Concepts](#key-concepts)
4. [Core Components](#core-components)
5. [Development Workflow](#development-workflow)
6. [Testing](#testing)

## Architecture Overview

The OVH Terminal Client is built using a modular architecture with clear separation of concerns. The application follows these key architectural principles:

- **Component-based**: Each major piece of functionality is encapsulated in its own package
- **Clean interfaces**: Components communicate through well-defined interfaces
- **Dependency injection**: Components receive their dependencies rather than creating them
- **Error handling**: Comprehensive error types and handling throughout the application

The high-level flow of the application is:

1. Load configuration (config.toml)
2. Initialize logging system
3. Set up API client
4. Start Terminal UI
5. Handle user interactions and API calls

## Component Structure

The application is organized into the following main packages:

```
internal/
├── api/          # OVH API client and types
├── commands/     # Command implementations
├── config/       # Configuration handling
├── format/       # Output formatting utilities
├── logger/       # Logging system
└── ui/          # Terminal user interface
    ├── common/   # Shared UI types and utilities
    ├── handlers/ # UI event handlers
    ├── help/     # Help system
    ├── layout/   # UI layout management
    ├── styles/   # UI styling
    └── types/    # UI type definitions
```

## Key Concepts

### Command Pattern
The application uses the Command pattern to encapsulate operations. Each command:
- Implements the `Command` interface
- Handles a specific operation (e.g., getting account info)
- Can be executed synchronously or asynchronously
- Returns formatted output

Example Command interface:
```go
type Command interface {
    Execute() (string, error)
    ExecuteWithOptions(opts ...CommandOption) (string, error)
    GetType() CommandType
    ExecuteAsync(ctx context.Context) (<-chan CommandResult, error)
}
```

### UI Architecture
The UI is built using the Bubble Tea framework and follows an Model-View-Update (MVU) pattern:

- **Model**: Holds the application state
- **View**: Renders the current state
- **Update**: Handles events and updates the state

The UI is split into two main panes:
1. Left pane: Navigation menu
2. Right pane: Content viewport

### Error Handling
The application uses custom error types for different scenarios:
- `APIError`: For OVH API related errors
- `ValidationError`: For configuration validation errors
- Each error type includes context-specific information and user-friendly messages

## Core Components

### Configuration System
- Located in `internal/config/`
- Uses TOML format
- Supports multiple accounts
- Validates configuration on startup
- Handles sensitive data (API keys)

### API Client
- Located in `internal/api/`
- Wraps the OVH API SDK
- Implements retry logic
- Handles authentication
- Provides type-safe API operations

Example API client usage:
```go
client, err := api.NewClient(cfg, logger)
if err != nil {
    return err
}

info, err := client.GetAccountInfo()
if err != nil {
    return err
}
```

### Logging System
- Located in `internal/logger/`
- Supports multiple log levels
- Can log to file and/or console
- Includes context fields
- Thread-safe operations

### UI System
The UI system is composed of several key parts:

1. **Model** (`ui/types/model.go`):
   - Holds the application state
   - Manages the active pane
   - Handles content updates

2. **Layout Manager** (`ui/layout/layout.go`):
   - Calculates component dimensions
   - Handles window resizing
   - Maintains minimum size requirements

3. **Styles** (`ui/styles/`):
   - Defines color schemes
   - Manages component styling
   - Handles theme switching

4. **Event Handlers** (`ui/handlers/`):
   - Processes keyboard input
   - Handles command execution
   - Manages UI state updates

## Development Workflow

### Setting Up Development Environment

1. Clone the repository
2. Copy `config-example.toml` to `config.toml`
3. Set up OVH API credentials
4. Install dependencies: `go mod download`

### Building and Running

Development build:
```bash
go build
./ovh-terminal-go
```

With specific config:
```bash
./ovh-terminal-go -config=/path/to/config.toml
```

### Adding New Features

When adding new features:

1. **Commands**:
   - Add new command type in `commands/command.go`
   - Implement command in new file
   - Register command in UI handlers

2. **API Operations**:
   - Add new types in `api/types.go`
   - Implement handler in `api/handlers.go`
   - Add endpoint in `api/endpoints.go`

3. **UI Components**:
   - Add new types in appropriate package
   - Update layout if needed
   - Add event handlers

Example of adding a new command:
```go
// commands/new_command.go
type NewCommand struct {
    BaseCommand
    client *api.Client
    log    *logger.Logger
}

func NewNewCommand(client *api.Client) *NewCommand {
    return &NewCommand{
        BaseCommand: NewBaseCommand(TypeInfo),
        client:     client,
        log:        logger.Log.With(map[string]interface{}{"command": "new_command"}),
    }
}

// Implement Command interface methods...
```

## Testing

The application includes several types of tests:

1. **Unit Tests**:
   - Located alongside source files
   - Focus on individual components
   - Use mocking for external dependencies

2. **API Tests**:
   - Located in `api/handlers_test.go`
   - Use mock client for API calls
   - Test error handling

Example test:
```go
func TestNewCommand(t *testing.T) {
    client := setupMockClient()
    cmd := NewNewCommand(client)
    
    output, err := cmd.Execute()
    if err != nil {
        t.Errorf("Command execution failed: %v", err)
    }
    
    // Add assertions...
}
```

### Best Practices

1. **Code Style**:
   - Follow Go standard formatting (`go fmt`)
   - Use meaningful variable names
   - Document public functions
   - Keep functions focused and small

2. **Error Handling**:
   - Use custom error types
   - Include context in errors
   - Log errors appropriately
   - Provide user-friendly messages

3. **Configuration**:
   - Never commit sensitive data
   - Use configuration validation
   - Provide sensible defaults
   - Document all options

4. **UI Development**:
   - Maintain consistent styling
   - Handle window resizing
   - Support keyboard navigation
   - Provide user feedback

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Update documentation
5. Add tests
6. Submit a pull request

For detailed contribution guidelines, see CONTRIBUTING.md (TODO).

---

# Part 2: Implementation Details

This section dives deeper into the technical implementation details of key components.

## API Client Implementation

### Client Structure
The API client is built around the official OVH SDK but adds several important features:

```go
type Client struct {
    client  *ovh.Client
    logger  *logger.Logger
    retry   RetryConfig
    timeout time.Duration
}
```

### Retry Mechanism
The client implements a sophisticated retry mechanism:

```go
func (c *Client) executeWithRetry(operation string, fn func() error) error {
    var lastErr error
    for attempt := 0; attempt < c.retry.MaxRetries; attempt++ {
        if attempt > 0 {
            delay := c.calculateDelay(attempt)
            time.Sleep(delay)
        }
        err := fn()
        if err == nil {
            return nil
        }
        lastErr = err
        if !c.shouldRetry(err, attempt) {
            break
        }
    }
    return lastErr
}
```

Key features:
- Exponential backoff
- Configurable retry counts
- Specific error type handling
- Operation context preservation

## UI Implementation Details

### Model State Management
The UI model maintains state through a structured type system:

```go
type Model struct {
    List          list.Model
    Viewport      viewport.Model
    apiClient     *api.Client
    ActiveCommand commands.Command
    Content       string
    StatusMessage string
    Ready        bool
    ActivePane   string
    Width        int
    Height       int
    ShowHelp     bool
}
```

### Event Flow
The UI event flow follows this pattern:

1. **Event Reception**:
```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return handlers.HandleKeyMsg(m, msg)
    case tea.WindowSizeMsg:
        return handlers.HandleWindowSizeMsg(m, msg)
    }
    // ...
}
```

2. **Event Processing**:
```go
func HandleKeyMsg(model common.UIModel, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    if handler, exists := KeyMap[msg.String()]; exists {
        return handler(model)
    }
    return model, nil
}
```

3. **State Updates**:
```go
func (m *Model) SetContent(content string) {
    m.Content = content
    if m.Viewport.Width > 0 {
        m.Viewport.SetContent(content)
    }
}
```

### Menu System Implementation

The menu system uses a tree-like structure for nested items:

```go
type ListItem struct {
    text       string
    desc       string
    itemType   common.ItemType
    expanded   bool
    indent     int
    selectable bool
}
```

Key features:
- Hierarchical menu structure
- Expandable/collapsible sections
- Visual indentation
- Selection state management

### Layout Management

The layout system handles dynamic resizing:

```go
func (m *Manager) calculateDimensions() Dimensions {
    return Dimensions{
        MenuWidth:     MenuWidth,
        ContentWidth:  totalWidth - MenuWidth - HorizontalSpace,
        ContentHeight: totalHeight - StatusBarSpace - UIElementsSpace,
        StatusWidth:   totalWidth - 2,
    }
}
```

## Command Implementation Details

### Command Execution Flow

Commands follow this execution pattern:

1. **Command Creation**:
```go
type MeCommand struct {
    BaseCommand
    client *api.Client
    log    *logger.Logger
}

func NewMeCommand(client *api.Client) *MeCommand {
    return &MeCommand{
        BaseCommand: NewBaseCommand(TypeInfo),
        client:     client,
        log:        logger.Log.With(map[string]interface{}{"command": "me"}),
    }
}
```

2. **Execution**:
```go
func (c *MeCommand) Execute() (string, error) {
    return c.ExecuteWithOptions()
}

func (c *MeCommand) ExecuteWithOptions(opts ...CommandOption) (string, error) {
    // Apply options
    for _, opt := range opts {
        opt(&c.config)
    }
    
    // Execute with timeout
    return c.executeWithTimeout(context.Background(), func() (string, error) {
        return c.executeCommand()
    })
}
```

3. **Result Formatting**:
```go
func (c *MeCommand) executeCommand() (string, error) {
    info, err := c.client.GetAccountInfo()
    if err != nil {
        return "", fmt.Errorf("failed to get account info: %w", err)
    }

    output := format.NewOutputFormatter(
        format.WithMaxWidth(maxWidth),
        format.WithSeparator("\n"),
    )
    
    // Format sections...
    return output.String(), nil
}
```

## Error Handling Patterns

### API Errors
The application uses structured error types:

```go
type APIError struct {
    Type    ErrorType
    Message string
    Details interface{}
    Err     error
}

func (e *APIError) UserError() string {
    baseMsg := e.Type.UserMessage()
    switch e.Type {
    case ErrorTypeAuth:
        if strings.Contains(strings.ToLower(e.Error()), "invalid") {
            return fmt.Sprintf("%s The credentials appear to be invalid.", baseMsg)
        }
        // ...
    }
    return baseMsg
}
```

### Error Propagation
Errors are enriched as they propagate up the stack:

```go
func (c *Client) GetDedicatedServerInfo(serverID string) (*ServerInfo, error) {
    var info ServerInfo
    err := c.Get(GetServerEndpoint(serverID), &info)
    if err != nil {
        return nil, fmt.Errorf("failed to get server info for %s: %w", serverID, err)
    }
    return &info, nil
}
```

## Configuration Implementation

### Validation System
The configuration system implements thorough validation:

```go
func validateConfig(cfg *Config) error {
    if err := validateGeneral(&cfg.General); err != nil {
        return err
    }
    if err := validateUI(&cfg.UI); err != nil {
        return err
    }
    if err := validateAccounts(cfg.Accounts, cfg.General.DefaultAccount); err != nil {
        return err
    }
    return validateKeyBinds(&cfg.KeyBinds)
}
```

### Security Considerations
Configuration handling includes security checks:

```go
func LoadConfig(path string) (*Config, error) {
    info, err := os.Stat(path)
    if err != nil {
        return nil, err
    }
    mode := info.Mode()
    if mode.Perm()&0o077 != 0 {
        return nil, &ValidationError{
            Field: "permissions",
            Message: fmt.Sprintf(
                "config file %s has too broad permissions %v, should be 600",
                path,
                mode.Perm(),
            ),
        }
    }
    // ...
}
```

## Common Development Tasks

### Adding a New API Endpoint

1. Add endpoint definition:
```go
const endpointNewFeature = "/new/feature"
```

2. Add endpoint builder method:
```go
func GetNewFeatureEndpoint(id string) string {
    return NewEndpointBuilder(ResourceNewFeature).
        WithID(id).
        Build()
}
```

3. Add handler method:
```go
func (c *Client) GetNewFeatureInfo(id string) (*NewFeatureInfo, error) {
    var info NewFeatureInfo
    err := c.Get(GetNewFeatureEndpoint(id), &info)
    if err != nil {
        return nil, fmt.Errorf("failed to get new feature info: %w", err)
    }
    return &info, nil
}
```

### Adding a New Command

1. Create command type:
```go
type NewFeatureCommand struct {
    BaseCommand
    client *api.Client
    log    *logger.Logger
}
```

2. Implement command interface:
```go
func (c *NewFeatureCommand) Execute() (string, error) {
    return c.ExecuteWithOptions()
}

func (c *NewFeatureCommand) executeCommand() (string, error) {
    // Implementation
}
```

3. Add to command registry:
```go
var commandRegistry = map[string]CommandHandler{
    "New Feature": func(client *api.Client) commands.Command {
        return commands.NewNewFeatureCommand(client)
    },
    // ...
}
```

### Adding Menu Items

1. Create new menu items:
```go
NewListItem("New Feature", common.TypeHeader,
    WithDesc("New feature description"),
    WithIndent(1),
    WithExpanded(false))
```

2. Update menu structure:
```go
func (m *Model) UpdateMenuItems() {
    // Add new items to appropriate section
    switch curr.Title() {
    case "New Section":
        addChildItems([]*ListItem{
            NewListItem("New Feature", common.TypeTreeItem,
                WithDesc("Description"),
                WithIndent(1)),
        })
    }
}
```

## Performance Considerations

### Memory Management
- Use appropriate buffer sizes
- Clean up resources in defer statements
- Avoid unnecessary allocations in loops
- Use sync.Pool for frequently allocated objects

### UI Performance
- Limit unnecessary redraws
- Cache computed values when possible
- Use efficient string building
- Handle window resize events efficiently

### API Optimization
- Use appropriate timeout values
- Implement request batching where possible
- Cache frequently used data
- Use connection pooling

## Debugging Tips

1. **Logging**:
```go
log.Debug("Operation debug info",
    "operation", operation,
    "attempt", attempt,
    "delay", delay.String())
```

2. **Error Context**:
```go
if err != nil {
    log.Error("Operation failed",
        "operation", operation,
        "error", err,
        "context", additionalContext)
}
```

3. **UI State**:
```go
log.Debug("UI state update",
    "active_pane", m.GetActivePane(),
    "width", m.GetWidth(),
    "height", m.GetHeight(),
    "content_length", len(m.Content))
```

## Common Pitfalls

1. **Error Handling**:
   - Always wrap errors with context
   - Don't swallow errors
   - Use appropriate error types
   - Provide user-friendly messages

2. **UI Updates**:
   - Handle window resize properly
   - Maintain proper focus management
   - Clean up resources
   - Handle edge cases

3. **API Integration**:
   - Handle rate limiting
   - Implement proper timeouts
   - Handle connection issues
   - Validate responses

## Future Development

Planned improvements and areas for contribution:

1. **Features**:
   - Account switching
   - Data refresh mechanism
   - Search functionality
   - Theme customization
   - Configuration hot reload

2. **Technical Improvements**:
   - Enhanced error handling
   - Better test coverage
   - Performance optimization
   - Documentation updates

3. **UI Enhancements**:
   - Mouse support
   - Context menus
   - Status indicators
   - Progress feedback