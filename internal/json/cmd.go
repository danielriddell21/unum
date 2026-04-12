// Package jsontool implements the `unum json` subcommand.
package jsontool

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/unum/internal/config"
	"github.com/danielriddell21/unum/internal/json/analyze"
	"github.com/danielriddell21/unum/internal/json/lens/merkle"
	"github.com/danielriddell21/unum/internal/json/lens/query"
	"github.com/danielriddell21/unum/internal/json/lens/stats"
	"github.com/danielriddell21/unum/internal/json/lens/transform"
	"github.com/danielriddell21/unum/internal/json/lens/typegen"
	"github.com/danielriddell21/unum/internal/json/parse"
	"github.com/danielriddell21/unum/internal/json/render/static"
	"github.com/danielriddell21/unum/internal/json/render/tui"
	"github.com/danielriddell21/unum/internal/json/render/web"
)

// flags holds all flag values for the json subcommand.
type flags struct {
	// Output modes
	ui  bool
	web bool

	// Static output options
	validateOnly bool
	compact      bool
	theme        string
	lightTheme   string

	// Annotation lenses
	showStats  bool
	showMerkle bool
	hashOnly   bool

	// Output-mode lenses (mutually exclusive with each other)
	transformYAML bool
	typegenTarget string
	typegenPkg    string
	typegenType   string
	jqExpr        string

	// Web options
	port int

	// Global
	quiet   bool
	noColor bool
}

// Command returns the cobra command for `unum json`.
func Command(globalNoColor *bool, globalQuiet *bool) *cobra.Command {
	f := &flags{}
	cfg := config.Load()
	f.theme = cfg.DarkTheme
	f.lightTheme = cfg.LightTheme

	cmd := &cobra.Command{
		Use:   "json [file]",
		Short: "View, validate, and analyze JSON",
		Long: `Parse, validate, and explore JSON files with multiple analysis lenses.

Three output modes:
  (default)  Syntax-highlighted pretty-print with optional annotations
  --ui       Interactive TUI navigator (lazygit-style 3-panel layout)
  --web      Local web server with browser-based explorer

Omit [file] with --ui or --web to open an interactive file picker.

Config file (~/.config/unum/config.json):
  { "dark_theme": "cyber", "light_theme": "clean", "mode": "dark" }`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.noColor = f.noColor || *globalNoColor
			f.quiet = f.quiet || *globalQuiet
			if len(args) == 0 {
				return runJSONNoFile(f)
			}
			return runJSON(f, args[0])
		},
		SilenceUsage: true,
	}

	// Output modes
	cmd.Flags().BoolVar(&f.ui, "ui", false, "launch interactive TUI navigator")
	cmd.Flags().BoolVar(&f.web, "web", false, "launch web UI in browser")
	cmd.Flags().IntVar(&f.port, "port", 0, "port for --web (default: random free port)")

	// Static output
	cmd.Flags().BoolVar(&f.validateOnly, "validate-only", false, "exit 0/1 based on validity only")
	cmd.Flags().BoolVar(&f.compact, "compact", false, "output minified JSON")


	// Annotation lenses
	cmd.Flags().BoolVar(&f.showStats, "stats", false, "annotate numeric arrays with statistics")
	cmd.Flags().BoolVar(&f.showMerkle, "merkle", false, "annotate nodes with SHA256 hashes")
	cmd.Flags().BoolVar(&f.hashOnly, "hash-only", false, "print root merkle hash only (requires --merkle)")

	// Output-mode lenses
	cmd.Flags().BoolVar(&f.transformYAML, "transform", false, "convert to YAML (use: --transform yaml)")
	cmd.Flags().StringVar(&f.typegenTarget, "typegen", "", "generate types: go, ts, jsonschema")
	cmd.Flags().StringVar(&f.typegenPkg, "pkg", "", "Go package name for --typegen go")
	cmd.Flags().StringVar(&f.typegenType, "type", "", "root type name for --typegen")
	cmd.Flags().StringVar(&f.jqExpr, "query", "", "jq expression to evaluate")

	// Utility
	cmd.Flags().BoolVar(&f.quiet, "quiet", false, "suppress informational output")

	return cmd
}

func runJSON(f *flags, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", filename, err)
	}

	// Strip UTF-8 BOM — some editors (Notepad, VS) add \xEF\xBB\xBF.
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	theme := resolveStaticTheme(f.theme)
	renderOpts := static.Options{
		Theme:        theme,
		ShowLineNums: true,
		Compact:      f.compact,
		Quiet:        f.quiet,
		NoColor:      f.noColor,
		Filename:     filename,
	}

	// --validate-only: no boot line, no render — just exit code.
	if f.validateOnly {
		if err := parse.Validate(data); err != nil {
			static.RenderError(os.Stderr, err, renderOpts)
			os.Exit(1)
		}
		return nil
	}

	start := time.Now()
	done := static.Boot(os.Stderr, filename, renderOpts)

	if err := parse.Validate(data); err != nil {
		fmt.Fprintln(os.Stderr)
		static.RenderError(os.Stderr, err, renderOpts)
		os.Exit(1)
	}

	root, err := parse.Parse(data)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	nodeCount := root.CountNodes()
	done(nodeCount, time.Since(start))

	ctx := context.Background()

	// ── Output-mode lenses (mutually exclusive) ───────────────────────────────

	if f.jqExpr != "" {
		result, err := query.Execute(root, f.jqExpr)
		if err != nil {
			static.RenderError(os.Stderr, err, renderOpts)
			os.Exit(2)
		}
		fmt.Println(result)
		return nil
	}

	if f.typegenTarget != "" {
		result, err := typegen.Generate(root, typegen.Options{
			Target:      typegen.Target(f.typegenTarget),
			PackageName: f.typegenPkg,
			TypeName:    f.typegenType,
		})
		if err != nil {
			return fmt.Errorf("typegen: %w", err)
		}
		fmt.Print(result)
		return nil
	}

	if f.transformYAML {
		yamlBytes, err := transform.ToYAML(root)
		if err != nil {
			return fmt.Errorf("transform: %w", err)
		}
		fmt.Print(string(yamlBytes))
		return nil
	}

	// ── Annotation lenses ─────────────────────────────────────────────────────

	analyzeOpts := analyze.Options{
		RunStats:  f.showStats,
		RunMerkle: f.showMerkle || f.hashOnly,
	}

	allAnalyzers := []analyze.Analyzer{
		stats.Analyzer{},
		merkle.Analyzer{},
	}
	suite := analyze.Build(analyzeOpts, allAnalyzers)
	if err := suite.Run(ctx, root, analyzeOpts); err != nil {
		return fmt.Errorf("analyze: %w", err)
	}

	if f.hashOnly {
		fmt.Println(merkle.RootHash(root))
		return nil
	}

	// ── TUI / Web modes ───────────────────────────────────────────────────────

	if f.ui {
		return tui.Start(root, filename, f.theme)
	}

	if f.web {
		return web.Start(root, web.Options{
			Port:       f.port,
			Filename:   filename,
			Quiet:      f.quiet,
			DarkTheme:  f.theme,
			LightTheme: f.lightTheme,
		})
	}

	// ── Default: static render ────────────────────────────────────────────────

	renderOpts.ShowStats = f.showStats
	renderOpts.ShowMerkle = f.showMerkle
	return static.Render(os.Stdout, root, renderOpts)
}

func runJSONNoFile(f *flags) error {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	if f.ui {
		return tui.StartWithPicker(cwd, f.theme)
	}
	if f.web {
		return web.StartBrowser(web.Options{Port: f.port, Quiet: f.quiet, DarkTheme: f.theme, LightTheme: f.lightTheme})
	}
	return fmt.Errorf("missing argument: <file> (or use --ui / --web to pick interactively)")
}

func resolveStaticTheme(name string) static.Theme {
	switch name {
	case "matrix":
		return static.Matrix
	case "dracula":
		return static.Dracula
	case "nord":
		return static.Nord
	default:
		return static.Cyber
	}
}
