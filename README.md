# OVH Terminal Client

A terminal-based client for managing OVH services, built with Go using the
Bubbletea TUI framework.

## Features

- View account information
- Manage dedicated servers
- Handle domain management
- Overview cloud projects
- Manage IP addresses
- Terminal user interface with vim-style navigation

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/ovh-terminal-go
cd ovh-terminal-go
```

2. Create your configuration file:
```bash
cp config-example.toml config.toml
```

3. Edit `config.toml` with your OVH API credentials. You can get these from:
   https://api.ovh.com/createToken/

   Required API permissions:
   - GET /me
   - GET /dedicated/server
   - GET /domain
   - GET /cloud/project
   - GET /ip

4. Build the application:
```bash
go build
```

## Usage

Start the application:
```bash
./ovh-terminal-go
```

Navigation:
- Arrow keys to move through menu items
- Enter to select
- q to quit
- ? for help (coming soon)

## Configuration

The application uses a TOML configuration file. See `config-example.toml` for
all available options:

- Multiple account support
- Configurable logging
- UI preferences
- Custom key bindings

## Logging

Logs are stored in the `logs` directory by default. The log level and location
can be configured in `config.toml`.

## Development

Requirements:
- Go 1.23.4 or higher (lower versions may work, ymmv)
- Various Go packages (see go.mod)

Main components:
- `internal/api/`: OVH API client implementation
- `internal/commands/`: Command implementations
- `internal/config/`: Configuration handling
- `internal/logger/`: Logging functionality
- `internal/ui/`: Terminal user interface

## Contributing

1. Fork the repository
2. Create your feature branch
3. Make your changes
4. Ensure code style consistency
5. Submit a pull request

## License

[zlib License](LICENSE) - see the LICENSE file for details.

## Credits

Built using:
- [go-ovh](https://github.com/ovh/go-ovh)
- [bubbletea](https://github.com/charmbracelet/bubbletea)
- [lipgloss](https://github.com/charmbracelet/lipgloss)

