package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/unum/internal/config"
	"github.com/danielriddell21/unum/internal/telemetry"
)

func telemetryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "telemetry",
		Short: "Inspect and control usage telemetry",
		Args:  cobra.NoArgs,
		Long: `Show what unum collects and turn it on or off.

Telemetry is anonymous usage data — tool name, output mode, OS, flag names,
input size, and duration. It never includes file contents, filenames, paths,
or flag values.

A randomly generated client id is sent alongside it so runs from one install
can be counted together. It identifies nothing about you or your machine, and
"telemetry off" deletes it. See PRIVACY.md for the full list.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTelemetryStatus(cmd)
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:          "status",
		Short:        "Show whether telemetry is enabled and why",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTelemetryStatus(cmd)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:          "on",
		Short:        "Enable telemetry and persist the choice",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return setTelemetry(cmd, true)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:          "off",
		Short:        "Disable telemetry and persist the choice",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return setTelemetry(cmd, false)
		},
	})

	return cmd
}

func runTelemetryStatus(cmd *cobra.Command) error {
	cfg := config.Load()
	out := cmd.OutOrStdout()

	state := "disabled"
	if cfg.TelemetryEnabled() {
		state = "enabled"
	}
	fmt.Fprintf(out, "telemetry: %s (%s)\n", state, telemetrySource(cfg))

	endpoint := telemetry.Endpoint()
	if endpoint == "" {
		endpoint = "none — nothing is sent from this build"
	}
	fmt.Fprintf(out, "endpoint:  %s\n", endpoint)

	if cfg.ClientID == "" {
		fmt.Fprintln(out, "client id: none stored")
	} else {
		fmt.Fprintf(out, "client id: %s\n", cfg.ClientID)
	}

	path, err := config.Path()
	if err == nil {
		fmt.Fprintf(out, "config:    %s\n", path)
	}
	return nil
}

func telemetrySource(cfg config.Config) string {
	if v := os.Getenv("DO_NOT_TRACK"); v != "" {
		return "DO_NOT_TRACK=" + v
	}
	if v := os.Getenv("UNUM_NO_TELEMETRY"); v != "" {
		return "UNUM_NO_TELEMETRY=" + v
	}
	if cfg.Telemetry != nil {
		return "set in config"
	}
	return "default"
}

func setTelemetry(cmd *cobra.Command, on bool) error {
	cfg := config.Load()
	cfg.Telemetry = &on
	if !on {
		cfg.ClientID = ""
	}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	out := cmd.OutOrStdout()
	if on {
		fmt.Fprintln(out, "telemetry enabled")
	} else {
		fmt.Fprintln(out, "telemetry disabled; stored client id removed")
	}

	if !cfg.TelemetryEnabled() && on {
		fmt.Fprintln(out, "note: an environment variable is still overriding this for the current shell")
	}
	return nil
}
