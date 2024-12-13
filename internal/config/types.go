// internal/config/types.go
package config

// Config represents the root configuration structure
type Config struct {
	General  GeneralConfig            `toml:"general"`
	UI       UIConfig                 `toml:"ui"`
	Accounts map[string]AccountConfig `toml:"accounts"`
	KeyBinds KeyBindConfig            `toml:"keybindings"`
}

// GeneralConfig holds general application settings
type GeneralConfig struct {
	DefaultAccount string `toml:"default_account"`
	LogLevel       string `toml:"log_level"`
	LogFile        string `toml:"log_file"`
	LogConsole     bool   `toml:"log_console"`
}

// UIConfig holds UI-related settings
type UIConfig struct {
	Theme           string `toml:"theme"`
	CompactView     bool   `toml:"compact_view"`
	StatusBar       bool   `toml:"status_bar"`
	RefreshInterval int    `toml:"refresh_interval"`
}

// AccountConfig holds OVH API credentials
type AccountConfig struct {
	Name        string `toml:"name"`
	Endpoint    string `toml:"endpoint"`
	AppKey      string `toml:"app_key"`
	AppSecret   string `toml:"app_secret"`
	ConsumerKey string `toml:"consumer_key"`
}

// KeyBindConfig holds keyboard shortcuts configuration
type KeyBindConfig struct {
	Quit          []string `toml:"quit"`
	Help          []string `toml:"help"`
	Refresh       []string `toml:"refresh"`
	SwitchAccount []string `toml:"switch_account"`
	ToggleView    []string `toml:"toggle_view"`
}
