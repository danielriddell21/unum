// Package config handles loading user configuration from disk.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// Config holds persistent user preferences for unum.
type Config struct {
	DarkTheme  string `json:"dark_theme"`          // cyber | matrix | dracula | nord
	LightTheme string `json:"light_theme"`         // clean | solarized
	Telemetry  *bool  `json:"telemetry,omitempty"` // nil = enabled (opt-out default), false = disabled
	ClientID   string `json:"client_id,omitempty"` // random UUID for anonymous telemetry identity
}

// TelemetryEnabled reports whether telemetry should be active.
// Priority: DO_NOT_TRACK=1 > UNUM_NO_TELEMETRY=1 > config file > default (true).
func (c Config) TelemetryEnabled() bool {
	if os.Getenv("DO_NOT_TRACK") == "1" {
		return false
	}
	if os.Getenv("UNUM_NO_TELEMETRY") == "1" {
		return false
	}
	if c.Telemetry != nil {
		return *c.Telemetry
	}
	return true
}

// Load reads the config file from the OS config directory.
// Returns defaults if the file is absent or malformed.
// Location: $XDG_CONFIG_HOME/unum/config.json  (Linux/macOS)
//
//	%APPDATA%\unum\config.json             (Windows)
func Load() Config {
	cfg := Config{DarkTheme: "cyber", LightTheme: "clean"}

	dir, err := os.UserConfigDir()
	if err != nil {
		return cfg
	}

	data, err := os.ReadFile(filepath.Join(dir, "unum", "config.json"))
	if err != nil {
		return cfg // file absent — use defaults
	}

	_ = json.Unmarshal(data, &cfg)
	return cfg
}

// Save writes the config back to the OS config directory.
func Save(cfg Config) error {
	dir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("config dir: %w", err)
	}

	cfgDir := filepath.Join(dir, "unum")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), data, 0o600); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

// EnsureClientID loads the config, generates a client ID if missing, saves,
// and returns the config. If saving fails, the in-memory config is still
// returned with the generated ID.
func EnsureClientID() Config {
	cfg := Load()
	if cfg.ClientID != "" {
		return cfg
	}
	cfg.ClientID = uuid.NewString()
	_ = Save(cfg)
	return cfg
}
