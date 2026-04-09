// Package difftool implements the `unum diff` subcommand.
package difftool

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/unum/internal/config"
	"github.com/danielriddell21/unum/internal/diff/format"
	"github.com/danielriddell21/unum/internal/diff/parse"
	"github.com/danielriddell21/unum/internal/diff/render/static"
)

type flags struct {
	ui      bool
	web     bool
	format  string
	theme   string
	context int
	stat    bool
	port    int
	quiet   bool
	noColor bool
}

// Command returns the cobra command for `unum diff`.
func Command(globalNoColor *bool, globalQuiet *bool) *cobra.Command {
	f := &flags{}
	cfg := config.Load()

	cmd := &cobra.Command{
		Use:   "diff <file-a> <file-b>",
		Short: "Visualise the diff between two files",
		Long: `Compare two files and render a coloured, navigable diff.

Format is auto-detected from the file extension:
  .json          → semantic JSON diff
  .yaml / .yml   → semantic YAML diff
  .tfplan        → Terraform plan diff
  (anything else) → generic line-by-line diff

Output modes:
  (default)  Syntax-highlighted unified diff
  --ui       Interactive TUI (unified + split views, v to toggle)
  --web      Browser-based diff viewer`,
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f.noColor = *globalNoColor
			f.quiet = *globalQuiet
			return runDiff(f, args[0], args[1])
		},
	}

	cmd.Flags().BoolVar(&f.ui, "ui", false, "launch interactive TUI")
	cmd.Flags().BoolVar(&f.web, "web", false, "launch web UI in browser")
	cmd.Flags().IntVar(&f.port, "port", 0, "port for --web (default: random free port)")
	cmd.Flags().StringVar(&f.format, "format", "", "force format: text, json, yaml, terraform (default: auto)")
	cmd.Flags().StringVar(&f.theme, "theme", cfg.Theme, "color theme: cyber, matrix, dracula, nord")
	cmd.Flags().IntVar(&f.context, "context", 3, "lines of context around each change")
	cmd.Flags().BoolVar(&f.stat, "stat", false, "show summary only (no diff body)")
	cmd.Flags().BoolVar(&f.quiet, "quiet", false, "suppress the boot line")

	return cmd
}

func runDiff(f *flags, fileA, fileB string) error {
	dataA, err := os.ReadFile(fileA)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", fileA, err)
	}
	dataB, err := os.ReadFile(fileB)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", fileB, err)
	}

	fmt_ := format.Parse(f.format, fileA, fileB)

	theme := static.ResolveTheme(f.theme)
	opts := static.Options{
		Theme:   theme,
		Context: f.context,
		Stat:    f.stat,
		NoColor: f.noColor,
		Quiet:   f.quiet,
	}

	static.Boot(os.Stderr, fileA, fileB, opts)

	diff, err := parse.Text(dataA, dataB, f.context)
	if err != nil {
		return fmt.Errorf("diff: %w", err)
	}
	diff.FileA = fileA
	diff.FileB = fileB
	diff.Format = fmt_

	if f.ui {
		fmt.Fprintln(os.Stderr, "TUI not yet implemented — coming soon")
		return nil
	}
	if f.web {
		fmt.Fprintln(os.Stderr, "web UI not yet implemented — coming soon")
		return nil
	}

	return static.Render(os.Stdout, diff, opts)
}
