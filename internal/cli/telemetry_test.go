package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/config"
)

func isolateConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("APPDATA", dir)
	t.Setenv("DO_NOT_TRACK", "")
	t.Setenv("UNUM_NO_TELEMETRY", "")
	return dir
}

func runCmd(t *testing.T, args ...string) string {
	t.Helper()
	cmd := telemetryCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute %v: %v", args, err)
	}
	return out.String()
}

func TestTelemetryStatus_Default(t *testing.T) {
	isolateConfig(t)

	out := runCmd(t, "status")
	if !strings.Contains(out, "telemetry: enabled (default)") {
		t.Errorf("status should report the default state, got %q", out)
	}
	if !strings.Contains(out, "endpoint:") {
		t.Errorf("status should report the endpoint, got %q", out)
	}
	if !strings.Contains(out, "client id: none stored") {
		t.Errorf("status should report no id before one is generated, got %q", out)
	}
}

func TestTelemetryStatus_ReportsStoredClientID(t *testing.T) {
	isolateConfig(t)
	config.EnsureClientID()

	out := runCmd(t, "status")
	if strings.Contains(out, "none stored") {
		t.Errorf("status should print the stored id, got %q", out)
	}
	if !strings.Contains(out, "client id:") {
		t.Errorf("status should label the id, got %q", out)
	}
}

func TestTelemetryOff_PersistsAndClearsID(t *testing.T) {
	dir := isolateConfig(t)
	config.EnsureClientID()

	out := runCmd(t, "off")
	if !strings.Contains(out, "telemetry disabled") {
		t.Errorf("off should confirm, got %q", out)
	}

	data, err := os.ReadFile(filepath.Join(dir, "unum", "config.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if cfg["telemetry"] != false {
		t.Errorf("telemetry should be persisted as false, got %v", cfg["telemetry"])
	}
	if _, ok := cfg["client_id"]; ok {
		t.Errorf("client id should be cleared, got %v", cfg["client_id"])
	}
}

func TestTelemetryOn_ReEnables(t *testing.T) {
	isolateConfig(t)

	runCmd(t, "off")
	out := runCmd(t, "on")
	if !strings.Contains(out, "telemetry enabled") {
		t.Errorf("on should confirm, got %q", out)
	}

	status := runCmd(t, "status")
	if !strings.Contains(status, "telemetry: enabled (set in config)") {
		t.Errorf("status should show the config source, got %q", status)
	}
}

func TestTelemetryOn_WarnsWhenEnvOverrides(t *testing.T) {
	isolateConfig(t)
	t.Setenv("DO_NOT_TRACK", "1")

	out := runCmd(t, "on")
	if !strings.Contains(out, "environment variable is still overriding") {
		t.Errorf("on should warn that the env still wins, got %q", out)
	}
}

func TestTelemetrySource(t *testing.T) {
	tests := []struct {
		name        string
		doNotTrack  string
		unumNoTelem string
		cfg         config.Config
		want        string
	}{
		{name: "default", want: "default"},
		{name: "do not track", doNotTrack: "yes", want: "DO_NOT_TRACK=yes"},
		{name: "unum var", unumNoTelem: "1", want: "UNUM_NO_TELEMETRY=1"},
		{name: "config", cfg: config.Config{Telemetry: new(bool)}, want: "set in config"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DO_NOT_TRACK", tt.doNotTrack)
			t.Setenv("UNUM_NO_TELEMETRY", tt.unumNoTelem)
			if got := telemetrySource(tt.cfg); got != tt.want {
				t.Errorf("telemetrySource()=%q, want %q", got, tt.want)
			}
		})
	}
}

func TestTelemetryBareCommand_ShowsStatus(t *testing.T) {
	isolateConfig(t)

	out := runCmd(t)
	if !strings.Contains(out, "telemetry:") {
		t.Errorf("bare command should show status, got %q", out)
	}
}
