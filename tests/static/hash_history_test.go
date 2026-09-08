package static_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func historyEnv(t *testing.T) (dir string, env []string) {
	t.Helper()
	dir = t.TempDir()
	return dir, []string{"XDG_CONFIG_HOME=" + dir, "APPDATA=" + dir}
}

func historyFile(dir string) string {
	return filepath.Join(dir, "unum", "hash-history.json")
}

func TestHashHistory_RecordedByDefault(t *testing.T) {
	dir, env := historyEnv(t)

	if _, _, code := runEnv(env, "hash", "recorded-service"); code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}

	data, err := os.ReadFile(historyFile(dir))
	if err != nil {
		t.Fatalf("read history: %v", err)
	}
	if !strings.Contains(string(data), "recorded-service") {
		t.Errorf("history should contain the input, got %s", data)
	}
}

func TestHashHistory_NoHistoryFlag(t *testing.T) {
	dir, env := historyEnv(t)

	if _, _, code := runEnv(env, "hash", "--no-history", "secret-value"); code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}

	if _, err := os.Stat(historyFile(dir)); !os.IsNotExist(err) {
		t.Error("--no-history should not create the history file")
	}
}

func TestHashHistory_EnvOptOut(t *testing.T) {
	dir, env := historyEnv(t)
	env = append(env, "UNUM_NO_HASH_HISTORY=1")

	if _, _, code := runEnv(env, "hash", "secret-value"); code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}

	if _, err := os.Stat(historyFile(dir)); !os.IsNotExist(err) {
		t.Error("UNUM_NO_HASH_HISTORY should not create the history file")
	}
}

func TestHashHistory_ClearHistory(t *testing.T) {
	dir, env := historyEnv(t)

	runEnv(env, "hash", "one")
	runEnv(env, "hash", "two")
	if _, err := os.Stat(historyFile(dir)); err != nil {
		t.Fatalf("history should exist before clearing: %v", err)
	}

	stdout, _, code := runEnv(env, "hash", "--clear-history")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if !strings.Contains(stdout, "cleared") {
		t.Errorf("clear should confirm on stdout, got %q", stdout)
	}
	if _, err := os.Stat(historyFile(dir)); !os.IsNotExist(err) {
		t.Error("--clear-history should remove the history file")
	}
}

func TestHashHistory_ClearWhenAbsent(t *testing.T) {
	_, env := historyEnv(t)

	if _, _, code := runEnv(env, "hash", "--clear-history"); code != 0 {
		t.Errorf("clearing an absent history should exit 0, got %d", code)
	}
}

func TestHashHistory_ClearNoColor(t *testing.T) {
	_, env := historyEnv(t)

	stdout, _, code := runEnv(env, "hash", flagNoColor, "--clear-history")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if strings.Contains(stdout, "\x1b[") {
		t.Error("output should not contain ANSI escape codes with --no-color")
	}
}

func TestHashHistory_StillDerivesWithNoHistory(t *testing.T) {
	_, env := historyEnv(t)

	stdout, _, code := runEnv(env, "hash", "--no-history", "--port", "some-service")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Error("--no-history should not affect derivation output")
	}
}
