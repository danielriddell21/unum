package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/danielriddell21/unum/internal/diff/node"
)

const noDifferencesPadded = "  (no differences)"

// UnifiedPanel renders the diff in classic unified format.
type UnifiedPanel struct {
	diff     *node.Diff
	viewport viewport.Model
	width    int
	height   int
	focused  bool
	search   string
	matchIdx int
	matches  []int // line indices with search matches
	lines    []renderedLine
	textOnly bool // force text hunk rendering even for semantic diffs
}

type renderedLine struct {
	raw     string // plain text for search
	display string // styled
}

// NewUnifiedPanel creates a unified diff panel.
func NewUnifiedPanel(d *node.Diff, w, h int) UnifiedPanel {
	p := UnifiedPanel{diff: d, width: w, height: h}
	p.viewport = viewport.New(w, h)
	p.build()
	return p
}

func (p *UnifiedPanel) SetFocused(f bool) { p.focused = f }
func (p *UnifiedPanel) SetTextOnly(v bool) {
	if p.textOnly == v {
		return
	}
	p.textOnly = v
	p.build()
}

func (p *UnifiedPanel) Resize(w, h int) {
	p.width = w
	p.height = h
	p.viewport.Width = w
	p.viewport.Height = h
}

func (p *UnifiedPanel) SetSearch(q string) {
	p.search = q
	p.matchIdx = 0
	p.build()
}

func (p *UnifiedPanel) NextMatch() {
	if len(p.matches) == 0 {
		return
	}
	p.matchIdx = (p.matchIdx + 1) % len(p.matches)
	p.scrollToMatch()
}

func (p *UnifiedPanel) PrevMatch() {
	if len(p.matches) == 0 {
		return
	}
	p.matchIdx = (p.matchIdx - 1 + len(p.matches)) % len(p.matches)
	p.scrollToMatch()
}

func (p *UnifiedPanel) MatchCount() int { return len(p.matches) }

func (p *UnifiedPanel) scrollToMatch() {
	if p.matchIdx < len(p.matches) {
		p.viewport.SetYOffset(p.matches[p.matchIdx])
	}
}

// CurrentHunkIdx returns the hunk index based on viewport position.
func (p *UnifiedPanel) CurrentHunkIdx() int {
	return 0 // simplified — info panel shows hunk 1 by default for now
}

func (p *UnifiedPanel) ScrollDown(n int) { p.viewport.ScrollDown(n) }
func (p *UnifiedPanel) ScrollUp(n int)   { p.viewport.ScrollUp(n) }
func (p *UnifiedPanel) HalfPageDown()    { p.viewport.HalfPageDown() }
func (p *UnifiedPanel) HalfPageUp()      { p.viewport.HalfPageUp() }

func (p *UnifiedPanel) View() string {
	return p.viewport.View()
}

func (p *UnifiedPanel) build() {
	p.lines = nil
	p.matches = nil

	if p.diff == nil {
		p.viewport.SetContent(diffUnchanged.Render(noDifferencesPadded))
		return
	}

	if p.diff.Root != nil && !p.textOnly {
		p.buildJSON()
		return
	}

	if len(p.diff.Hunks) == 0 {
		p.viewport.SetContent(diffUnchanged.Render(noDifferencesPadded))
		return
	}

	q := strings.ToLower(p.search)

	for _, h := range p.diff.Hunks {
		hdr := fmt.Sprintf("@@ -%d,%d +%d,%d @@", h.OldStart, h.OldCount, h.NewStart, h.NewCount)
		p.lines = append(p.lines, renderedLine{
			raw:     hdr,
			display: diffHunkHdr.Render(hdr),
		})

		for _, l := range h.Lines {
			raw, display := p.renderLine(l, q)
			if q != "" && strings.Contains(strings.ToLower(raw), q) {
				p.matches = append(p.matches, len(p.lines))
			}
			p.lines = append(p.lines, renderedLine{raw: raw, display: display})
		}
	}

	var sb strings.Builder
	for _, l := range p.lines {
		sb.WriteString(l.display)
		sb.WriteByte('\n')
	}
	p.viewport.SetContent(sb.String())
}

func (p *UnifiedPanel) renderLine(l node.Line, searchQ string) (raw, display string) {
	var prefix string
	var style lipgloss.Style
	var numStyle lipgloss.Style

	switch l.Kind {
	case node.Added:
		prefix = "+"
		style = diffAdded
		numStyle = diffAdded
	case node.Removed:
		prefix = "-"
		style = diffRemoved
		numStyle = diffRemoved
	default:
		prefix = " "
		style = diffUnchanged
		numStyle = diffLineNum
	}

	oldStr := "    "
	newStr := "    "
	if l.OldNum > 0 {
		oldStr = fmt.Sprintf("%4d", l.OldNum)
	}
	if l.NewNum > 0 {
		newStr = fmt.Sprintf("%4d", l.NewNum)
	}

	gutter := numStyle.Render(oldStr+" "+newStr) + " "
	content := prefix + l.Content
	raw = oldStr + " " + newStr + " " + content

	// Highlight search match
	if searchQ != "" {
		low := strings.ToLower(l.Content)
		if idx := strings.Index(low, searchQ); idx >= 0 {
			before := l.Content[:idx]
			match := l.Content[idx : idx+len(searchQ)]
			after := l.Content[idx+len(searchQ):]
			content = prefix + style.Render(before) + diffSearchHL.Render(match) + style.Render(after)
			display = gutter + content
			return raw, display
		}
	}

	display = gutter + style.Render(content)
	return raw, display
}

// SplitPanel renders old and new file side-by-side with synchronised scrolling.
type SplitPanel struct {
	diff     *node.Diff
	left     viewport.Model // old file
	right    viewport.Model // new file
	focused  int            // 0=left, 1=right
	width    int
	height   int
	search   string
	matches  []int
	matchIdx int
}

// NewSplitPanel creates a side-by-side diff panel.
func NewSplitPanel(d *node.Diff, w, h int) SplitPanel {
	hw := (w - 1) / 2 // -1 for the centre divider
	p := SplitPanel{diff: d, width: w, height: h}
	p.left = viewport.New(hw, h)
	p.right = viewport.New(w-hw-1, h)
	p.build()
	return p
}

func (p *SplitPanel) SetSearch(q string) {
	p.search = q
	p.matchIdx = 0
	p.build()
}

func (p *SplitPanel) NextMatch() {
	if len(p.matches) == 0 {
		return
	}
	p.matchIdx = (p.matchIdx + 1) % len(p.matches)
	p.left.SetYOffset(p.matches[p.matchIdx])
	p.right.SetYOffset(p.matches[p.matchIdx])
}

func (p *SplitPanel) PrevMatch() {
	if len(p.matches) == 0 {
		return
	}
	p.matchIdx = (p.matchIdx - 1 + len(p.matches)) % len(p.matches)
	p.left.SetYOffset(p.matches[p.matchIdx])
	p.right.SetYOffset(p.matches[p.matchIdx])
}

func (p *SplitPanel) MatchCount() int { return len(p.matches) }

func (p *SplitPanel) Resize(w, h int) {
	p.width = w
	p.height = h
	hw := (w - 1) / 2
	p.left.Width = hw
	p.left.Height = h
	p.right.Width = w - hw - 1
	p.right.Height = h
}

func (p *SplitPanel) ScrollDown(n int) {
	p.left.ScrollDown(n)
	p.right.ScrollDown(n)
}

func (p *SplitPanel) ScrollUp(n int) {
	p.left.ScrollUp(n)
	p.right.ScrollUp(n)
}

func (p *SplitPanel) HalfPageDown() {
	p.left.HalfPageDown()
	p.right.HalfPageDown()
}

func (p *SplitPanel) HalfPageUp() {
	p.left.HalfPageUp()
	p.right.HalfPageUp()
}
func (p *SplitPanel) CycleFocus() { p.focused = 1 - p.focused }

func (p *SplitPanel) View() string {
	lv := p.left.View()
	rv := p.right.View()

	lLines := strings.Split(lv, "\n")
	rLines := strings.Split(rv, "\n")

	// Pad to same height
	maxL := len(lLines)
	if len(rLines) > maxL {
		maxL = len(rLines)
	}
	for len(lLines) < maxL {
		lLines = append(lLines, "")
	}
	for len(rLines) < maxL {
		rLines = append(rLines, "")
	}

	sep := diffUnchanged.Render("│")
	var sb strings.Builder
	for i := range lLines {
		sb.WriteString(lLines[i])
		sb.WriteString(sep)
		sb.WriteString(rLines[i])
		if i < len(lLines)-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

func (p *SplitPanel) build() { //nolint:cyclop,gocognit // NOSONAR: walks diff tree building side-by-side lines; branching on kind/added/removed is the algorithm
	if p.diff == nil {
		return
	}
	if len(p.diff.Hunks) == 0 {
		p.left.SetContent(diffUnchanged.Render("(no differences)"))
		p.right.SetContent(diffUnchanged.Render("(no differences)"))
		return
	}

	q := strings.ToLower(p.search)
	p.matches = nil

	var leftLines, rightLines []string

	for _, h := range p.diff.Hunks {
		hdr := diffHunkHdr.Render(fmt.Sprintf("@@ -%d,%d +%d,%d @@", h.OldStart, h.OldCount, h.NewStart, h.NewCount))
		leftLines = append(leftLines, hdr)
		rightLines = append(rightLines, hdr)

		// Pair up removed/added lines so they sit on the same row
		var removed, added []node.Line
		var unchanged []node.Line
		flush := func() {
			maxPairs := len(removed)
			if len(added) > maxPairs {
				maxPairs = len(added)
			}
			for i := 0; i < maxPairs; i++ {
				lineIdx := len(leftLines)
				var lLine, rLine string
				if i < len(removed) {
					lLine = p.renderSide(removed[i], node.Removed, q)
					if q != "" && strings.Contains(strings.ToLower(removed[i].Content), q) {
						p.matches = append(p.matches, lineIdx)
					}
				} else {
					lLine = ""
				}
				if i < len(added) {
					rLine = p.renderSide(added[i], node.Added, q)
					if q != "" && strings.Contains(strings.ToLower(added[i].Content), q) {
						p.matches = append(p.matches, lineIdx)
					}
				} else {
					rLine = ""
				}
				leftLines = append(leftLines, lLine)
				rightLines = append(rightLines, rLine)
			}
			removed = removed[:0]
			added = added[:0]
			for _, u := range unchanged {
				leftLines = append(leftLines, p.renderSide(u, node.Unchanged, q))
				rightLines = append(rightLines, p.renderSide(u, node.Unchanged, q))
			}
			unchanged = unchanged[:0]
		}

		for _, l := range h.Lines {
			switch l.Kind {
			case node.Removed:
				if len(unchanged) > 0 {
					flush()
				}
				removed = append(removed, l)
			case node.Added:
				added = append(added, l)
			default:
				if len(removed) > 0 || len(added) > 0 {
					flush()
				}
				unchanged = append(unchanged, l)
			}
		}
		flush()
	}

	p.left.SetContent(strings.Join(leftLines, "\n"))
	p.right.SetContent(strings.Join(rightLines, "\n"))
}

func (p *SplitPanel) renderSide(l node.Line, kind node.ChangeKind, searchQ string) string {
	var style lipgloss.Style
	var numStyle lipgloss.Style
	var prefix string

	switch kind {
	case node.Added:
		style = diffAdded
		numStyle = diffAdded
		prefix = "+"
	case node.Removed:
		style = diffRemoved
		numStyle = diffRemoved
		prefix = "-"
	default:
		style = diffUnchanged
		numStyle = diffLineNum
		prefix = " "
	}

	num := "    "
	if kind != node.Added && l.OldNum > 0 {
		num = fmt.Sprintf("%4d", l.OldNum)
	} else if kind != node.Removed && l.NewNum > 0 {
		num = fmt.Sprintf("%4d", l.NewNum)
	}

	gutter := numStyle.Render(num) + " "
	content := l.Content

	if searchQ != "" {
		low := strings.ToLower(content)
		if idx := strings.Index(low, searchQ); idx >= 0 {
			before := content[:idx]
			match := content[idx : idx+len(searchQ)]
			after := content[idx+len(searchQ):]
			return gutter + style.Render(prefix+before) + diffSearchHL.Render(match) + style.Render(after)
		}
	}

	return gutter + style.Render(prefix+content)
}

// buildJSON renders the DiffNode tree as a flat list of changed paths.
func (p *UnifiedPanel) buildJSON() {
	if p.diff.Root == nil {
		p.viewport.SetContent(diffUnchanged.Render(noDifferencesPadded))
		return
	}

	if p.diff.Added == 0 && p.diff.Removed == 0 && p.diff.Modified == 0 {
		p.viewport.SetContent(diffUnchanged.Render(noDifferencesPadded))
		return
	}

	q := strings.ToLower(p.search)

	var walkFn func(dn *node.DiffNode)
	walkFn = func(dn *node.DiffNode) {
		if dn == nil {
			return
		}
		var raw, display string
		switch dn.Kind {
		case node.Added:
			raw = fmt.Sprintf("+ %-40s  %s", dn.Path, dn.NewValue)
			display = diffAdded.Render(raw)
		case node.Removed:
			raw = fmt.Sprintf("- %-40s  %s", dn.Path, dn.OldValue)
			display = diffRemoved.Render(raw)
		case node.Modified:
			raw = fmt.Sprintf("~ %-40s  %s → %s", dn.Path, dn.OldValue, dn.NewValue)
			display = diffHunkHdr.Render(raw)
		default:
			// Unchanged container — recurse only
			for _, child := range dn.Children {
				walkFn(child)
			}
			return
		}

		lineIdx := len(p.lines)
		if q != "" && strings.Contains(strings.ToLower(raw), q) {
			p.matches = append(p.matches, lineIdx)
		}
		p.lines = append(p.lines, renderedLine{raw: raw, display: display})
	}
	walkFn(p.diff.Root)

	var sb strings.Builder
	for _, l := range p.lines {
		sb.WriteString(l.display)
		sb.WriteByte('\n')
	}
	p.viewport.SetContent(sb.String())
}
