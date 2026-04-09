// Package config handles loading user configuration from disk.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds persistent user preferences for unum.
type Config struct {
	Theme string `json:"theme"` // cyber | matrix | dracula | nord
}

// Load reads the config file from the OS config directory.
// Returns defaults if the file is absent or malformed.
// Location: $XDG_CONFIG_HOME/unum/config.json  (Linux/macOS)
//
//	%APPDATA%\unum\config.json             (Windows)
func Load() Config {
	cfg := Config{Theme: "cyber"}

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
