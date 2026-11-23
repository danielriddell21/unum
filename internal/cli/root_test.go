package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestIsMetaCommand(t *testing.T) {
	root := &cobra.Command{Use: "unum"}
	telemetry := &cobra.Command{Use: "telemetry"}
	status := &cobra.Command{Use: "status"}
	completion := &cobra.Command{Use: "completion"}
	jsonCmd := &cobra.Command{Use: "json"}

	telemetry.AddCommand(status)
	root.AddCommand(telemetry, completion, jsonCmd)

	tests := []struct {
		name string
		cmd  *cobra.Command
		want bool
	}{
		{"telemetry", telemetry, true},
		{"telemetry subcommand", status, true},
		{"completion", completion, true},
		{"a tool", jsonCmd, false},
		{"root", root, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMetaCommand(tt.cmd); got != tt.want {
				t.Errorf("isMetaCommand(%s)=%v, want %v", tt.cmd.Name(), got, tt.want)
			}
		})
	}
}

func TestNoticeTextDisclosesTheClientID(t *testing.T) {
	// The notice is the only place a first-time user is told what is sent, so
	// "anonymous" is only defensible while the id is disclosed alongside it.
	for _, want := range []string{"randomly generated id", "turn telemetry off", "flag names"} {
		if !strings.Contains(noticeText, want) {
			t.Errorf("notice should mention %q", want)
		}
	}
}

func TestExecute_RunsACommand(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("APPDATA", dir)
	t.Setenv("DO_NOT_TRACK", "")
	t.Setenv("UNUM_NO_TELEMETRY", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })
	os.Args = []string{"unum", "telemetry", "status"}

	if err := Execute("v0.0.1-test"); err != nil {
		t.Fatalf("Execute: %v", err)
	}
}

func TestExecute_NoTelemetryFlagSkipsClientID(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("APPDATA", dir)
	t.Setenv("DO_NOT_TRACK", "")
	t.Setenv("UNUM_NO_TELEMETRY", "")

	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })
	os.Args = []string{"unum", "--no-telemetry", "hash", "--no-history", "--port", "svc"}

	if err := Execute("v0.0.1-test"); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "unum", "config.json")); !os.IsNotExist(err) {
		t.Error("--no-telemetry should not write a config or generate a client id")
	}
}

func TestExecute_UnknownCommandErrors(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("DO_NOT_TRACK", "1")

	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })
	os.Args = []string{"unum", "definitely-not-a-command"}

	if err := Execute("v0.0.1-test"); err == nil {
		t.Error("unknown command should return an error")
	}
}

type codedError int

func (c codedError) Error() string { return "coded" }
func (c codedError) ExitCode() int { return int(c) }

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"plain error", errors.New("boom"), 1},
		{"coded error", codedError(2), 2},
		{"wrapped coded error", fmt.Errorf("execute: %w", codedError(2)), 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ExitCode(tc.err); got != tc.want {
				t.Errorf("ExitCode = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestExecute_QueryFailureReturnsExitCodeTwo(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("DO_NOT_TRACK", "1")

	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })
	sample := filepath.Join("..", "..", "testdata", "sample.json")
	os.Args = []string{"unum", "json", sample, "--query", "!!invalid!!", "--no-color"}

	// Reaching the assertion at all proves Execute returned instead of
	// calling os.Exit, so its deferred telemetry flush ran.
	err := Execute("v0.0.1-test")
	if err == nil {
		t.Fatal("a failed query should return an error")
	}
	if got := ExitCode(err); got != 2 {
		t.Errorf("ExitCode = %d, want 2", got)
	}
}
