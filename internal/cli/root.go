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
	jsontool "github.com/danielriddell21/unum/internal/json"
	"github.com/danielriddell21/unum/internal/telemetry"
)

func Execute(version string) error {
	cfg := config.EnsureClientID()
	tel := telemetry.Init(cfg, "unum", version)
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
				tel.Disable()
			}
		},
	}

	root.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output (also honors NO_COLOR env var)")
	root.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress informational output")
	root.PersistentFlags().BoolVar(&noTelemetry, "no-telemetry", false, "disable anonymous usage telemetry (also: DO_NOT_TRACK=1)")

	// Register subcommands
	root.AddCommand(jsontool.Command(&noColor, &quiet, version, tel))
	root.AddCommand(difftool.Command(&noColor, &quiet, version, tel))
	root.AddCommand(hashtool.Command(&noColor, &quiet, version, tel))
	root.AddCommand(diagramtool.Command(&noColor, &quiet, version, tel))
	root.AddCommand(completionCmd())

	if err := root.Execute(); err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return nil
}
