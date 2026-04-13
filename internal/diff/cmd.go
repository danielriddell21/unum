// Package difftool implements the `unum diff` subcommand.
package difftool

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/danielriddell21/unum/internal/config"
	"github.com/danielriddell21/unum/internal/diff/format"
	"github.com/danielriddell21/unum/internal/diff/node"
	"github.com/danielriddell21/unum/internal/diff/parse"
	"github.com/danielriddell21/unum/internal/diff/render/static"
	"github.com/danielriddell21/unum/internal/diff/render/tui"
	diffweb "github.com/danielriddell21/unum/internal/diff/render/web"
	"github.com/danielriddell21/unum/internal/telemetry"
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
	version    string
	tel        *telemetry.Telemetry
}

// Command returns the cobra command for `unum diff`.
func Command(globalNoColor *bool, globalQuiet *bool, version string, tel *telemetry.Telemetry) *cobra.Command {
	f := &flags{}
	f.version = version
	f.tel = tel
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
				_, span := f.tel.Tracer().Start(context.Background(), "diff.execute",
					trace.WithAttributes(
						attribute.String("tool", "diff"),
						attribute.String("mode", "web"),
						attribute.String("os", runtime.GOOS),
						attribute.String("arch", runtime.GOARCH),
					),
				)
				defer span.End()
				return diffweb.StartServer(diffweb.Options{Port: f.port, Quiet: f.quiet, DarkTheme: f.theme, LightTheme: f.lightTheme, Version: f.version})
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
	mode := "cli"
	if f.ui {
		mode = "tui"
	} else if f.web {
		mode = "web"
	}
	_, span := f.tel.Tracer().Start(context.Background(), "diff.execute",
		trace.WithAttributes(
			attribute.String("tool", "diff"),
			attribute.String("mode", mode),
			attribute.String("os", runtime.GOOS),
			attribute.String("arch", runtime.GOARCH),
			attribute.StringSlice("flags", activeDiffFlags(f)),
		),
	)
	defer span.End()

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
		if err := tui.Start(diff, f.theme, f.version); err != nil {
			return fmt.Errorf("diff TUI: %w", err)
		}
		return nil
	}
	if f.web {
		if err := diffweb.Start(diff, diffweb.Options{Port: f.port, Quiet: f.quiet, DarkTheme: f.theme, LightTheme: f.lightTheme, Version: f.version}); err != nil {
			return fmt.Errorf("diff web: %w", err)
		}
		return nil
	}

	if err := static.Render(os.Stdout, diff, opts); err != nil {
		return fmt.Errorf("diff render: %w", err)
	}
	return nil
}

func activeDiffFlags(f *flags) []string {
	var flags []string
	if f.format != "" {
		flags = append(flags, "format")
	}
	if f.stat {
		flags = append(flags, "stat")
	}
	if f.context != 3 {
		flags = append(flags, "context")
	}
	return flags
}
