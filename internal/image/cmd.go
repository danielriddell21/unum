package imagetool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	ometric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/danielriddell21/unum/internal/config"
	"github.com/danielriddell21/unum/internal/image/optimize"
	"github.com/danielriddell21/unum/internal/image/render/static"
	"github.com/danielriddell21/unum/internal/telemetry"
)

type flags struct {
	ui         bool
	web        bool
	webPort    int
	output     string
	quality    int
	scale      string
	maxWidth   int
	maxHeight  int
	to         string
	target     string
	theme      string
	lightTheme string
	quiet      bool
	noColor    bool

	version string
	tel     *telemetry.Telemetry
}

func Command(globalNoColor *bool, globalQuiet *bool, version string, tel *telemetry.Telemetry) *cobra.Command {
	f := &flags{}
	f.version = version
	f.tel = tel
	cfg := config.Load()
	f.theme = cfg.DarkTheme
	f.lightTheme = cfg.LightTheme

	cmd := &cobra.Command{
		Use:   "image [file]",
		Short: "Shrink images — trade quality and scale against file size",
		Long: `Optimize an image: make the file smaller while keeping it looking right.

Two levers do the work. Quality controls how much detail the encoder keeps,
and scale controls how many pixels there are to encode in the first place.

Run without --output to see the tradeoff rather than guess at it: unum prints
the source details and a ladder of what each quality level would cost.

Output modes:
  (default)  Analysis report and quality/size ladder
  --output   Write the optimized image (use -o - for stdout)
  --ui       Interactive terminal viewer with live preview
  --web      Browser UI with before/after preview

Quality means different things per format:
  jpeg   The encoder's own quality knob, 1-100.
  png    The palette size. 100 stays lossless truecolor; below that the image
         is reduced to roughly that percentage of 256 colors. A file is never
         written larger than its lossless version.
  gif    Always palettized; quality sets the number of colors.
  webp   Lossless, and usually the smallest of the four. Quality drives the
         same palette reduction as png.

Reads jpeg, png, gif, webp, tiff and bmp. Writes jpeg, png, gif and webp.
WebP output is lossless only — for photographs, jpeg is still smaller.
EXIF orientation is applied to the pixels so rotated photos stay upright;
all other metadata is dropped, which is usually wanted and shrinks the file.

Examples:
  unum image photo.jpg
  unum image photo.jpg -o small.jpg -q 70
  unum image photo.jpg -o thumb.jpg --max-width 400
  unum image photo.jpg -o web.jpg --target 200kb
  unum image shot.png -o shot.webp --to webp
  unum image shot.png -o shot.jpg --to jpeg --scale 50%`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f.noColor = f.noColor || *globalNoColor
			f.quiet = f.quiet || *globalQuiet
			if len(args) == 0 {
				return runNoFile(f)
			}
			return runImage(f, args[0])
		},
	}

	cmd.Flags().BoolVar(&f.ui, "ui", false, "launch interactive terminal viewer")
	cmd.Flags().BoolVar(&f.web, "web", false, "launch web UI in browser")
	cmd.Flags().IntVar(&f.webPort, "web-port", 0, "port for --web (default: random free port)")
	cmd.Flags().StringVarP(&f.output, "output", "o", "", "write the optimized image to a file (- for stdout)")
	cmd.Flags().IntVarP(&f.quality, "quality", "q", 80, "encoder quality, 1-100")
	cmd.Flags().StringVar(&f.scale, "scale", "", "resize by a factor or percentage, e.g. 50% or 0.5")
	cmd.Flags().IntVar(&f.maxWidth, "max-width", 0, "shrink so the width is at most this many pixels")
	cmd.Flags().IntVar(&f.maxHeight, "max-height", 0, "shrink so the height is at most this many pixels")
	cmd.Flags().StringVar(&f.to, "to", "", "output format: jpeg, png, gif, webp (default: match the source)")
	cmd.Flags().StringVar(&f.target, "target", "", "target file size, e.g. 200kb — picks the best quality that fits")
	cmd.Flags().StringVar(&f.theme, "theme", cfg.DarkTheme, "color theme: cyber, matrix, dracula, nord, clean, solarized")
	cmd.Flags().BoolVar(&f.quiet, "quiet", false, "suppress the boot line")

	return cmd
}

func runNoFile(f *flags) error {
	if !f.web {
		return fmt.Errorf("no image file given (use --web to load one in the browser)")
	}
	return runWeb(f, "", nil, optimize.FormatUnknown)
}

type settings struct {
	opts   optimize.Options
	budget int
}

func (f *flags) settings(src optimize.Source) (settings, error) {
	scale, err := optimize.ParseScale(f.scale)
	if err != nil {
		return settings{}, fmt.Errorf("--scale: %w", err)
	}

	format := optimize.FormatUnknown
	if f.to != "" {
		parsed, ok := optimize.ParseFormat(f.to)
		if !ok {
			return settings{}, fmt.Errorf("unknown format %q (want jpeg, png, gif, or webp)", f.to)
		}
		format = parsed
	} else if inferred, ok := formatFromPath(f.output); ok {
		format = inferred
	}

	resolved := optimize.ResolveOutputFormat(format, src.Format)
	if !resolved.CanEncode() {
		return settings{}, fmt.Errorf("cannot write %s: use --to jpeg, png, gif, or webp", resolved)
	}

	budget := 0
	if f.target != "" {
		budget, err = optimize.ParseBytes(f.target)
		if err != nil {
			return settings{}, fmt.Errorf("--target: %w", err)
		}
	}

	return settings{
		opts: optimize.Options{
			Quality:   f.quality,
			Scale:     scale,
			MaxWidth:  f.maxWidth,
			MaxHeight: f.maxHeight,
			Format:    resolved,
		},
		budget: budget,
	}, nil
}

func formatFromPath(path string) (optimize.Format, bool) {
	if path == "" || path == "-" {
		return optimize.FormatUnknown, false
	}
	ext := filepath.Ext(path)
	if ext == "" {
		return optimize.FormatUnknown, false
	}
	return optimize.ParseFormat(ext[1:])
}

func runImage(f *flags, file string) error {
	mode := "cli"
	if f.ui {
		mode = "tui"
	} else if f.web {
		mode = "web"
	}

	ctx, span := f.tel.Tracer().Start(context.Background(), "image.execute",
		trace.WithAttributes(
			attribute.String("tool", "image"),
			attribute.String("mode", mode),
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
		attribute.String("tool", "image"),
		attribute.String("mode", mode),
		attribute.String("os", runtime.GOOS),
		attribute.String("arch", runtime.GOARCH),
		attribute.String("version", f.version),
	))
	f.tel.M.InputBytes.Record(ctx, int64(len(data)), ometric.WithAttributes(attribute.String("tool", "image")))

	src, err := optimize.Decode(filepath.Base(file), data)
	if err != nil {
		recordError(ctx, f, span, "decode")
		return fmt.Errorf("cannot decode %s: %w", file, err)
	}

	set, err := f.settings(src)
	if err != nil {
		recordError(ctx, f, span, "flags")
		return err
	}

	span.SetAttributes(
		attribute.String("image.source_format", src.Format.String()),
		attribute.String("image.output_format", set.opts.Format.String()),
		attribute.Int("image.source_bytes", src.Bytes),
	)

	if f.web {
		return runWeb(f, src.Name, data, src.Format)
	}
	if f.ui {
		return runTUI(f, src, set)
	}
	return runStatic(ctx, f, span, src, set)
}

func runStatic(ctx context.Context, f *flags, span trace.Span, src optimize.Source, set settings) error {
	opts := static.Options{Theme: static.ResolveTheme(f.theme), NoColor: f.noColor, Quiet: f.quiet}
	static.Boot(os.Stderr, src.Name, opts)

	if f.output == "" {
		return renderAnalysis(ctx, f, span, src, set, opts)
	}
	return writeOptimized(ctx, f, span, src, set, opts)
}

func renderAnalysis(ctx context.Context, f *flags, span trace.Span, src optimize.Source, set settings, opts static.Options) error {
	steps, err := optimize.Ladder(src, set.opts, optimize.DefaultLadder)
	if err != nil {
		recordError(ctx, f, span, "encode")
		return fmt.Errorf("build quality ladder: %w", err)
	}

	w, h := optimize.TargetDimensions(src.Width, src.Height, set.opts.Scale, set.opts.MaxWidth, set.opts.MaxHeight)
	static.RenderReport(os.Stdout, static.Report{
		Source: src,
		Format: set.opts.Format,
		Width:  w,
		Height: h,
		Steps:  steps,
		Active: optimize.ClampQuality(set.opts.Quality),
	}, opts)
	static.RenderNote(os.Stdout, "write one out with: -o <file> [-q <quality>]", opts)
	return nil
}

func writeOptimized(ctx context.Context, f *flags, span trace.Span, src optimize.Source, set settings, opts static.Options) error {
	result, err := optimizeOrFit(src, set)
	if err != nil {
		recordError(ctx, f, span, "encode")
		return err
	}

	if err := writeOutput(f, result); err != nil {
		recordError(ctx, f, span, "write")
		return err
	}

	span.SetAttributes(
		attribute.Int("image.output_bytes", result.Bytes),
		attribute.Int("image.quality", result.Quality),
	)

	if f.output == "-" {
		return nil
	}

	static.RenderSummary(os.Stdout, static.Summary{Source: src, Result: result, Output: f.output}, opts)
	if set.budget > 0 && result.Bytes > set.budget {
		static.RenderNote(os.Stdout, fmt.Sprintf(
			"could not reach %s even at quality 1 — try --scale or --max-width",
			optimize.HumanBytes(set.budget)), opts)
	}
	return nil
}

func optimizeOrFit(src optimize.Source, set settings) (optimize.Result, error) {
	if set.budget > 0 {
		result, err := optimize.FitTo(src, set.opts, set.budget)
		if err != nil {
			return optimize.Result{}, fmt.Errorf("fit to %s: %w", optimize.HumanBytes(set.budget), err)
		}
		return result, nil
	}

	result, err := optimize.Optimize(src, set.opts)
	if err != nil {
		return optimize.Result{}, fmt.Errorf("optimize: %w", err)
	}
	return result, nil
}

func writeOutput(f *flags, result optimize.Result) error {
	if f.output == "-" {
		if isTerminal(os.Stdout) {
			return fmt.Errorf("refusing to write %s to the terminal: use -o <file> or redirect stdout", result.Format)
		}
		if err := static.Write(os.Stdout, result.Data); err != nil {
			return fmt.Errorf("write stdout: %w", err)
		}
		return nil
	}
	if err := os.WriteFile(f.output, result.Data, 0o644); err != nil { //nolint:gosec // user-facing artifact, not a secret
		return fmt.Errorf("write %s: %w", f.output, err)
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
		attribute.String("tool", "image"),
		attribute.String("error_type", kind),
	))
}
