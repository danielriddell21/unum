// Package static renders JSON node trees to the terminal with syntax highlighting
// and optional annotation overlays (stats, merkle hashes).
package static

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/danielriddell21/unum/internal/json/lens/merkle"
	"github.com/danielriddell21/unum/internal/json/lens/stats"
	"github.com/danielriddell21/unum/internal/json/node"
	"github.com/danielriddell21/unum/internal/theme"
)

// Theme holds the color palette for static rendering.
type Theme struct {
	ObjectKey   lipgloss.Style
	ArrayIndex  lipgloss.Style
	StringVal   lipgloss.Style
	NumberVal   lipgloss.Style
	BoolVal     lipgloss.Style
	NullVal     lipgloss.Style
	Punctuation lipgloss.Style
	LineNumber  lipgloss.Style
	HashTag     lipgloss.Style
	StatTag     lipgloss.Style
	ErrorStyle  lipgloss.Style
	Banner      lipgloss.Style
	BannerOK    lipgloss.Style
}

func themeFromPalette(p theme.Palette) Theme {
	return Theme{
		ObjectKey:   lipgloss.NewStyle().Foreground(lipgloss.Color(p.ObjectKey)),
		ArrayIndex:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.ArrayIndex)),
		StringVal:   lipgloss.NewStyle().Foreground(lipgloss.Color(p.StringVal)),
		NumberVal:   lipgloss.NewStyle().Foreground(lipgloss.Color(p.NumberVal)),
		BoolVal:     lipgloss.NewStyle().Foreground(lipgloss.Color(p.BoolTrue)),
		NullVal:     lipgloss.NewStyle().Foreground(lipgloss.Color(p.NullVal)),
		Punctuation: lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)),
		LineNumber:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)),
		HashTag:     lipgloss.NewStyle().Foreground(lipgloss.Color(p.Hash)),
		StatTag:     lipgloss.NewStyle().Foreground(lipgloss.Color(p.Stats)),
		ErrorStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color(p.Error)),
		Banner:      lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary)),
		BannerOK:    lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary)).Bold(true),
	}
}

var (
	Cyber   = themeFromPalette(theme.PaletteCyber)
	Matrix  = themeFromPalette(theme.PaletteMatrix)
	Dracula = themeFromPalette(theme.PaletteDracula)
	Nord    = themeFromPalette(theme.PaletteNord)
)

// Options controls rendering behaviour.
type Options struct {
	Theme        Theme
	ShowStats    bool
	ShowMerkle   bool
	ShowLineNums bool
	Compact      bool
	Quiet        bool
	NoColor      bool
	Filename     string
}

// ResolveTheme returns the named theme, defaulting to Cyber.
func ResolveTheme(name string) Theme {
	return themeFromPalette(theme.ResolvePalette(name))
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Theme:        Cyber,
		ShowLineNums: true,
	}
}

// Boot prints the neural-interface boot sequence to stderr.
// Call before Render. Suppressed when opts.Quiet is true.
func Boot(w io.Writer, filename string, opts Options) func(nodeCount int, elapsed time.Duration) {
	if opts.Quiet || opts.NoColor {
		return func(nodeCount int, elapsed time.Duration) {
			if !opts.Quiet {
				_, _ = fmt.Fprintf(w, "[ UNUM ] json  %d nodes · %dms ✓\n", nodeCount, elapsed.Milliseconds())
			}
		}
	}

	t := opts.Theme
	_, _ = fmt.Fprintf(w, "%s %s\r",
		t.BannerOK.Render("[ UNUM ]"),
		t.Punctuation.Render("json  scanning..."),
	)

	return func(nodeCount int, elapsed time.Duration) {
		_, _ = fmt.Fprintf(w, "%s %s  %s\n",
			t.BannerOK.Render("[ UNUM ]"),
			t.BannerOK.Render(fmt.Sprintf("json  %d nodes", nodeCount)),
			t.Punctuation.Render(fmt.Sprintf("· %dms ✓", elapsed.Milliseconds())),
		)
	}
}

// Render writes a colorized JSON tree to w.
func Render(w io.Writer, root *node.Node, opts Options) error {
	r := &renderer{w: w, opts: opts, lineNum: 1}

	if opts.Compact {
		b, err := root.MarshalJSON()
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		_, _ = fmt.Fprintln(w, string(b))
		return nil
	}

	r.renderNode(root, 0)
	_, _ = fmt.Fprintln(w)
	return nil
}

// RenderError prints a formatted error to w.
func RenderError(w io.Writer, err error, opts Options) {
	if opts.NoColor {
		_, _ = fmt.Fprintf(w, "error: %s\n", err)
		return
	}
	_, _ = fmt.Fprintf(w, "%s %s\n", opts.Theme.ErrorStyle.Render("✗"), err)
}

type renderer struct {
	w       io.Writer
	opts    Options
	lineNum int
	buf     strings.Builder
}

func (r *renderer) renderNode(n *node.Node, depth int) {
	indent := strings.Repeat("  ", depth)

	switch n.Kind {
	case node.KindObject:
		r.renderObject(n, depth, indent)
	case node.KindArray:
		r.renderArray(n, depth, indent)
	default:
		r.renderLeaf(n, indent)
	}
}

func (r *renderer) renderObject(n *node.Node, depth int, indent string) {
	t := r.opts.Theme

	if len(n.Children) == 0 {
		r.writeLine(indent + t.Punctuation.Render("{}"))
		r.appendMerkleTag(n)
		return
	}

	r.writeLine(indent + t.Punctuation.Render("{"))
	r.appendMerkleTag(n)

	for i, child := range n.Children {
		isLast := i == len(n.Children)-1

		keyStr := t.ObjectKey.Render(`"`+child.Key+`"`) +
			t.Punctuation.Render(": ")

		childIndent := strings.Repeat("  ", depth+1)

		switch child.Kind {
		case node.KindObject, node.KindArray:
			r.startLine()
			r.writef("%s%s", childIndent, keyStr)
			r.renderInline(child, depth+1)
			if !isLast {
				r.writes(t.Punctuation.Render(","))
			}
			r.endLine()
		default:
			r.startLine()
			r.writef("%s%s%s", childIndent, keyStr, r.leafColor(child))
			if !isLast {
				r.writes(t.Punctuation.Render(","))
			}
			r.appendStatsTag(child)
			r.endLine()
		}
	}

	r.writeLine(indent + t.Punctuation.Render("}"))
}

func (r *renderer) renderArray(n *node.Node, depth int, indent string) {
	t := r.opts.Theme

	if len(n.Children) == 0 {
		r.writeLine(indent + t.Punctuation.Render("[]"))
		return
	}

	r.startLine()
	r.writes(indent + t.Punctuation.Render("["))
	r.appendStatsTag(n)
	r.endLine()

	for i, child := range n.Children {
		isLast := i == len(n.Children)-1
		childIndent := strings.Repeat("  ", depth+1)

		switch child.Kind {
		case node.KindObject, node.KindArray:
			r.startLine()
			r.writes(childIndent)
			r.renderInline(child, depth+1)
			if !isLast {
				r.writes(t.Punctuation.Render(","))
			}
			r.endLine()
		default:
			r.startLine()
			r.writef("%s%s", childIndent, r.leafColor(child))
			if !isLast {
				r.writes(t.Punctuation.Render(","))
			}
			r.endLine()
		}
	}

	r.writeLine(indent + t.Punctuation.Render("]"))
}

func (r *renderer) renderLeaf(n *node.Node, indent string) {
	r.writeLine(indent + r.leafColor(n))
}

// renderInline renders an object or array opening + children inline from the current position.
func (r *renderer) renderInline(n *node.Node, depth int) { //nolint:gocognit // renders every value kind inline; branching on kind is the algorithm
	t := r.opts.Theme
	indent := strings.Repeat("  ", depth)

	switch n.Kind {
	case node.KindObject:
		if len(n.Children) == 0 {
			r.writes(t.Punctuation.Render("{}"))
			r.appendMerkleTag(n)
			return
		}
		r.writes(t.Punctuation.Render("{"))
		r.appendMerkleTag(n)
		r.endLine()

		for i, child := range n.Children {
			isLast := i == len(n.Children)-1
			childIndent := strings.Repeat("  ", depth+1)
			keyStr := t.ObjectKey.Render(`"`+child.Key+`"`) + t.Punctuation.Render(": ")

			switch child.Kind {
			case node.KindObject, node.KindArray:
				r.startLine()
				r.writef("%s%s", childIndent, keyStr)
				r.renderInline(child, depth+1)
				if !isLast {
					r.writes(t.Punctuation.Render(","))
				}
				r.endLine()
			default:
				r.startLine()
				r.writef("%s%s%s", childIndent, keyStr, r.leafColor(child))
				if !isLast {
					r.writes(t.Punctuation.Render(","))
				}
				r.endLine()
			}
		}
		r.startLine()
		r.writes(indent + t.Punctuation.Render("}"))

	case node.KindArray:
		if len(n.Children) == 0 {
			r.writes(t.Punctuation.Render("[]"))
			return
		}
		r.writes(t.Punctuation.Render("["))
		r.appendStatsTag(n)
		r.endLine()

		for i, child := range n.Children {
			isLast := i == len(n.Children)-1
			childIndent := strings.Repeat("  ", depth+1)
			switch child.Kind {
			case node.KindObject, node.KindArray:
				r.startLine()
				r.writes(childIndent)
				r.renderInline(child, depth+1)
				if !isLast {
					r.writes(t.Punctuation.Render(","))
				}
				r.endLine()
			default:
				r.startLine()
				r.writef("%s%s", childIndent, r.leafColor(child))
				if !isLast {
					r.writes(t.Punctuation.Render(","))
				}
				r.endLine()
			}
		}
		r.startLine()
		r.writes(indent + t.Punctuation.Render("]"))
	}
}

func (r *renderer) leafColor(n *node.Node) string {
	t := r.opts.Theme
	switch n.Kind {
	case node.KindString:
		var s string
		if json.Unmarshal([]byte(n.Raw), &s) == nil {
			return t.StringVal.Render(`"` + escapeForDisplay(s) + `"`)
		}
		return t.StringVal.Render(n.Raw)
	case node.KindNumber:
		return t.NumberVal.Render(n.Raw)
	case node.KindBool:
		return t.BoolVal.Render(n.Raw)
	case node.KindNull:
		return t.NullVal.Render("null")
	}
	return n.Raw
}

// appendMerkleTag appends a merkle hash tag to the current buffered line.
func (r *renderer) appendMerkleTag(n *node.Node) {
	if !r.opts.ShowMerkle {
		return
	}
	if v, ok := n.GetAnnotation(merkle.Lens, "hash"); ok {
		if h, ok := v.(string); ok {
			r.writes("  " + r.opts.Theme.HashTag.Render("#"+merkle.Short(h)))
		}
	}
}

// appendStatsTag appends a stats summary to the current buffered line.
func (r *renderer) appendStatsTag(n *node.Node) {
	if !r.opts.ShowStats || n.Kind != node.KindArray {
		return
	}
	nc, ok := n.GetAnnotation(stats.Lens, "numeric_count")
	if !ok {
		return
	}
	numCount := nc.(int)
	if numCount == 0 {
		return
	}
	parts := []string{}
	if v, ok := n.GetAnnotation(stats.Lens, "count"); ok {
		parts = append(parts, fmt.Sprintf("n=%d", v.(int)))
	}
	if v, ok := n.GetAnnotation(stats.Lens, "min"); ok {
		parts = append(parts, fmt.Sprintf("min=%s", fmtFloat(v.(float64))))
	}
	if v, ok := n.GetAnnotation(stats.Lens, "max"); ok {
		parts = append(parts, fmt.Sprintf("max=%s", fmtFloat(v.(float64))))
	}
	if v, ok := n.GetAnnotation(stats.Lens, "mean"); ok {
		parts = append(parts, fmt.Sprintf("mean=%s", fmtFloat(v.(float64))))
	}
	r.writes("  " + r.opts.Theme.StatTag.Render("// "+strings.Join(parts, " ")))
}

func fmtFloat(f float64) string {
	s := fmt.Sprintf("%.4g", f)
	return s
}

// line buffer helpers

func (r *renderer) startLine() {
	r.buf.Reset()
}

// writes appends a pre-formatted string to the current line buffer.
func (r *renderer) writes(s string) {
	r.buf.WriteString(s)
}

// writef appends a formatted string to the current line buffer.
func (r *renderer) writef(format string, args ...any) {
	fmt.Fprintf(&r.buf, format, args...)
}

func (r *renderer) endLine() {
	line := r.buf.String()
	r.buf.Reset()
	if r.opts.ShowLineNums {
		lineNum := r.opts.Theme.LineNumber.Render(fmt.Sprintf("%4d", r.lineNum))
		_, _ = fmt.Fprintf(r.w, "%s  %s\n", lineNum, line)
	} else {
		_, _ = fmt.Fprintln(r.w, line)
	}
	r.lineNum++
}

func (r *renderer) writeLine(s string) {
	r.startLine()
	r.writes(s)
	r.endLine()
}

func escapeForDisplay(s string) string {
	// Escape control characters for display but keep the string readable
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
