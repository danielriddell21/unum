// Package jsontool implements the `unum json` subcommand.
package jsontool

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	ometric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

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
	"github.com/danielriddell21/unum/internal/telemetry"
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
	version string
	tel     *telemetry.Telemetry
}

// Command returns the cobra command for `unum json`.
func Command(globalNoColor *bool, globalQuiet *bool, version string, tel *telemetry.Telemetry) *cobra.Command {
	f := &flags{}
	f.version = version
	f.tel = tel
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

func runJSON(f *flags, filename string) error { //nolint:cyclop,gocognit // NOSONAR: dispatch on output flags + render modes; each branch is a distinct user-facing mode
	mode := "cli"
	if f.ui {
		mode = "tui"
	} else if f.web {
		mode = "web"
	}
	ctx, span := f.tel.Tracer().Start(context.Background(), "json.execute",
		trace.WithAttributes(
			attribute.String("tool", "json"),
			attribute.String("mode", mode),
			attribute.String("os", runtime.GOOS),
			attribute.String("arch", runtime.GOARCH),
			attribute.StringSlice("flags", activeJSONFlags(f)),
		),
	)
	defer span.End()

	data, err := os.ReadFile(filename)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "read")
		f.tel.M.Errors.Add(ctx, 1, ometric.WithAttributes(
			attribute.String("tool", "json"),
			attribute.String("error_type", "read"),
		))
		return fmt.Errorf("cannot read %s: %w", filename, err)
	}

	// Strip UTF-8 BOM — some editors (Notepad, VS) add \xEF\xBB\xBF.
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	theme := static.ResolveTheme(f.theme)
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
			span.RecordError(err)
			span.SetStatus(codes.Error, "validate")
			f.tel.M.Errors.Add(ctx, 1, ometric.WithAttributes(
				attribute.String("tool", "json"),
				attribute.String("error_type", "parse"),
			))
			span.End()
			os.Exit(1) //nolint:gocritic // validate-only mode requires hard exit for CI tooling
		}
		return nil
	}

	start := time.Now()
	done := static.Boot(os.Stderr, filename, renderOpts)

	if err := parse.Validate(data); err != nil {
		fmt.Fprintln(os.Stderr)
		static.RenderError(os.Stderr, err, renderOpts)
		span.RecordError(err)
		span.SetStatus(codes.Error, "validate")
		f.tel.M.Errors.Add(ctx, 1, ometric.WithAttributes(
			attribute.String("tool", "json"),
			attribute.String("error_type", "parse"),
		))
		span.End()
		os.Exit(1) //nolint:gocritic // validation failure requires hard exit for CI tooling
	}

	_, parseSpan := f.tel.Tracer().Start(ctx, "json.parse")
	root, err := parse.Parse(data)
	if err != nil {
		parseSpan.RecordError(err)
		parseSpan.SetStatus(codes.Error, err.Error())
		parseSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, "parse")
		f.tel.M.Errors.Add(ctx, 1, ometric.WithAttributes(
			attribute.String("tool", "json"),
			attribute.String("error_type", "parse"),
		))
		return fmt.Errorf("parse: %w", err)
	}
	parseSpan.End()

	nodeCount := root.CountNodes()
	done(nodeCount, time.Since(start))
	span.SetAttributes(
		attribute.Int("json.size_bytes", len(data)),
		attribute.Int("json.node_count", nodeCount),
		attribute.String("json.root_kind", root.Kind.String()),
	)
	f.tel.M.JSONNodes.Record(ctx, int64(nodeCount))

	// ── Output-mode lenses (mutually exclusive) ───────────────────────────────

	if f.jqExpr != "" {
		result, err := query.Execute(root, f.jqExpr)
		if err != nil {
			static.RenderError(os.Stderr, err, renderOpts)
			span.RecordError(err)
			span.SetStatus(codes.Error, "query")
			f.tel.M.Errors.Add(ctx, 1, ometric.WithAttributes(
				attribute.String("tool", "json"),
				attribute.String("error_type", "query"),
			))
			span.End()
			os.Exit(2) //nolint:gocritic // query failure requires hard exit with distinct code
		}
		recordSuccess(ctx, f, mode, len(data), time.Since(start))
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
			span.RecordError(err)
			span.SetStatus(codes.Error, "typegen")
			f.tel.M.Errors.Add(ctx, 1, ometric.WithAttributes(
				attribute.String("tool", "json"),
				attribute.String("error_type", "render"),
			))
			return fmt.Errorf("typegen: %w", err)
		}
		recordSuccess(ctx, f, mode, len(data), time.Since(start))
		fmt.Print(result)
		return nil
	}

	if f.transformYAML {
		yamlBytes, err := transform.ToYAML(root)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "transform")
			f.tel.M.Errors.Add(ctx, 1, ometric.WithAttributes(
				attribute.String("tool", "json"),
				attribute.String("error_type", "render"),
			))
			return fmt.Errorf("transform: %w", err)
		}
		recordSuccess(ctx, f, mode, len(data), time.Since(start))
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
	_, analyzeSpan := f.tel.Tracer().Start(ctx, "json.analyze")
	if err := suite.Run(ctx, root, analyzeOpts); err != nil {
		analyzeSpan.RecordError(err)
		analyzeSpan.SetStatus(codes.Error, err.Error())
		analyzeSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, "analyze")
		f.tel.M.Errors.Add(ctx, 1, ometric.WithAttributes(
			attribute.String("tool", "json"),
			attribute.String("error_type", "analyze"),
		))
		return fmt.Errorf("analyze: %w", err)
	}
	analyzeSpan.End()

	if f.hashOnly {
		recordSuccess(ctx, f, mode, len(data), time.Since(start))
		fmt.Println(merkle.RootHash(root))
		return nil
	}

	// ── TUI / Web modes ───────────────────────────────────────────────────────

	if f.ui {
		recordSuccess(ctx, f, mode, len(data), time.Since(start))
		if err := tui.Start(root, filename, f.theme, f.version); err != nil {
			return fmt.Errorf("json TUI: %w", err)
		}
		return nil
	}

	if f.web {
		recordSuccess(ctx, f, mode, len(data), time.Since(start))
		if err := web.Start(root, web.Options{
			Port:       f.port,
			Filename:   filename,
			Quiet:      f.quiet,
			DarkTheme:  f.theme,
			LightTheme: f.lightTheme,
			Version:    f.version,
			Tel:        f.tel,
		}); err != nil {
			return fmt.Errorf("json web: %w", err)
		}
		return nil
	}

	// ── Default: static render ────────────────────────────────────────────────

	renderOpts.ShowStats = f.showStats
	renderOpts.ShowMerkle = f.showMerkle
	if err := static.Render(os.Stdout, root, renderOpts); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "render")
		f.tel.M.Errors.Add(ctx, 1, ometric.WithAttributes(
			attribute.String("tool", "json"),
			attribute.String("error_type", "render"),
		))
		return fmt.Errorf("json render: %w", err)
	}
	recordSuccess(ctx, f, mode, len(data), time.Since(start))
	return nil
}

func recordSuccess(ctx context.Context, f *flags, mode string, sizeBytes int, elapsed time.Duration) {
	f.tel.M.Invocations.Add(ctx, 1, ometric.WithAttributes(
		attribute.String("tool", "json"),
		attribute.String("mode", mode),
		attribute.String("os", runtime.GOOS),
		attribute.String("arch", runtime.GOARCH),
		attribute.String("version", f.version),
	))
	f.tel.M.InputBytes.Record(ctx, int64(sizeBytes), ometric.WithAttributes(
		attribute.String("tool", "json"),
	))
	f.tel.M.Duration.Record(ctx, elapsed.Seconds(), ometric.WithAttributes(
		attribute.String("tool", "json"),
		attribute.String("mode", mode),
	))
}

func runJSONNoFile(f *flags) error {
	mode := "tui"
	if f.web {
		mode = "web"
	}
	_, span := f.tel.Tracer().Start(context.Background(), "json.execute",
		trace.WithAttributes(
			attribute.String("tool", "json"),
			attribute.String("mode", mode),
			attribute.String("os", runtime.GOOS),
			attribute.String("arch", runtime.GOARCH),
		),
	)
	defer span.End()

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	if f.ui {
		if err := tui.StartWithPicker(cwd, f.theme, f.version); err != nil {
			return fmt.Errorf("json TUI: %w", err)
		}
		return nil
	}
	if f.web {
		if err := web.StartServer(web.Options{Port: f.port, Quiet: f.quiet, DarkTheme: f.theme, LightTheme: f.lightTheme, Version: f.version, Tel: f.tel}); err != nil {
			return fmt.Errorf("json web: %w", err)
		}
		return nil
	}
	return fmt.Errorf("missing argument: <file> (or use --ui / --web to pick interactively)")
}

func activeJSONFlags(f *flags) []string {
	var flags []string
	if f.showStats {
		flags = append(flags, "stats")
	}
	if f.showMerkle {
		flags = append(flags, "merkle")
	}
	if f.hashOnly {
		flags = append(flags, "hash-only")
	}
	if f.transformYAML {
		flags = append(flags, "transform")
	}
	if f.typegenTarget != "" {
		flags = append(flags, "typegen")
	}
	if f.jqExpr != "" {
		flags = append(flags, "query")
	}
	if f.compact {
		flags = append(flags, "compact")
	}
	if f.validateOnly {
		flags = append(flags, "validate-only")
	}
	return flags
}
