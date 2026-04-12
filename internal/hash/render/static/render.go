// Package static renders hash derivation results to a terminal writer.
package static

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/danielriddell21/unum/internal/hash/types"
)

// Theme holds lipgloss styles for terminal output.
type Theme struct {
	Banner lipgloss.Style
	Label  lipgloss.Style
	Value  lipgloss.Style
	Dim    lipgloss.Style
}

var Cyber = Theme{
	Banner: lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4FF")).Bold(true),
	Label:  lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4FF")),
	Value:  lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E5E5")),
	Dim:    lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A")),
}

var Matrix = Theme{
	Banner: lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF41")).Bold(true),
	Label:  lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF41")),
	Value:  lipgloss.NewStyle().Foreground(lipgloss.Color("#CCFFCC")),
	Dim:    lipgloss.NewStyle().Foreground(lipgloss.Color("#1A3A1A")),
}

var Dracula = Theme{
	Banner: lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true),
	Label:  lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")),
	Value:  lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")),
	Dim:    lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A")),
}

var Nord = Theme{
	Banner: lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")).Bold(true),
	Label:  lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")),
	Value:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ECEFF4")),
	Dim:    lipgloss.NewStyle().Foreground(lipgloss.Color("#4C566A")),
}

func ResolveTheme(name string) Theme {
	switch name {
	case "matrix":
		return Matrix
	case "dracula":
		return Dracula
	case "nord":
		return Nord
	default:
		return Cyber
	}
}

// Options configures static rendering.
type Options struct {
	Theme   Theme
	NoColor bool
	Quiet   bool
}

// Boot prints the banner line to w.
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

// RenderTable prints the full derivation table for r.
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

// RenderSingle prints just the requested field value — pipe-friendly.
func RenderSingle(w io.Writer, value string) {
	_, _ = fmt.Fprintln(w, value)
}
