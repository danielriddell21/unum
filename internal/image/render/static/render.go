package static

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/danielriddell21/unum/internal/image/optimize"
	"github.com/danielriddell21/unum/internal/theme"
)

type Theme struct {
	Banner lipgloss.Style
	Label  lipgloss.Style
	Value  lipgloss.Style
	Dim    lipgloss.Style
	Win    lipgloss.Style
}

func themeFromPalette(p theme.Palette) Theme {
	return Theme{
		Banner: lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary)).Bold(true),
		Label:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary)),
		Value:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.Text)),
		Dim:    lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)),
		Win:    lipgloss.NewStyle().Foreground(lipgloss.Color(p.Added)),
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

func Boot(w io.Writer, name string, opts Options) {
	if opts.Quiet {
		return
	}
	if opts.NoColor {
		_, _ = fmt.Fprintf(w, "[ UNUM ] image  %s  · ✓\n", name)
		return
	}
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A"))
	_, _ = fmt.Fprintf(w, "%s %s  %s\n",
		opts.Theme.Banner.Render("[ UNUM ]"),
		opts.Theme.Banner.Render("image"),
		muted.Render(fmt.Sprintf("%s  · ✓", name)),
	)
}

type Report struct {
	Source optimize.Source
	Format optimize.Format
	Width  int
	Height int
	Steps  []optimize.Step
	Active int
}

func RenderReport(w io.Writer, rep Report, opts Options) {
	rows := []struct{ label, value string }{
		{"source", rep.Source.Name},
		{"format", rep.Source.Format.String()},
		{"dimensions", fmt.Sprintf("%d × %d", rep.Source.Width, rep.Source.Height)},
		{"size", optimize.HumanBytes(rep.Source.Bytes)},
		{"output", fmt.Sprintf("%s  %d × %d", rep.Format, rep.Width, rep.Height)},
	}

	sep := strings.Repeat("─", 46)
	writeLine(w, opts, opts.Theme.Dim, sep)
	for _, row := range rows {
		label := fmt.Sprintf("  %-11s", row.label)
		if opts.NoColor {
			_, _ = fmt.Fprintf(w, "%s  %s\n", label, row.value)
			continue
		}
		_, _ = fmt.Fprintf(w, "%s  %s\n", opts.Theme.Label.Render(label), opts.Theme.Value.Render(row.value))
	}

	if len(rep.Steps) == 0 {
		writeLine(w, opts, opts.Theme.Dim, sep)
		return
	}

	_, _ = fmt.Fprintln(w)
	header := fmt.Sprintf("  %-9s %-11s %s", "quality", "size", "saving")
	writeLine(w, opts, opts.Theme.Dim, header)
	writeLine(w, opts, opts.Theme.Dim, sep)

	for _, step := range rep.Steps {
		renderStep(w, rep, step, opts)
	}
	writeLine(w, opts, opts.Theme.Dim, sep)
}

func renderStep(w io.Writer, rep Report, step optimize.Step, opts Options) {
	saving := optimize.Saving(rep.Source.Bytes, step.Bytes)
	marker := ""
	if step.Quality == rep.Active {
		marker = "  ← default"
	}

	quality := fmt.Sprintf("  %-9d", step.Quality)
	size := fmt.Sprintf("%-11s", optimize.HumanBytes(step.Bytes))
	pct := fmt.Sprintf("%+.0f%%", -saving*100)

	if opts.NoColor {
		_, _ = fmt.Fprintf(w, "%s %s %s%s\n", quality, size, pct, marker)
		return
	}

	pctStyle := opts.Theme.Value
	if saving > 0 {
		pctStyle = opts.Theme.Win
	}
	_, _ = fmt.Fprintf(w, "%s %s %s%s\n",
		opts.Theme.Label.Render(quality),
		opts.Theme.Value.Render(size),
		pctStyle.Render(pct),
		opts.Theme.Dim.Render(marker),
	)
}

type Summary struct {
	Source optimize.Source
	Result optimize.Result
	Output string
}

func RenderSummary(w io.Writer, sum Summary, opts Options) {
	saving := optimize.Saving(sum.Source.Bytes, sum.Result.Bytes)
	detail := fmt.Sprintf("(%s q%d, %d × %d)",
		sum.Result.Format, sum.Result.Quality, sum.Result.Width, sum.Result.Height)
	pct := fmt.Sprintf("%+.0f%%", -saving*100)

	if opts.NoColor {
		_, _ = fmt.Fprintf(w, "  %s %s  →  %s %s   %s   %s\n",
			sum.Source.Name, optimize.HumanBytes(sum.Source.Bytes),
			sum.Output, optimize.HumanBytes(sum.Result.Bytes), pct, detail)
		return
	}

	pctStyle := opts.Theme.Value
	if saving > 0 {
		pctStyle = opts.Theme.Win
	}
	_, _ = fmt.Fprintf(w, "  %s %s  →  %s %s   %s   %s\n",
		opts.Theme.Value.Render(sum.Source.Name),
		opts.Theme.Dim.Render(optimize.HumanBytes(sum.Source.Bytes)),
		opts.Theme.Label.Render(sum.Output),
		opts.Theme.Value.Render(optimize.HumanBytes(sum.Result.Bytes)),
		pctStyle.Render(pct),
		opts.Theme.Dim.Render(detail),
	)
}

func RenderNote(w io.Writer, note string, opts Options) {
	if opts.NoColor {
		_, _ = fmt.Fprintf(w, "  %s\n", note)
		return
	}
	_, _ = fmt.Fprintf(w, "  %s\n", opts.Theme.Dim.Render(note))
}

func Write(w io.Writer, data []byte) error {
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func writeLine(w io.Writer, opts Options, style lipgloss.Style, s string) {
	if opts.NoColor {
		_, _ = fmt.Fprintln(w, s)
		return
	}
	_, _ = fmt.Fprintln(w, style.Render(s))
}
