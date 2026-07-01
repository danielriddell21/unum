package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type Config struct {
	DarkTheme  string `json:"dark_theme"`
	LightTheme string `json:"light_theme"`
	Telemetry  *bool  `json:"telemetry,omitempty"`
	ClientID   string `json:"client_id,omitempty"`
}

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

func EnsureClientID() Config {
	cfg := Load()
	if cfg.ClientID != "" {
		return cfg
	}
	cfg.ClientID = uuid.NewString()
	_ = Save(cfg)
	return cfg
}
