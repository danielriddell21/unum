// Package cli wires together the root Cobra command and all subcommands.
package cli

import (
	"os"

	jsontool "github.com/danielriddell21/unum/internal/json"
	"github.com/spf13/cobra"
)

// Execute builds and runs the root command. Returns non-nil on error.
func Execute(version string) error {
	var (
		noColor bool
		quiet   bool
	)

	root := &cobra.Command{
		Use:   "unum",
		Short: "A nerdy all-in-one dev tool suite",
		Long: `unum (Latin: "one") — a collection of powerful developer tools with a nerdy twist.

Each tool is useful on its own, with a unique analytical or technical angle
baked in to make your daily workflow more interesting.`,
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
	root.AddCommand(jsontool.Command(&noColor, &quiet))

	return root.Execute()
}
