// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidLogLevels defines allowed log levels
var ValidLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

// ValidEndpoints defines allowed OVH API endpoints
var ValidEndpoints = map[string]bool{
	"ovh-eu":     true,
	"ovh-us":     true,
	"ovh-ca":     true,
	"kimsufi-eu": true,
	"kimsufi-ca": true,
	"soyoustart": true,
	"runabove":   true,
}

// LoadConfig reads and parses the configuration file
func LoadConfig(path string) (*Config, error) {
	var cfg Config

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, &ValidationError{
			Field:   "config_file",
			Message: fmt.Sprintf("configuration file not found: %s", path),
		}
	}

	// Check file permissions
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("error checking file permissions: %w", err)
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

	// Decode TOML
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing configuration: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validateConfig performs validation of the configuration
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

	if err := validateKeyBinds(&cfg.KeyBinds); err != nil {
		return err
	}

	return nil
}

// validateGeneral validates general configuration
func validateGeneral(gen *GeneralConfig) error {
	if gen.DefaultAccount == "" {
		return &ValidationError{
			Field:   "general.default_account",
			Message: "no default account specified",
		}
	}

	if !ValidLogLevels[strings.ToLower(gen.LogLevel)] {
		return &ValidationError{
			Field:   "general.log_level",
			Message: fmt.Sprintf("invalid log level: %s", gen.LogLevel),
		}
	}

	if gen.LogFile != "" && gen.LogFile != "none" {
		dir := filepath.Dir(gen.LogFile)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return &ValidationError{
				Field:   "general.log_file",
				Message: fmt.Sprintf("cannot create log directory: %v", err),
			}
		}
	}

	return nil
}

// validateUI validates UI configuration
func validateUI(ui *UIConfig) error {
	if ui.RefreshInterval < 0 {
		return &ValidationError{
			Field:   "ui.refresh_interval",
			Message: "refresh interval cannot be negative",
		}
	}

	return nil
}

// validateAccounts validates account configurations
func validateAccounts(accounts map[string]AccountConfig, defaultAccount string) error {
	if len(accounts) == 0 {
		return &ValidationError{
			Field:   "accounts",
			Message: "no accounts configured",
		}
	}

	if _, exists := accounts[defaultAccount]; !exists {
		return &ValidationError{
			Field: "general.default_account",
			Message: fmt.Sprintf(
				"default account '%s' not found in accounts configuration",
				defaultAccount,
			),
		}
	}

	for name, acc := range accounts {
		if err := validateAccount(name, &acc); err != nil {
			return err
		}
	}

	return nil
}

// validateAccount validates a single account configuration
func validateAccount(name string, acc *AccountConfig) error {
	if acc.Endpoint == "" {
		return &ValidationError{
			Field:   fmt.Sprintf("accounts.%s.endpoint", name),
			Message: "missing endpoint",
		}
	}
	if !ValidEndpoints[acc.Endpoint] {
		return &ValidationError{
			Field:   fmt.Sprintf("accounts.%s.endpoint", name),
			Message: fmt.Sprintf("invalid endpoint: %s", acc.Endpoint),
		}
	}

	if acc.AppKey == "" {
		return &ValidationError{
			Field:   fmt.Sprintf("accounts.%s.app_key", name),
			Message: "missing app_key",
		}
	}

	if acc.AppSecret == "" {
		return &ValidationError{
			Field:   fmt.Sprintf("accounts.%s.app_secret", name),
			Message: "missing app_secret",
		}
	}

	if acc.ConsumerKey == "" {
		return &ValidationError{
			Field:   fmt.Sprintf("accounts.%s.consumer_key", name),
			Message: "missing consumer_key",
		}
	}

	return nil
}

// validateKeyBinds validates keybinding configuration
func validateKeyBinds(kb *KeyBindConfig) error {
	// Ensure required keybindings are present
	if len(kb.Quit) == 0 {
		return &ValidationError{
			Field:   "keybindings.quit",
			Message: "at least one quit keybinding is required",
		}
	}

	if len(kb.Help) == 0 {
		return &ValidationError{
			Field:   "keybindings.help",
			Message: "at least one help keybinding is required",
		}
	}

	return nil
}

