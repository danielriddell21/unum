// Package hash implements the `unum hash` subcommand — a deterministic deriver
// that maps any string to a stable set of useful values.
package hash

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/unum/internal/config"
	"github.com/danielriddell21/unum/internal/hash/render/static"
)

type flags struct {
	theme   string
	quiet   bool
	noColor bool

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

	cmd := &cobra.Command{
		Use:   "hash [text]",
		Short: "Deterministically derive useful values from any string",
		Long: `Hash any string into a stable set of derived values using SHA256.

The same input always produces the same output — useful for generating
consistent ports, UUIDs, colors, and identifiers for named services,
environments, or any repeatable string.

Single-field flags (pipe-friendly, skips the table):
  --port, --uuid, --color, --short, --emoji, --phrase`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f.noColor = f.noColor || *globalNoColor
			f.quiet = f.quiet || *globalQuiet
			return runHash(f, args[0])
		},
	}

	cmd.Flags().StringVar(&f.theme, "theme", cfg.Theme, "color theme: cyber, matrix, dracula, nord")
	cmd.Flags().BoolVar(&f.quiet, "quiet", false, "suppress the boot line")
	cmd.Flags().BoolVar(&f.portOnly, "port", false, "print derived port only")
	cmd.Flags().BoolVar(&f.uuidOnly, "uuid", false, "print UUID only")
	cmd.Flags().BoolVar(&f.colorOnly, "color", false, "print hex color only")
	cmd.Flags().BoolVar(&f.shortOnly, "short", false, "print short ID only")
	cmd.Flags().BoolVar(&f.emojiOnly, "emoji", false, "print emoji only")
	cmd.Flags().BoolVar(&f.phraseOnly, "phrase", false, "print passphrase only")

	return cmd
}

func runHash(f *flags, input string) error {
	r := Derive(input)
	_ = AppendHistory(input)

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
