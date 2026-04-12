// Package hash implements the `unum hash` subcommand — a deterministic deriver
// that maps any string to a stable set of useful values.
package hash

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/danielriddell21/unum/internal/config"
	"github.com/danielriddell21/unum/internal/hash/derive"
	"github.com/danielriddell21/unum/internal/hash/history"
	"github.com/danielriddell21/unum/internal/hash/render/static"
	hashTUI "github.com/danielriddell21/unum/internal/hash/render/tui"
	hashWeb "github.com/danielriddell21/unum/internal/hash/render/web"
)

type flags struct {
	ui         bool
	web        bool
	theme      string
	lightTheme string
	webPort    int
	quiet      bool
	noColor    bool

	// Single-field output flags
	portOnly   bool
	uuidOnly   bool
	colorOnly  bool
	shortOnly  bool
	emojiOnly  bool
	phraseOnly bool
}

// Command returns the cobra command for `unum hash`.
func Command(globalNoColor *bool, globalQuiet *bool) *cobra.Command {
	f := &flags{}
	cfg := config.Load()
	f.theme = cfg.DarkTheme
	f.lightTheme = cfg.LightTheme

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
	cmd.Flags().BoolVar(&f.portOnly, "port", false, "print derived port only")
	cmd.Flags().BoolVar(&f.uuidOnly, "uuid", false, "print UUID only")
	cmd.Flags().BoolVar(&f.colorOnly, "color", false, "print hex color only")
	cmd.Flags().BoolVar(&f.shortOnly, "short", false, "print short ID only")
	cmd.Flags().BoolVar(&f.emojiOnly, "emoji", false, "print emoji only")
	cmd.Flags().BoolVar(&f.phraseOnly, "phrase", false, "print passphrase only")

	return cmd
}

func runHash(f *flags, args []string) error {
	if f.ui {
		static.Boot(os.Stderr, static.Options{Theme: static.ResolveTheme(f.theme), Quiet: f.quiet})
		applyTUITheme(f.theme)
		hashTUI.SetFuncs(derive.Derive, history.Append, history.Load)
		p := tea.NewProgram(hashTUI.NewModel(), tea.WithAltScreen())
		_, err := p.Run()
		return err
	}

	if f.web {
		static.Boot(os.Stderr, static.Options{Theme: static.ResolveTheme(f.theme), Quiet: f.quiet})
		hashWeb.SetFuncs(derive.Derive)
		return hashWeb.Start(hashWeb.Options{Port: f.webPort, Quiet: f.quiet, DarkTheme: f.theme, LightTheme: f.lightTheme})
	}

	if len(args) == 0 {
		return fmt.Errorf("text argument required (or use --ui / --web)")
	}
	input := args[0]

	r := derive.Derive(input)
	_ = history.Append(input)

	switch {
	case f.portOnly:
		static.RenderSingle(os.Stdout, fmt.Sprintf("%d", r.Port))
	case f.uuidOnly:
		static.RenderSingle(os.Stdout, r.UUID)
	case f.colorOnly:
		static.RenderSingle(os.Stdout, r.Color)
	case f.shortOnly:
		static.RenderSingle(os.Stdout, r.Short)
	case f.emojiOnly:
		static.RenderSingle(os.Stdout, r.Emoji)
	case f.phraseOnly:
		static.RenderSingle(os.Stdout, r.Phrase)
	default:
		opts := static.Options{
			Theme:   static.ResolveTheme(f.theme),
			NoColor: f.noColor,
			Quiet:   f.quiet,
		}
		static.Boot(os.Stdout, opts)
		static.RenderTable(os.Stdout, r, opts)
	}

	return nil
}

// applyTUITheme maps a theme name to TUI styles.
func applyTUITheme(name string) {
	type palette struct{ accent, dim, value, border string }
	themes := map[string]palette{
		"matrix":  {"#00FF41", "#1A3A1A", "#CCFFCC", "#0A1A0A"},
		"dracula": {"#BD93F9", "#44475A", "#F8F8F2", "#282A36"},
		"nord":    {"#88C0D0", "#4C566A", "#ECEFF4", "#2E3440"},
	}
	p, ok := themes[name]
	if !ok {
		p = palette{"#00D4FF", "#3A3A3A", "#E5E5E5", "#1A1A2E"}
	}
	hashTUI.ApplyTheme(p.accent, p.dim, p.value, p.border)
}
