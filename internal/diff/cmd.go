// Package difftool implements the `unum diff` subcommand.
package difftool

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/unum/internal/config"
	"github.com/danielriddell21/unum/internal/diff/format"
	"github.com/danielriddell21/unum/internal/diff/node"
	"github.com/danielriddell21/unum/internal/diff/parse"
	"github.com/danielriddell21/unum/internal/diff/render/static"
	"github.com/danielriddell21/unum/internal/diff/render/tui"
	diffweb "github.com/danielriddell21/unum/internal/diff/render/web"
)

type flags struct {
	ui         bool
	web        bool
	format     string
	theme      string
	lightTheme string
	context    int
	stat       bool
	port       int
	quiet      bool
	noColor    bool
}

// Command returns the cobra command for `unum diff`.
func Command(globalNoColor *bool, globalQuiet *bool) *cobra.Command {
	f := &flags{}
	cfg := config.Load()
	f.theme = cfg.DarkTheme
	f.lightTheme = cfg.LightTheme

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
		Args:         cobra.RangeArgs(0, 2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f.noColor = f.noColor || *globalNoColor
			f.quiet = f.quiet || *globalQuiet
			if len(args) == 0 {
				if !f.web {
					return fmt.Errorf("requires two file arguments (or --web for browser input mode)")
				}
				return diffweb.StartServer(diffweb.Options{Port: f.port, Quiet: f.quiet, DarkTheme: f.theme, LightTheme: f.lightTheme})
			}
			return runDiff(f, args[0], args[1])
		},
	}

	cmd.Flags().BoolVar(&f.ui, "ui", false, "launch interactive TUI")
	cmd.Flags().BoolVar(&f.web, "web", false, "launch web UI in browser")
	cmd.Flags().IntVar(&f.port, "port", 0, "port for --web (default: random free port)")
	cmd.Flags().StringVar(&f.format, "format", "", "force format: text, json, yaml, terraform (default: auto)")
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

	bootDone := static.Boot(os.Stderr, fileA, fileB, opts)
	start := time.Now()

	var diff *node.Diff
	switch fmt_ {
	case node.FormatJSON:
		diff, err = parse.JSON(dataA, dataB)
	case node.FormatYAML:
		diff, err = parse.YAML(dataA, dataB)
	case node.FormatTerraform:
		diff, err = parse.Terraform(dataA)
	default:
		diff, err = parse.Text(dataA, dataB, f.context)
	}
	if err != nil {
		return fmt.Errorf("diff: %w", err)
	}
	bootDone(diff.Added, diff.Removed, diff.Modified, time.Since(start))
	diff.FileA = fileA
	diff.FileB = fileB
	diff.Format = fmt_

	if f.ui {
		if err := tui.Start(diff, f.theme); err != nil {
			return fmt.Errorf("diff TUI: %w", err)
		}
		return nil
	}
	if f.web {
		if err := diffweb.Start(diff, diffweb.Options{Port: f.port, Quiet: f.quiet, DarkTheme: f.theme, LightTheme: f.lightTheme}); err != nil {
			return fmt.Errorf("diff web: %w", err)
		}
		return nil
	}

	if err := static.Render(os.Stdout, diff, opts); err != nil {
		return fmt.Errorf("diff render: %w", err)
	}
	return nil
}
