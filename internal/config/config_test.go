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
}

func TestLoad_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Write corrupt JSON to the config file.
	cfgDir := filepath.Join(dir, "unum")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte("{bad"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Should silently return defaults.
	cfg := Load()
	if cfg.DarkTheme != "cyber" {
		t.Errorf("malformed JSON: DarkTheme=%q, want 'cyber' (default)", cfg.DarkTheme)
	}
}

func TestTelemetryEnabled_DONotTrackAnyValue(t *testing.T) {
	for _, v := range []string{"1", "true", "yes", "0"} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("DO_NOT_TRACK", v)
			t.Setenv("UNUM_NO_TELEMETRY", "")
			if (Config{}).TelemetryEnabled() {
				t.Errorf("DO_NOT_TRACK=%q should disable telemetry", v)
			}
		})
	}
}

func TestTelemetryEnabled_UnumNoTelemetryAnyValue(t *testing.T) {
	t.Setenv("DO_NOT_TRACK", "")
	t.Setenv("UNUM_NO_TELEMETRY", "true")
	if (Config{}).TelemetryEnabled() {
		t.Error("UNUM_NO_TELEMETRY=true should disable telemetry")
	}
}

func TestHashHistoryEnabled(t *testing.T) {
	tests := []struct {
		name string
		env  string
		cfg  *bool
		want bool
	}{
		{"default", "", nil, true},
		{"config false", "", boolPtr(false), false},
		{"config true", "", boolPtr(true), true},
		{"env overrides", "1", boolPtr(true), false},
		{"env any value", "no", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("UNUM_NO_HASH_HISTORY", tt.env)
			got := Config{HashHistory: tt.cfg}.HashHistoryEnabled()
			if got != tt.want {
				t.Errorf("HashHistoryEnabled()=%v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarkNoticeShown(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if Load().NoticeShown {
		t.Fatal("NoticeShown should start false")
	}
	MarkNoticeShown(Load())
	if !Load().NoticeShown {
		t.Error("NoticeShown should persist as true")
	}
}

func TestPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)

	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "config.json" {
		t.Errorf("Path()=%q, want it to end in config.json", path)
	}
}
