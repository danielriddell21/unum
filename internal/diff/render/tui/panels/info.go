package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/danielriddell21/unum/internal/diff/node"
)

// InfoPanel shows diff summary and hunk navigation.
type InfoPanel struct {
	diff     *node.Diff
	viewport viewport.Model
	width    int
	height   int
	focused  bool
	hunkIdx  int // index of visible hunk (for display)
}

// NewInfoPanel creates an info panel.
func NewInfoPanel(d *node.Diff, w, h int) InfoPanel {
	p := InfoPanel{
		diff:   d,
		width:  w,
		height: h,
	}
	p.viewport = viewport.New(w, h)
	p.refresh()
	return p
}

func (p *InfoPanel) SetFocused(f bool) { p.focused = f }

func (p *InfoPanel) Resize(w, h int) {
	p.width = w
	p.height = h
	p.viewport.Width = w
	p.viewport.Height = h
}

// SetHunkIdx updates which hunk is highlighted in the info panel.
func (p *InfoPanel) SetHunkIdx(i int) {
	p.hunkIdx = i
	p.refresh()
}

func (p *InfoPanel) View() string {
	return p.viewport.View()
}

func (p *InfoPanel) refresh() {
	if p.diff == nil {
		p.viewport.SetContent(sbMuted.Render("no diff"))
		return
	}

	var sb strings.Builder

	// File names
	fmt.Fprintf(&sb, "%s\n%s\n\n",
		sbAccent.Render("─── OLD"),
		sbMuted.Render(p.diff.FileA),
	)
	fmt.Fprintf(&sb, "%s\n%s\n\n",
		sbAccent.Render("─── NEW"),
		sbMuted.Render(p.diff.FileB),
	)

	// Summary
	if p.diff.Modified > 0 {
		fmt.Fprintf(&sb, "%s  %s  %s\n\n",
			sbAdded.Render(fmt.Sprintf("+%d added", p.diff.Added)),
			sbRemoved.Render(fmt.Sprintf("-%d removed", p.diff.Removed)),
			sbAccent.Render(fmt.Sprintf("~%d modified", p.diff.Modified)),
		)
	} else {
		fmt.Fprintf(&sb, "%s  %s\n\n",
			sbAdded.Render(fmt.Sprintf("+%d added", p.diff.Added)),
			sbRemoved.Render(fmt.Sprintf("-%d removed", p.diff.Removed)),
		)
	}

	// Hunk info (text diffs only)
	if len(p.diff.Hunks) > 0 {
		fmt.Fprintf(&sb, "%s\n", sbMuted.Render(fmt.Sprintf("Hunk %d of %d", p.hunkIdx+1, len(p.diff.Hunks))))
		h := p.diff.Hunks[p.hunkIdx]
		fmt.Fprintf(&sb, "%s\n\n",
			sbAccent.Render(fmt.Sprintf("@@ -%d,%d +%d,%d @@", h.OldStart, h.OldCount, h.NewStart, h.NewCount)),
		)
	}

	// Format badge (semantic diffs)
	if p.diff.Root != nil {
		fmt.Fprintf(&sb, "%s\n\n", sbAccent.Render("["+p.diff.Format.String()+"]"))
	}

	// Keybindings hint
	if p.diff.Root == nil {
		fmt.Fprintln(&sb, sbMuted.Render("v  toggle split/unified"))
	}
	fmt.Fprintln(&sb, sbMuted.Render("/  search"))
	fmt.Fprintln(&sb, sbMuted.Render("n/N  next/prev match"))
	fmt.Fprintln(&sb, sbMuted.Render("q  quit"))

	p.viewport.SetContent(sb.String())
}
