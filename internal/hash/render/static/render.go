package static

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"

	"github.com/danielriddell21/unum/internal/hash/types"
	"github.com/danielriddell21/unum/internal/theme"
)

type Theme struct {
	Banner lipgloss.Style
	Label  lipgloss.Style
	Value  lipgloss.Style
	Dim    lipgloss.Style
}

func themeFromPalette(p theme.Palette) Theme {
	return Theme{
		Banner: lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary)).Bold(true),
		Label:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary)),
		Value:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.Text)),
		Dim:    lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)),
	}
}

var (
	Cyber   = themeFromPalette(theme.PaletteCyber)
	Matrix  = themeFromPalette(theme.PaletteMatrix)
	Dracula = themeFromPalette(theme.PaletteDracula)
	Nord    = themeFromPalette(theme.PaletteNord)
)

func ResolveTheme(name string) Theme {
	return themeFromPalette(theme.ResolvePalette(name))
}

type Options struct {
	Theme   Theme
	NoColor bool
	Quiet   bool
}

func Boot(w io.Writer, opts Options) {
	if opts.Quiet {
		return
	}
	if opts.NoColor {
		_, _ = fmt.Fprintln(w, "[ UNUM ] hash  · ✓")
		return
	}
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A"))
	_, _ = fmt.Fprintf(w, "%s %s  %s\n",
		opts.Theme.Banner.Render("[ UNUM ]"),
		opts.Theme.Banner.Render("hash"),
		muted.Render("· ✓"),
	)
}

func RenderTable(w io.Writer, r types.Result, opts Options) {
	rows := []struct{ label, value string }{
		{"input", r.Input},
		{"port", fmt.Sprintf("%d", r.Port)},
		{"uuid", r.UUID},
		{"color", r.Color},
		{"short", r.Short},
		{"emoji", r.Emoji},
		{"phrase", r.Phrase},
	}

	sep := opts.Theme.Dim.Render("────────────────────────────────────")

	if !opts.NoColor {
		_, _ = fmt.Fprintln(w, sep)
	}

	for _, row := range rows {
		label := fmt.Sprintf("  %-7s", row.label)
		if opts.NoColor {
			_, _ = fmt.Fprintf(w, "%s  %s\n", label, row.value)
		} else {
			_, _ = fmt.Fprintf(w, "%s  %s\n",
				opts.Theme.Label.Render(label),
				opts.Theme.Value.Render(row.value),
			)
		}
	}

	if !opts.NoColor {
		_, _ = fmt.Fprintln(w, sep)
	}
}

func RenderSingle(w io.Writer, value string) {
	_, _ = fmt.Fprintln(w, value)
}
