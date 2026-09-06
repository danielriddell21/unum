package hash

import (
	"context"
	"fmt"
	"os"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
	ometric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/danielriddell21/unum/internal/config"
	"github.com/danielriddell21/unum/internal/hash/derive"
	"github.com/danielriddell21/unum/internal/hash/history"
	"github.com/danielriddell21/unum/internal/hash/render/static"
	hashTUI "github.com/danielriddell21/unum/internal/hash/render/tui"
	hashWeb "github.com/danielriddell21/unum/internal/hash/render/web"
	"github.com/danielriddell21/unum/internal/telemetry"
	"github.com/danielriddell21/unum/internal/tui/panels"
)

type flags struct {
	ui         bool
	web        bool
	theme      string
	lightTheme string
	webPort    int
	quiet      bool
	noColor    bool

	version string
	tel     *telemetry.Telemetry

	noHistory    bool
	clearHistory bool
	historyOn    bool

	portOnly   bool
	uuidOnly   bool
	colorOnly  bool
	shortOnly  bool
	emojiOnly  bool
	phraseOnly bool
}

func Command(globalNoColor *bool, globalQuiet *bool, version string, tel *telemetry.Telemetry) *cobra.Command {
	f := &flags{}
	f.version = version
	f.tel = tel
	cfg := config.Load()
	f.theme = cfg.DarkTheme
	f.lightTheme = cfg.LightTheme
	f.historyOn = cfg.HashHistoryEnabled()

	cmd := &cobra.Command{
		Use:   "hash [text]",
		Short: "Deterministically derive useful values from any string",
		Long: `Hash any string into a stable set of derived values using SHA256.

The same input always produces the same output — useful for generating
consistent ports, UUIDs, colors, and identifiers for named services,
environments, or any repeatable string.

Output modes:
  (default)  Full derivation table
  --ui       Interactive TUI with history
  --web      Browser-based UI with history

Single-field flags (pipe-friendly, skips the table):
  --port, --uuid, --color, --short, --emoji, --phrase`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f.noColor = f.noColor || *globalNoColor
			f.quiet = f.quiet || *globalQuiet
			return runHash(f, args)
		},
	}

	cmd.Flags().BoolVar(&f.ui, "ui", false, "launch interactive TUI with history")
	cmd.Flags().BoolVar(&f.web, "web", false, "launch web UI in browser")
	cmd.Flags().IntVar(&f.webPort, "web-port", 0, "port for --web (default: random free port)")
	cmd.Flags().BoolVar(&f.quiet, "quiet", false, "suppress the boot line")
	cmd.Flags().BoolVar(&f.noHistory, "no-history", false, "do not record this input in the local history file")
	cmd.Flags().BoolVar(&f.clearHistory, "clear-history", false, "delete the local history file and exit")
	cmd.Flags().BoolVar(&f.portOnly, "port", false, "print derived port only")
	cmd.Flags().BoolVar(&f.uuidOnly, "uuid", false, "print UUID only")
	cmd.Flags().BoolVar(&f.colorOnly, "color", false, "print hex color only")
	cmd.Flags().BoolVar(&f.shortOnly, "short", false, "print short ID only")
	cmd.Flags().BoolVar(&f.emojiOnly, "emoji", false, "print emoji only")
	cmd.Flags().BoolVar(&f.phraseOnly, "phrase", false, "print passphrase only")

	return cmd
}

func runHash(f *flags, args []string) error {
	if f.clearHistory {
		if err := history.Clear(); err != nil {
			return fmt.Errorf("clear history: %w", err)
		}
		fmt.Fprintln(os.Stdout, "hash history cleared")
		return nil
	}

	mode := "cli"
	if f.ui {
		mode = "tui"
	} else if f.web {
		mode = "web"
	}
	ctx, span := f.tel.Tracer().Start(context.Background(), "hash.execute",
		trace.WithAttributes(
			attribute.String("tool", "hash"),
			attribute.String("mode", mode),
			attribute.String("os", runtime.GOOS),
			attribute.String("arch", runtime.GOARCH),
			attribute.StringSlice("flags", activeHashFlags(f)),
		),
	)
	defer span.End()

	if f.ui {
		f.tel.M.Invocations.Add(ctx, 1, ometric.WithAttributes(
			attribute.String("tool", "hash"),
			attribute.String("mode", "tui"),
			attribute.String("os", runtime.GOOS),
			attribute.String("arch", runtime.GOARCH),
			attribute.String("version", f.version),
		))
		static.Boot(os.Stderr, static.Options{Theme: static.ResolveTheme(f.theme), Quiet: f.quiet})
		hashTUI.ApplyPalette(panels.ResolvePalette(f.theme))
		deps := hashTUI.Deps{Derive: derive.Derive, AppendHistory: f.recordHistory, LoadHistory: history.Load}
		p := tea.NewProgram(hashTUI.NewModel(f.version, deps), tea.WithAltScreen())
		_, err := p.Run()
		if err != nil {
			return fmt.Errorf("hash TUI: %w", err)
		}
		return nil
	}

	if f.web {
		f.tel.M.Invocations.Add(ctx, 1, ometric.WithAttributes(
			attribute.String("tool", "hash"),
			attribute.String("mode", "web"),
			attribute.String("os", runtime.GOOS),
			attribute.String("arch", runtime.GOARCH),
			attribute.String("version", f.version),
		))
		static.Boot(os.Stderr, static.Options{Theme: static.ResolveTheme(f.theme), Quiet: f.quiet})
		if err := hashWeb.Start(hashWeb.Options{Port: f.webPort, Quiet: f.quiet, DarkTheme: f.theme, LightTheme: f.lightTheme, Version: f.version, Tel: f.tel, Derive: derive.Derive}); err != nil {
			return fmt.Errorf("hash web: %w", err)
		}
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("text argument required (or use --ui / --web)")
	}
	input := args[0]

	r := derive.Derive(input)
	if err := f.recordHistory(input); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not write hash history: %v\n", err)
	}

	outputField := "full"
	switch {
	case f.portOnly:
		outputField = "port"
		static.RenderSingle(os.Stdout, fmt.Sprintf("%d", r.Port))
	case f.uuidOnly:
		outputField = "uuid"
		static.RenderSingle(os.Stdout, r.UUID)
	case f.colorOnly:
		outputField = "color"
		static.RenderSingle(os.Stdout, r.Color)
	case f.shortOnly:
		outputField = "short"
		static.RenderSingle(os.Stdout, r.Short)
	case f.emojiOnly:
		outputField = "emoji"
		static.RenderSingle(os.Stdout, r.Emoji)
	case f.phraseOnly:
		outputField = "phrase"
		static.RenderSingle(os.Stdout, r.Phrase)
	default:
		opts := static.Options{
			Theme:   static.ResolveTheme(f.theme),
			NoColor: f.noColor,
			Quiet:   f.quiet,
		}
		static.Boot(os.Stderr, opts)
		static.RenderTable(os.Stdout, r, opts)
	}

	span.SetAttributes(attribute.String("hash.output_field", outputField))
	f.tel.M.Invocations.Add(ctx, 1, ometric.WithAttributes(
		attribute.String("tool", "hash"),
		attribute.String("mode", "cli"),
		attribute.String("os", runtime.GOOS),
		attribute.String("arch", runtime.GOARCH),
		attribute.String("version", f.version),
	))

	return nil
}

func (f *flags) recordHistory(input string) error {
	if f.noHistory || !f.historyOn {
		return nil
	}
	return history.Append(input)
}

func activeHashFlags(f *flags) []string {
	var flags []string
	if f.portOnly {
		flags = append(flags, "port")
	}
	if f.uuidOnly {
		flags = append(flags, "uuid")
	}
	if f.colorOnly {
		flags = append(flags, "color")
	}
	if f.shortOnly {
		flags = append(flags, "short")
	}
	if f.emojiOnly {
		flags = append(flags, "emoji")
	}
	if f.phraseOnly {
		flags = append(flags, "phrase")
	}
	if f.noHistory {
		flags = append(flags, "no-history")
	}
	return flags
}
