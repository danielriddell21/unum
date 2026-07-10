package static

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"

	"github.com/danielriddell21/unum/internal/theme"
)

type Theme struct {
	Banner lipgloss.Style
}

func themeFromPalette(p theme.Palette) Theme {
	return Theme{
		Banner: lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary)).Bold(true),
	}
}

func ResolveTheme(name string) Theme {
	return themeFromPalette(theme.ResolvePalette(name))
}

type Options struct {
	Theme   Theme
	NoColor bool
	Quiet   bool
}

func Boot(w io.Writer, lang, format string, opts Options) {
	if opts.Quiet {
		return
	}
	if opts.NoColor {
		_, _ = fmt.Fprintf(w, "[ UNUM ] render  %s → %s  · ✓\n", lang, format)
		return
	}
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A"))
	_, _ = fmt.Fprintf(w, "%s %s  %s\n",
		opts.Theme.Banner.Render("[ UNUM ]"),
		opts.Theme.Banner.Render("render"),
		muted.Render(fmt.Sprintf("%s → %s  · ✓", lang, format)),
	)
}

func Write(w io.Writer, data []byte) error {
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}
