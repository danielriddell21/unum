package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DarkTheme   string `json:"dark_theme"`
	LightTheme  string `json:"light_theme"`
	Telemetry   *bool  `json:"telemetry,omitempty"`
	HashHistory *bool  `json:"hash_history,omitempty"`
	NoticeShown bool   `json:"notice_shown,omitempty"`
}

func (c Config) TelemetryEnabled() bool {
	// Any non-empty value opts out, matching the NO_COLOR convention.
	if os.Getenv("DO_NOT_TRACK") != "" {
		return false
	}
	if os.Getenv("UNUM_NO_TELEMETRY") != "" {
		return false
	}
	if c.Telemetry != nil {
		return *c.Telemetry
	}
	return true
}

func (c Config) HashHistoryEnabled() bool {
	if os.Getenv("UNUM_NO_HASH_HISTORY") != "" {
		return false
	}
	if c.HashHistory != nil {
		return *c.HashHistory
	}
	return true
}

func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	return filepath.Join(dir, "unum", "config.json"), nil
}

func Load() Config {
	cfg := Config{DarkTheme: "cyber", LightTheme: "clean"}

	path, err := Path()
	if err != nil {
		return cfg
	}

	data, err := os.ReadFile(path)
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

func MarkNoticeShown(cfg Config) Config {
	cfg.NoticeShown = true
	_ = Save(cfg)
	return cfg
}
