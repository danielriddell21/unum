// Package cli wires together the root Cobra command and all subcommands.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	difftool "github.com/danielriddell21/unum/internal/diff"
	hashtool "github.com/danielriddell21/unum/internal/hash"
	jsontool "github.com/danielriddell21/unum/internal/json"
)

// Execute builds and runs the root command. Returns non-nil on error.
func Execute(version string) error {
	var (
		noColor bool
		quiet   bool
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
		},
	}

	root.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output (also honors NO_COLOR env var)")
	root.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress informational output")

	// Register subcommands
	root.AddCommand(jsontool.Command(&noColor, &quiet, version))
	root.AddCommand(difftool.Command(&noColor, &quiet, version))
	root.AddCommand(hashtool.Command(&noColor, &quiet, version))
	root.AddCommand(completionCmd())

	if err := root.Execute(); err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return nil
}
