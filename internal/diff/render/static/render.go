// Package static renders diffs to a terminal writer with ANSI colouring.
package static

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/danielriddell21/unum/internal/diff/node"
)

// Theme holds lipgloss styles for static diff rendering.
type Theme struct {
	Added      lipgloss.Style
	Removed    lipgloss.Style
	Unchanged  lipgloss.Style
	HunkHeader lipgloss.Style
	FileHeader lipgloss.Style
	LineNumber lipgloss.Style
	StatAdded  lipgloss.Style
	StatRemoved lipgloss.Style
	Banner     lipgloss.Style
}

// Cyber is the default dark/cyan theme.
var Cyber = Theme{
	Added:       lipgloss.NewStyle().Foreground(lipgloss.Color("#98C379")),
	Removed:     lipgloss.NewStyle().Foreground(lipgloss.Color("#E06C75")),
	Unchanged:   lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A")),
	HunkHeader:  lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4FF")),
	FileHeader:  lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4FF")).Bold(true),
	LineNumber:  lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A")),
	StatAdded:   lipgloss.NewStyle().Foreground(lipgloss.Color("#98C379")).Bold(true),
	StatRemoved: lipgloss.NewStyle().Foreground(lipgloss.Color("#E06C75")).Bold(true),
	Banner:      lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4FF")).Bold(true),
}

// Matrix is the green-on-black theme.
var Matrix = Theme{
	Added:       lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF41")),
	Removed:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF3300")),
	Unchanged:   lipgloss.NewStyle().Foreground(lipgloss.Color("#005500")),
	HunkHeader:  lipgloss.NewStyle().Foreground(lipgloss.Color("#39FF14")),
	FileHeader:  lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF41")).Bold(true),
	LineNumber:  lipgloss.NewStyle().Foreground(lipgloss.Color("#005500")),
	StatAdded:   lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF41")).Bold(true),
	StatRemoved: lipgloss.NewStyle().Foreground(lipgloss.Color("#FF3300")).Bold(true),
	Banner:      lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF41")).Bold(true),
}

// Dracula theme.
var Dracula = Theme{
	Added:       lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")),
	Removed:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")),
	Unchanged:   lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
	HunkHeader:  lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")),
	FileHeader:  lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true),
	LineNumber:  lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
	StatAdded:   lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true),
	StatRemoved: lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true),
	Banner:      lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true),
}

// Nord theme.
var Nord = Theme{
	Added:       lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")),
	Removed:     lipgloss.NewStyle().Foreground(lipgloss.Color("#BF616A")),
	Unchanged:   lipgloss.NewStyle().Foreground(lipgloss.Color("#4C566A")),
	HunkHeader:  lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")),
	FileHeader:  lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")).Bold(true),
	LineNumber:  lipgloss.NewStyle().Foreground(lipgloss.Color("#4C566A")),
	StatAdded:   lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")).Bold(true),
	StatRemoved: lipgloss.NewStyle().Foreground(lipgloss.Color("#BF616A")).Bold(true),
	Banner:      lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")).Bold(true),
}

// ResolveTheme returns the named theme, defaulting to Cyber.
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
	Context int  // lines of context per hunk (default 3)
	Stat    bool // print summary only, no diff body
	NoColor bool
	Quiet   bool
}

// Boot prints the diff header line to w.
func Boot(w io.Writer, fileA, fileB string, opts Options) {
	if opts.Quiet {
		return
	}
	t := opts.Theme
	if opts.NoColor {
		_, _ = fmt.Fprintf(w, "[ UNUM ] diff  %s  →  %s\n", fileA, fileB)
		return
	}
	_, _ = fmt.Fprintln(w, t.Banner.Render(fmt.Sprintf("[ UNUM ] diff  %s  →  %s", fileA, fileB)))
}

// Render writes the coloured diff to w.
func Render(w io.Writer, d *node.Diff, opts Options) error {
	t := opts.Theme

	// Summary line
	addedStr := fmt.Sprintf("+%d", d.Added)
	removedStr := fmt.Sprintf("-%d", d.Removed)
	if !opts.NoColor {
		addedStr = t.StatAdded.Render(addedStr)
		removedStr = t.StatRemoved.Render(removedStr)
	}
	_, _ = fmt.Fprintf(w, "%s  %s\n", addedStr, removedStr)

	if opts.Stat || len(d.Hunks) == 0 {
		return nil
	}

	// File headers
	if !opts.NoColor {
		_, _ = fmt.Fprintln(w, t.FileHeader.Render("--- a/"+d.FileA))
		_, _ = fmt.Fprintln(w, t.FileHeader.Render("+++ b/"+d.FileB))
	} else {
		_, _ = fmt.Fprintln(w, "--- a/"+d.FileA)
		_, _ = fmt.Fprintln(w, "+++ b/"+d.FileB)
	}

	for _, h := range d.Hunks {
		// Hunk header
		header := fmt.Sprintf("@@ -%d,%d +%d,%d @@", h.OldStart, h.OldCount, h.NewStart, h.NewCount)
		if !opts.NoColor {
			_, _ = fmt.Fprintln(w, t.HunkHeader.Render(header))
		} else {
			_, _ = fmt.Fprintln(w, header)
		}

		for _, l := range h.Lines {
			renderLine(w, l, t, opts.NoColor)
		}
	}
	return nil
}

func renderLine(w io.Writer, l node.Line, t Theme, noColor bool) {
	var prefix string
	var lineStyle lipgloss.Style
	var numStyle lipgloss.Style

	switch l.Kind {
	case node.Added:
		prefix = "+"
		lineStyle = t.Added
		numStyle = t.Added
	case node.Removed:
		prefix = "-"
		lineStyle = t.Removed
		numStyle = t.Removed
	default:
		prefix = " "
		lineStyle = t.Unchanged
		numStyle = t.LineNumber
	}

	// Build gutter: "oldN newN" — 0 means blank
	oldStr := "    "
	newStr := "    "
	if l.OldNum > 0 {
		oldStr = fmt.Sprintf("%4d", l.OldNum)
	}
	if l.NewNum > 0 {
		newStr = fmt.Sprintf("%4d", l.NewNum)
	}

	if noColor {
		_, _ = fmt.Fprintf(w, "%s %s %s%s\n", oldStr, newStr, prefix, l.Content)
		return
	}

	gutter := numStyle.Render(oldStr+" "+newStr) + " "
	content := lineStyle.Render(prefix + l.Content)
	// Strip trailing newline from content if any (lipgloss adds none, but source might)
	content = strings.TrimRight(content, "\n")
	_, _ = fmt.Fprintln(w, gutter+content)
}
