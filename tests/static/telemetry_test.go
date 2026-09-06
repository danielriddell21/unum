package static_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTelemetry_StatusExitZero(t *testing.T) {
	stdout, _, code := run("telemetry", "status")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if !strings.Contains(stdout, "telemetry:") {
		t.Errorf("status should report state, got %q", stdout)
	}
	if !strings.Contains(stdout, "endpoint:") {
		t.Errorf("status should report the endpoint, got %q", stdout)
	}
}

func TestTelemetry_StatusReportsEnvOptOut(t *testing.T) {
	stdout, _, code := runEnv([]string{"DO_NOT_TRACK=yes"}, "telemetry", "status")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if !strings.Contains(stdout, "disabled") {
		t.Errorf("DO_NOT_TRACK should disable telemetry, got %q", stdout)
	}
	if !strings.Contains(stdout, "DO_NOT_TRACK=yes") {
		t.Errorf("status should name the source of the setting, got %q", stdout)
	}
}

func TestTelemetry_NoColor(t *testing.T) {
	stdout, _, code := run("telemetry", flagNoColor, "status")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if strings.Contains(stdout, "\x1b[") {
		t.Error("output should not contain ANSI escape codes with --no-color")
	}
}

func TestTelemetry_OffThenOnPersists(t *testing.T) {
	dir := t.TempDir()
	env := []string{"XDG_CONFIG_HOME=" + dir, "APPDATA=" + dir, "DO_NOT_TRACK="}

	if _, _, code := runEnv(env, "telemetry", "off"); code != 0 {
		t.Fatalf("telemetry off: exit %d", code)
	}
	stdout, _, _ := runEnv(env, "telemetry", "status")
	if !strings.Contains(stdout, "disabled") {
		t.Errorf("after off, status should be disabled, got %q", stdout)
	}

	if _, _, code := runEnv(env, "telemetry", "on"); code != 0 {
		t.Fatalf("telemetry on: exit %d", code)
	}
	stdout, _, _ = runEnv(env, "telemetry", "status")
	if !strings.Contains(stdout, "enabled") {
		t.Errorf("after on, status should be enabled, got %q", stdout)
	}
}

func TestTelemetry_NothingWrittenWhenOptedOut(t *testing.T) {
	dir := t.TempDir()
	env := []string{"XDG_CONFIG_HOME=" + dir, "APPDATA=" + dir, "DO_NOT_TRACK=1"}

	runEnv(env, "hash", "some-service")

	if _, err := os.Stat(filepath.Join(dir, "unum", "config.json")); !os.IsNotExist(err) {
		t.Error("no config should be written for a user who opted out before first run")
	}
}

func TestTelemetry_FirstRunNoticeShownOnce(t *testing.T) {
	dir := t.TempDir()
	env := []string{"XDG_CONFIG_HOME=" + dir, "APPDATA=" + dir, "DO_NOT_TRACK="}

	_, stderr, code := runEnv(env, "hash", "first-run")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if !strings.Contains(stderr, "telemetry") {
		t.Errorf("first run should print the telemetry notice, got %q", stderr)
	}

	_, stderr2, _ := runEnv(env, "hash", "second-run")
	if strings.Contains(stderr2, "This notice is shown once") {
		t.Errorf("notice should not repeat, got %q", stderr2)
	}
}

func TestTelemetry_UnknownSubcommandExitsNonZero(t *testing.T) {
	_, _, code := run("telemetry", "bogus")
	if code == 0 {
		t.Error("unknown subcommand should exit non-zero")
	}
}
