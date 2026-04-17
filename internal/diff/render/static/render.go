// Package static renders diffs to a terminal writer with ANSI colouring.
package static

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/danielriddell21/unum/internal/diff/node"
	"github.com/danielriddell21/unum/internal/theme"
)

// Theme holds lipgloss styles for static diff rendering.
type Theme struct {
	Added       lipgloss.Style
	Removed     lipgloss.Style
	Modified    lipgloss.Style // structured diffs: changed value
	Unchanged   lipgloss.Style
	HunkHeader  lipgloss.Style
	FileHeader  lipgloss.Style
	LineNumber  lipgloss.Style
	StatAdded   lipgloss.Style
	StatRemoved lipgloss.Style
	Banner      lipgloss.Style
}

func themeFromPalette(p theme.Palette) Theme {
	return Theme{
		Added:       lipgloss.NewStyle().Foreground(lipgloss.Color(p.Added)),
		Removed:     lipgloss.NewStyle().Foreground(lipgloss.Color(p.Removed)),
		Modified:    lipgloss.NewStyle().Foreground(lipgloss.Color(p.NumberVal)),
		Unchanged:   lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)),
		HunkHeader:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary)),
		FileHeader:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary)).Bold(true),
		LineNumber:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)),
		StatAdded:   lipgloss.NewStyle().Foreground(lipgloss.Color(p.Added)).Bold(true),
		StatRemoved: lipgloss.NewStyle().Foreground(lipgloss.Color(p.Removed)).Bold(true),
		Banner:      lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary)).Bold(true),
	}
}

var (
	Cyber   = themeFromPalette(theme.PaletteCyber)
	Matrix  = themeFromPalette(theme.PaletteMatrix)
	Dracula = themeFromPalette(theme.PaletteDracula)
	Nord    = themeFromPalette(theme.PaletteNord)
)

// ResolveTheme returns the named theme, defaulting to Cyber.
func ResolveTheme(name string) Theme {
	return themeFromPalette(theme.ResolvePalette(name))
}

// Options configures static rendering.
type Options struct {
	Theme   Theme
	Context int  // lines of context per hunk (default 3)
	Stat    bool // print summary only, no diff body
	NoColor bool
	Quiet   bool
}

// Boot prints the diff boot sequence to w.
// Call before parsing; invoke the returned func after parsing to complete the line.
func Boot(w io.Writer, fileA, fileB string, opts Options) func(added, removed, modified int, elapsed time.Duration) {
	// noop is the quiet-mode completion callback — intentionally empty.
	noop := func(int, int, int, time.Duration) {}
	if opts.Quiet {
		return noop
	}
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A"))
	if opts.NoColor {
		return func(added, removed, modified int, elapsed time.Duration) {
			_, _ = fmt.Fprintf(w, "[ UNUM ] diff  +%d -%d · %dms ✓\n", added, removed, elapsed.Milliseconds())
		}
	}
	t := opts.Theme
	return func(added, removed, modified int, elapsed time.Duration) {
		addedStr := t.StatAdded.Render(fmt.Sprintf("+%d", added))
		removedStr := t.StatRemoved.Render(fmt.Sprintf("-%d", removed))
		_, _ = fmt.Fprintf(w, "%s %s  %s\n",
			t.Banner.Render("[ UNUM ]"),
			t.Banner.Render("diff  ")+addedStr+" "+removedStr,
			muted.Render(fmt.Sprintf("· %dms ✓", elapsed.Milliseconds())),
		)
	}
}

// Render writes the coloured diff to w.
func Render(w io.Writer, d *node.Diff, opts Options) error {
	if d.Root != nil {
		return renderTree(w, d, opts)
	}
	return renderHunks(w, d, opts)
}

func renderHunks(w io.Writer, d *node.Diff, opts Options) error {
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

func renderTree(w io.Writer, d *node.Diff, opts Options) error {
	t := opts.Theme

	// Summary line
	addedStr := fmt.Sprintf("+%d", d.Added)
	removedStr := fmt.Sprintf("-%d", d.Removed)
	modifiedStr := fmt.Sprintf("~%d", d.Modified)
	if !opts.NoColor {
		addedStr = t.StatAdded.Render(addedStr)
		removedStr = t.StatRemoved.Render(removedStr)
		modifiedStr = t.Modified.Render(modifiedStr)
	}
	_, _ = fmt.Fprintf(w, "%s  %s  %s\n", addedStr, removedStr, modifiedStr)

	if opts.Stat {
		return nil
	}

	if d.Added == 0 && d.Removed == 0 && d.Modified == 0 {
		_, _ = fmt.Fprintln(w, t.Unchanged.Render("  (no differences)"))
		return nil
	}

	walkTreeStatic(w, d.Root, t, opts.NoColor)
	return nil
}

// walkTreeStatic prints only changed nodes (Added/Removed/Modified).
func walkTreeStatic(w io.Writer, dn *node.DiffNode, t Theme, noColor bool) {
	if dn == nil {
		return
	}

	switch dn.Kind {
	case node.Added:
		line := fmt.Sprintf("+ %-40s  %s", dn.Path, dn.NewValue)
		if noColor {
			_, _ = fmt.Fprintln(w, line)
		} else {
			_, _ = fmt.Fprintln(w, t.Added.Render(line))
		}
	case node.Removed:
		line := fmt.Sprintf("- %-40s  %s", dn.Path, dn.OldValue)
		if noColor {
			_, _ = fmt.Fprintln(w, line)
		} else {
			_, _ = fmt.Fprintln(w, t.Removed.Render(line))
		}
	case node.Modified:
		line := fmt.Sprintf("~ %-40s  %s → %s", dn.Path, dn.OldValue, dn.NewValue)
		if noColor {
			_, _ = fmt.Fprintln(w, line)
		} else {
			_, _ = fmt.Fprintln(w, t.Modified.Render(line))
		}
	default:
		// Unchanged container — recurse into children
		for _, child := range dn.Children {
			walkTreeStatic(w, child, t, noColor)
		}
	}
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
