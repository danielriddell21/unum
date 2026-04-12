package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, v any) string {
	t.Helper()
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "unum")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestLoad_Defaults(t *testing.T) {
	// Point config dir somewhere empty.
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := Load()
	if cfg.DarkTheme != "cyber" {
		t.Errorf("DarkTheme=%q, want \"cyber\"", cfg.DarkTheme)
	}
	if cfg.LightTheme != "clean" {
		t.Errorf("LightTheme=%q, want \"clean\"", cfg.LightTheme)
	}
}

func TestLoad_DarkTheme(t *testing.T) {
	dir := writeConfig(t, map[string]string{"dark_theme": "dracula"})
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := Load()
	if cfg.DarkTheme != "dracula" {
		t.Errorf("DarkTheme=%q, want \"dracula\"", cfg.DarkTheme)
	}
	// Other fields keep defaults.
	if cfg.LightTheme != "clean" {
		t.Errorf("LightTheme=%q, want \"clean\"", cfg.LightTheme)
	}
}

func TestLoad_LightTheme(t *testing.T) {
	dir := writeConfig(t, map[string]string{"light_theme": "solarized"})
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := Load()
	if cfg.LightTheme != "solarized" {
		t.Errorf("LightTheme=%q, want \"solarized\"", cfg.LightTheme)
	}
}

func TestLoad_AllFields(t *testing.T) {
	dir := writeConfig(t, map[string]string{
		"dark_theme":  "nord",
		"light_theme": "solarized",
	})
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := Load()
	if cfg.DarkTheme != "nord" {
		t.Errorf("DarkTheme=%q, want \"nord\"", cfg.DarkTheme)
	}
	if cfg.LightTheme != "solarized" {
		t.Errorf("LightTheme=%q, want \"solarized\"", cfg.LightTheme)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// Should not panic — returns defaults.
	cfg := Load()
	if cfg.DarkTheme == "" {
		t.Error("DarkTheme should not be empty")
	}
}
