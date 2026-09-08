package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/unum/internal/config"
	diagramtool "github.com/danielriddell21/unum/internal/diagram"
	difftool "github.com/danielriddell21/unum/internal/diff"
	hashtool "github.com/danielriddell21/unum/internal/hash"
	imagetool "github.com/danielriddell21/unum/internal/image"
	jsontool "github.com/danielriddell21/unum/internal/json"
	"github.com/danielriddell21/unum/internal/telemetry"
)

const noticeText = `unum sends anonymous usage telemetry: tool name, output mode, OS, flag names,
input size, and duration. Never file contents, names, paths, or flag values.

It includes a randomly generated id so one install's runs can be counted
together. It is not derived from you or your machine, and is deleted when you
turn telemetry off.

  Turn it off:  unum telemetry off   (or --no-telemetry, or DO_NOT_TRACK=1)
  Full detail:  https://github.com/danielriddell21/unum/blob/trunk/PRIVACY.md

This notice is shown once.
`

func Execute(version string) error {
	tel := telemetry.New()
	defer tel.Close(context.Background()) //nolint:errcheck // best-effort flush on exit

	var (
		noColor     bool
		quiet       bool
		noTelemetry bool
	)

	root := &cobra.Command{
		Use:   "unum",
		Short: "A unified developer toolkit",
		Long: `unum — a suite of focused developer tools for analysis, comparison, and identification.

Each tool is composable and works standalone, with consistent output modes
across terminal, TUI, and web interfaces.`,
		Version:      version,
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if os.Getenv("NO_COLOR") != "" {
				noColor = true
			}
			if noTelemetry {
				return
			}
			// Telemetry starts here, not in Execute, so that --no-telemetry is
			// known before any client ID is generated or endpoint contacted.
			cfg := config.EnsureClientID()
			if !cfg.TelemetryEnabled() {
				return
			}
			if !cfg.NoticeShown && !quiet && !isMetaCommand(cmd) {
				fmt.Fprint(os.Stderr, noticeText)
				config.MarkNoticeShown(cfg)
			}
			tel.Start(cfg, "unum", version)
		},
	}

	root.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output (also honors NO_COLOR env var)")
	root.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress informational output")
	root.PersistentFlags().BoolVar(&noTelemetry, "no-telemetry", false, "disable usage telemetry for this run (also: DO_NOT_TRACK=1)")

	// Register subcommands
	root.AddCommand(jsontool.Command(&noColor, &quiet, version, tel))
	root.AddCommand(difftool.Command(&noColor, &quiet, version, tel))
	root.AddCommand(hashtool.Command(&noColor, &quiet, version, tel))
	root.AddCommand(diagramtool.Command(&noColor, &quiet, version, tel))
	root.AddCommand(imagetool.Command(&noColor, &quiet, version, tel))
	root.AddCommand(telemetryCmd())
	root.AddCommand(completionCmd())

	if err := root.Execute(); err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return nil
}

func isMetaCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "telemetry" || c.Name() == "completion" {
			return true
		}
	}
	return false
}
