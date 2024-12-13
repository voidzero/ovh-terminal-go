// internal/config/config.go
package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// LoadConfig reads and parses the configuration file
func LoadConfig(path string) (*Config, error) {
	var cfg Config

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", path)
	}

	// Decode TOML
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing configuration: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// validateConfig performs validation of the configuration
func validateConfig(cfg *Config) error {
	// Check general section
	if cfg.General.DefaultAccount == "" {
		return fmt.Errorf("no default account specified")
	}

	// Check if default account exists
	if _, exists := cfg.Accounts[cfg.General.DefaultAccount]; !exists {
		return fmt.Errorf(
			"default account '%s' not found in accounts configuration",
			cfg.General.DefaultAccount,
		)
	}

	// Validate all accounts
	for name, acc := range cfg.Accounts {
		if err := validateAccount(name, &acc); err != nil {
			return err
		}
	}

	return nil
}

// validateAccount validates a single account configuration
func validateAccount(name string, acc *AccountConfig) error {
	if acc.Endpoint == "" {
		return fmt.Errorf("missing endpoint for account '%s'", name)
	}
	if acc.AppKey == "" {
		return fmt.Errorf("missing app_key for account '%s'", name)
	}
	if acc.AppSecret == "" {
		return fmt.Errorf("missing app_secret for account '%s'", name)
	}
	if acc.ConsumerKey == "" {
		return fmt.Errorf("missing consumer_key for account '%s'", name)
	}
	return nil
}
