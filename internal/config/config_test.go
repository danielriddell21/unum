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

func boolPtr(v bool) *bool { return &v }

func TestTelemetryEnabled_Default(t *testing.T) {
	t.Setenv("DO_NOT_TRACK", "")
	t.Setenv("UNUM_NO_TELEMETRY", "")

	cfg := Config{}
	if !cfg.TelemetryEnabled() {
		t.Error("nil Telemetry pointer should default to enabled")
	}
}

func TestTelemetryEnabled_ExplicitFalse(t *testing.T) {
	t.Setenv("DO_NOT_TRACK", "")
	t.Setenv("UNUM_NO_TELEMETRY", "")

	cfg := Config{Telemetry: boolPtr(false)}
	if cfg.TelemetryEnabled() {
		t.Error("explicit false should disable telemetry")
	}
}

func TestTelemetryEnabled_ExplicitTrue(t *testing.T) {
	t.Setenv("DO_NOT_TRACK", "")
	t.Setenv("UNUM_NO_TELEMETRY", "")

	cfg := Config{Telemetry: boolPtr(true)}
	if !cfg.TelemetryEnabled() {
		t.Error("explicit true should enable telemetry")
	}
}

func TestTelemetryEnabled_DONotTrack(t *testing.T) {
	t.Setenv("DO_NOT_TRACK", "1")
	t.Setenv("UNUM_NO_TELEMETRY", "")

	cfg := Config{Telemetry: boolPtr(true)}
	if cfg.TelemetryEnabled() {
		t.Error("DO_NOT_TRACK=1 should override config")
	}
}

func TestTelemetryEnabled_UnumNoTelemetry(t *testing.T) {
	t.Setenv("DO_NOT_TRACK", "")
	t.Setenv("UNUM_NO_TELEMETRY", "1")

	cfg := Config{Telemetry: boolPtr(true)}
	if cfg.TelemetryEnabled() {
		t.Error("UNUM_NO_TELEMETRY=1 should override config")
	}
}

func TestSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)

	original := Config{
		DarkTheme:  "nord",
		LightTheme: "solarized",
		Telemetry:  boolPtr(false),
		ClientID:   "test-uuid",
	}
	if err := Save(original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded := Load()
	if loaded.DarkTheme != original.DarkTheme {
		t.Errorf("DarkTheme=%q, want %q", loaded.DarkTheme, original.DarkTheme)
	}
	if loaded.LightTheme != original.LightTheme {
		t.Errorf("LightTheme=%q, want %q", loaded.LightTheme, original.LightTheme)
	}
	if loaded.Telemetry == nil || *loaded.Telemetry != false {
		t.Error("Telemetry should be false after round-trip")
	}
	if loaded.ClientID != original.ClientID {
		t.Errorf("ClientID=%q, want %q", loaded.ClientID, original.ClientID)
	}
}

func TestEnsureClientID_Generates(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := EnsureClientID()
	if cfg.ClientID == "" {
		t.Fatal("EnsureClientID should generate a UUID")
	}
	if len(cfg.ClientID) != 36 {
		t.Errorf("ClientID=%q, want UUID format (36 chars)", cfg.ClientID)
	}
}

func TestEnsureClientID_Idempotent(t *testing.T) {
	dir := writeConfig(t, map[string]string{
		"dark_theme": "cyber",
		"client_id":  "existing-id",
	})
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := EnsureClientID()
	if cfg.ClientID != "existing-id" {
		t.Errorf("ClientID=%q, want \"existing-id\" (should not overwrite)", cfg.ClientID)
	}
}
