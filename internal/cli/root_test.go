package cli

import (
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
