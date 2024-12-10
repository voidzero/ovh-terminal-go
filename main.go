package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"ovh-terminal/internal/api"
	"ovh-terminal/internal/config"
	"ovh-terminal/internal/logger"
	"ovh-terminal/internal/ui"
)

// printError formats and prints an error message to stderr
func printError(msg string, details ...string) {
	fmt.Fprintf(os.Stderr, "\n❌ %s\n", msg)
	if len(details) > 0 {
		fmt.Fprintln(os.Stderr, "\nDetails:")
		for _, detail := range details {
			fmt.Fprintf(os.Stderr, "  • %s\n", detail)
		}
	}
}

// printHelp prints helpful instructions
func printHelp(configPath string) {
	fmt.Fprintln(os.Stderr, "\nTo fix this:")
	fmt.Fprintf(os.Stderr, "1. Get your API credentials from https://api.ovh.com/createToken/\n")
	fmt.Fprintf(os.Stderr, "2. Update %s with your credentials\n", configPath)
	fmt.Fprintf(os.Stderr, "3. Make sure you have the following rights:\n")
	fmt.Fprintf(os.Stderr, "   • GET /me\n")
	fmt.Fprintf(os.Stderr, "   • GET /dedicated/server\n")
	fmt.Fprintf(os.Stderr, "   • GET /domain\n")
	fmt.Fprintf(os.Stderr, "   • GET /cloud/project\n")
	fmt.Fprintf(os.Stderr, "   • GET /ip\n")
}

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.toml", "path to config file")
	flag.Parse()

	// Initialize logger early
	log := logger.NewLogger()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		if os.IsNotExist(err) {
			printError(fmt.Sprintf("Could not find configuration file: %s", *configPath))
			fmt.Fprintf(os.Stderr, "\nTo get started:\n")
			fmt.Fprintf(os.Stderr, "1. Copy config-example.toml to %s\n", *configPath)
			fmt.Fprintf(os.Stderr, "2. Edit the file with your API credentials\n")
		} else {
			printError("Configuration file is invalid", err.Error())
		}
		os.Exit(1)
	}

	// Configure logger to only write to file, not console
	if err := log.Configure(cfg.General.LogLevel, cfg.General.LogFile, false); err != nil {
		printError("Failed to configure logging", err.Error())
		os.Exit(1)
	}

	log.Info("Starting OVH Terminal Client")

	// Initialize API client
	account := cfg.Accounts[cfg.General.DefaultAccount]
	apiClient, err := api.NewClient(&account, log)
	if err != nil {
		log.Error("Failed to initialize API client", "error", err)
		printError("Could not initialize API client", err.Error())
		os.Exit(1)
	}

	// Validate credentials by attempting to get account info
	log.Info("Validating API credentials...")
	if _, err := apiClient.GetAccountInfo(); err != nil {
		log.Error("Failed to validate credentials", "error", err)

		var details []string
		if apiErr, ok := err.(*api.APIError); ok {
			// Extract the actual error message without the path/query ID
			errMsg := apiErr.Error()
			if idx := strings.Index(errMsg, " (X-OVH-Query-Id:"); idx != -1 {
				errMsg = errMsg[:idx]
			}
			details = append(details, errMsg)
		}

		printError("Invalid or expired API credentials", details...)
		printHelp(*configPath)
		os.Exit(1)
	}
	log.Info("API credentials validated successfully")

	// Initialize UI
	p := tea.NewProgram(
		ui.Initialize(apiClient),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the application
	if _, err := p.Run(); err != nil {
		log.Error("Error running application", "error", err)
		printError("Application crashed", err.Error())
		os.Exit(1)
	}
}
