package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"ovh-terminal/internal/api"
	"ovh-terminal/internal/config"
	"ovh-terminal/internal/logger"
	"ovh-terminal/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	exitSuccess = 0
	exitError   = 1
)

// AppConfig holds application configuration and components
type AppConfig struct {
	ConfigPath string
	Config     *config.Config
	Logger     *logger.Logger
	APIClient  *api.Client
}

// initLogger initializes the logging system
func initLogger(cfg *config.Config) (*logger.Logger, error) {
	log := logger.NewLogger()

	// Create logs directory if it doesn't exist
	if cfg.General.LogFile != "" && cfg.General.LogFile != "none" {
		logDir := filepath.Dir(cfg.General.LogFile)
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	if err := log.Configure(cfg.General.LogLevel, cfg.General.LogFile, false); err != nil {
		return nil, fmt.Errorf("failed to configure logging: %w", err)
	}

	return log, nil
}

// initAPIClient initializes the OVH API client
func initAPIClient(cfg *config.AccountConfig, log *logger.Logger) (*api.Client, error) {
	client, err := api.NewClient(cfg, log)
	if err != nil {
		log.Error("Failed to create API client", "error", err)
		return nil, err
	}

	// Verify credentials by attempting to get account info
	log.Info("Validating API credentials...")
	if _, err := client.GetAccountInfo(); err != nil {
		return nil, fmt.Errorf("invalid API credentials:\n%w", err)
	}

	return client, nil
}

// printError formats and prints an error message
func printError(msg string, details ...string) {
	fmt.Fprintf(os.Stderr, "\n❌ %s\n", msg)
	if len(details) > 0 {
		fmt.Fprintln(os.Stderr, "\nDetails:")
		for _, detail := range details {
			fmt.Fprintf(os.Stderr, "  • %s\n", detail)
		}
	}
}

// printHelp prints helpful instructions for API setup
func printHelp(configPath string) {
	fmt.Fprintln(os.Stderr, "\nTo set up OVH API access:")
	fmt.Fprintf(os.Stderr, "1. Get your API credentials from https://api.ovh.com/createToken/\n")
	fmt.Fprintf(os.Stderr, "2. Update %s with your credentials\n", configPath)
	fmt.Fprintf(os.Stderr, "3. Ensure you have the following API rights:\n")
	fmt.Fprintln(os.Stderr, "   • GET /me")
	fmt.Fprintln(os.Stderr, "   • GET /dedicated/server")
	fmt.Fprintln(os.Stderr, "   • GET /domain")
	fmt.Fprintln(os.Stderr, "   • GET /cloud/project")
	fmt.Fprintln(os.Stderr, "   • GET /ip")
}

// setupConfig loads and initializes all application components
func setupConfig() (*AppConfig, error) {
	app := &AppConfig{}

	// Parse command line flags
	flag.StringVar(&app.ConfigPath, "config", "config.toml", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(app.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("configuration file not found: %s", app.ConfigPath)
		}
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	app.Config = cfg

	// Initialize logger
	log, err := initLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("logging setup failed: %w", err)
	}
	app.Logger = log

	// Initialize API client
	account := cfg.Accounts[cfg.General.DefaultAccount]
	client, err := initAPIClient(&account, log)
	if err != nil {
		printError(err.Error(), "API client setup failed")
		printHelp(app.ConfigPath)
		return nil, err
	}
	app.APIClient = client

	return app, nil
}

func main() {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	// Set up application configuration and components
	app, err := setupConfig()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "\nTo get started:\n")
			fmt.Fprintf(os.Stderr, "1. Copy config-example.toml to %s\n", app.ConfigPath)
			fmt.Fprintf(os.Stderr, "2. Edit the file with your API credentials\n")
		} else {
			printHelp(app.ConfigPath)
			printError(err.Error())
		}
		exitCode = exitError
		return
	}

	app.Logger.Info("Starting OVH Terminal Client")

	// Initialize and run UI
	p := tea.NewProgram(
		ui.Initialize(app.APIClient),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		app.Logger.Error("Application crashed", "error", err)
		printError("Application crashed", err.Error())
		exitCode = exitError
		return
	}

	exitCode = exitSuccess
}
