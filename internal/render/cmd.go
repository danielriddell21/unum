package rendertool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	ometric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/danielriddell21/unum/internal/config"
	"github.com/danielriddell21/unum/internal/render/diagram"
	"github.com/danielriddell21/unum/internal/render/render/static"
	rendertui "github.com/danielriddell21/unum/internal/render/render/tui"
	renderweb "github.com/danielriddell21/unum/internal/render/render/web"
	"github.com/danielriddell21/unum/internal/telemetry"
)

type flags struct {
	ui         bool
	web        bool
	webPort    int
	lang       string
	format     string
	output     string
	theme      string
	lightTheme string
	quiet      bool
	noColor    bool
	version    string
	tel        *telemetry.Telemetry
}

func Command(globalNoColor *bool, globalQuiet *bool, version string, tel *telemetry.Telemetry) *cobra.Command {
	f := &flags{}
	f.version = version
	f.tel = tel
	cfg := config.Load()
	f.theme = cfg.DarkTheme
	f.lightTheme = cfg.LightTheme

	cmd := &cobra.Command{
		Use:   "render <file>",
		Short: "Render mermaid and d2 diagrams to images and draw.io",
		Long: `Render a diagram source file to an image or an editable draw.io document.

Language is auto-detected from the file extension:
  .mmd / .mermaid → mermaid
  .d2             → d2
Override with --lang.

Output formats (--format):
  svg      Scalable vector image (default)
  png      Rasterised image
  drawio   diagrams.net document — an editable node graph for d2 and mermaid
           flowcharts, an embedded image for other mermaid diagram types

Output modes:
  (default)  Write the rendered diagram to stdout or --output
  --ui       Terminal viewer with source and render details
  --web      Browser-based live preview

Rendering is pure Go and needs no browser. d2 draws via the terrastruct
library; mermaid draws via go-mermaid, which supports flowchart, sequence,
class, state, er, pie, journey, quadrant, gitgraph, timeline, mindmap,
gantt, c4, requirement, sankey, xychart, block, kanban, packet and radar.`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f.noColor = f.noColor || *globalNoColor
			f.quiet = f.quiet || *globalQuiet
			return runRender(f, args[0])
		},
	}

	cmd.Flags().BoolVar(&f.ui, "ui", false, "launch terminal viewer")
	cmd.Flags().BoolVar(&f.web, "web", false, "launch web UI in browser")
	cmd.Flags().IntVar(&f.webPort, "web-port", 0, "port for --web (default: random free port)")
	cmd.Flags().StringVar(&f.lang, "lang", "", "force language: mermaid, d2 (default: auto)")
	cmd.Flags().StringVar(&f.theme, "theme", cfg.DarkTheme, "color theme: cyber, matrix, dracula, nord, clean, solarized")
	cmd.Flags().StringVar(&f.format, "format", "svg", "output format: svg, png, drawio")
	cmd.Flags().StringVarP(&f.output, "output", "o", "", "write to a file (default: stdout)")
	cmd.Flags().BoolVar(&f.quiet, "quiet", false, "suppress the boot line")

	return cmd
}

func runRender(f *flags, file string) error {
	lang := ParseLanguage(f.lang, file)
	if lang == LangUnknown {
		return fmt.Errorf("cannot determine diagram language for %s (use --lang mermaid|d2)", file)
	}
	format, ok := ParseFormat(f.format)
	if !ok {
		return fmt.Errorf("unknown format %q (want svg, png, or drawio)", f.format)
	}

	mode := "cli"
	if f.ui {
		mode = "tui"
	} else if f.web {
		mode = "web"
	}

	ctx, span := f.tel.Tracer().Start(context.Background(), "render.execute",
		trace.WithAttributes(
			attribute.String("tool", "render"),
			attribute.String("mode", mode),
			attribute.String("lang", lang.String()),
			attribute.String("format", format.String()),
			attribute.String("os", runtime.GOOS),
			attribute.String("arch", runtime.GOARCH),
		),
	)
	defer span.End()

	data, err := os.ReadFile(file)
	if err != nil {
		recordError(ctx, f, span, "read")
		return fmt.Errorf("cannot read %s: %w", file, err)
	}
	f.tel.M.Invocations.Add(ctx, 1, ometric.WithAttributes(
		attribute.String("tool", "render"),
		attribute.String("mode", mode),
		attribute.String("os", runtime.GOOS),
		attribute.String("arch", runtime.GOARCH),
		attribute.String("version", f.version),
	))
	f.tel.M.InputBytes.Record(ctx, int64(len(data)), ometric.WithAttributes(attribute.String("tool", "render")))

	if f.web {
		return runWeb(f, lang, data)
	}
	if f.ui {
		return runTUI(f, file, lang, data)
	}
	return runStatic(ctx, f, span, lang, format, data)
}

func runStatic(ctx context.Context, f *flags, span trace.Span, lang Language, format Format, data []byte) error {
	opts := static.Options{Theme: static.ResolveTheme(f.theme), NoColor: f.noColor, Quiet: f.quiet}
	static.Boot(os.Stderr, lang.String(), format.String(), opts)

	start := time.Now()
	out, err := render(f, lang, format, data)
	if err != nil {
		recordError(ctx, f, span, "render")
		return err
	}
	elapsed := time.Since(start)

	if err := writeOutput(f, format, out); err != nil {
		recordError(ctx, f, span, "write")
		return err
	}

	span.SetAttributes(attribute.Int("render.output_bytes", len(out)))
	f.tel.M.Duration.Record(ctx, elapsed.Seconds(), ometric.WithAttributes(
		attribute.String("tool", "render"),
		attribute.String("mode", "cli"),
	))
	return nil
}

func runWeb(f *flags, lang Language, data []byte) error {
	if err := renderweb.Start(renderweb.Options{
		Port:       f.webPort,
		Quiet:      f.quiet,
		DarkTheme:  f.theme,
		LightTheme: f.lightTheme,
		Version:    f.version,
		Tel:        f.tel,
		Source:     string(data),
		Lang:       lang.String(),
	}); err != nil {
		return fmt.Errorf("render web: %w", err)
	}
	return nil
}

func render(f *flags, lang Language, format Format, data []byte) ([]byte, error) {
	out, _, err := diagram.Render(lang.String(), format.String(), string(data), diagram.ThemeByName(f.theme))
	if err != nil {
		return nil, fmt.Errorf("render %s: %w", lang, err)
	}
	return out, nil
}

func runTUI(f *flags, file string, lang Language, data []byte) error {
	info := rendertui.Info{
		File:   filepath.Base(file),
		Lang:   lang.String(),
		Source: string(data),
		Shapes: -1,
	}
	if lang == LangD2 {
		return runTUID2(f, info, data)
	}
	return runTUIMermaid(f, info, data)
}

func runTUID2(f *flags, info rendertui.Info, data []byte) error {
	d, err := diagram.RenderD2(string(data), diagram.ThemeByName(f.theme))
	if err != nil {
		info.Err = err
		return startTUI(f, info)
	}
	info.SVG = d.SVG
	info.Width, info.Height = diagram.SVGSize(d.SVG)
	info.Shapes, info.Conns = d.NumShapes(), d.NumConnections()
	info.Drawio, _ = d.Drawio()
	if ascii, err := diagram.RenderD2ASCII(string(data)); err == nil {
		info.ASCII = ascii
	}
	info.PNG, _ = diagram.SVGToPNG(d.SVG)
	return startTUI(f, info)
}

func runTUIMermaid(f *flags, info rendertui.Info, data []byte) error {
	svg, err := diagram.RenderMermaid(string(data), diagram.ThemeByName(f.theme))
	if err != nil {
		info.Err = err
		return startTUI(f, info)
	}
	info.SVG = svg
	info.Width, info.Height = diagram.SVGSize(svg)
	info.Drawio, _ = diagram.MermaidDrawio(string(data), svg, diagram.ThemeByName(f.theme))
	info.PNG, _ = diagram.MermaidSVGToPNG(svg)

	// Graph-shaped types convert to d2 for a crisp Unicode preview; chart and
	// timeline types yield "", leaving the TUI to show its no-preview note.
	if d2src, err := diagram.MermaidToD2(string(data)); err == nil && d2src != "" {
		if ascii, err := diagram.RenderD2ASCII(d2src); err == nil {
			info.ASCII = ascii
		}
	}
	return startTUI(f, info)
}

func startTUI(f *flags, info rendertui.Info) error {
	if err := rendertui.Start(info, f.theme, f.version); err != nil {
		return fmt.Errorf("render TUI: %w", err)
	}
	return nil
}

func writeOutput(f *flags, format Format, out []byte) error {
	if f.output != "" {
		if err := os.WriteFile(f.output, out, 0o644); err != nil { //nolint:gosec // user-facing artifact, not a secret
			return fmt.Errorf("write %s: %w", f.output, err)
		}
		return nil
	}
	if format == FormatPNG && isTerminal(os.Stdout) {
		return fmt.Errorf("refusing to write PNG to the terminal: use --output <file> or redirect stdout")
	}
	if err := static.Write(os.Stdout, out); err != nil {
		return fmt.Errorf("write stdout: %w", err)
	}
	return nil
}

func isTerminal(file *os.File) bool {
	fi, err := file.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func recordError(ctx context.Context, f *flags, span trace.Span, kind string) {
	span.SetStatus(codes.Error, kind)
	f.tel.M.Errors.Add(ctx, 1, ometric.WithAttributes(
		attribute.String("tool", "render"),
		attribute.String("error_type", kind),
	))
}
