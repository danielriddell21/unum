package panels

import (
	"fmt"
	"strings"

	"github.com/danielriddell21/unum/internal/hash/types"
)

// HistoryPanel renders the hash history list.
type HistoryPanel struct {
	history []types.HistoryEntry
	cursor  int
	width   int
	height  int
	focused bool
}

// NewHistoryPanel creates a history panel.
func NewHistoryPanel(w, h int) HistoryPanel {
	return HistoryPanel{width: w, height: h}
}

func (p *HistoryPanel) Resize(w, h int) {
	p.width = w
	p.height = h
}

func (p *HistoryPanel) SetFocused(f bool) { p.focused = f }

// SetHistory updates the history entries and clamps the cursor.
func (p *HistoryPanel) SetHistory(h []types.HistoryEntry) {
	p.history = h
	if p.cursor >= len(p.history) {
		p.cursor = 0
	}
}

// ResetCursor moves the cursor to the top of the list.
func (p *HistoryPanel) ResetCursor() { p.cursor = 0 }

func (p *HistoryPanel) ScrollUp() {
	if p.cursor > 0 {
		p.cursor--
	}
}

func (p *HistoryPanel) ScrollDown() {
	if p.cursor < len(p.history)-1 {
		p.cursor++
	}
}

// SelectedInput returns the input string of the currently highlighted entry,
// or "" if history is empty.
func (p *HistoryPanel) SelectedInput() string {
	if len(p.history) == 0 {
		return ""
	}
	return p.history[p.cursor].Input
}

// View renders the panel content (no border, no title).
func (p *HistoryPanel) View() string {
	visible, cursorInView := p.viewport()
	innerW := p.width
	if innerW < 1 {
		innerW = 1
	}

	lines := make([]string, 0, len(visible))
	if len(visible) == 0 {
		lines = append(lines, styleHint.Render("  no history"))
	}
	for i, e := range visible {
		text := e.Input
		maxLen := innerW - 4
		if maxLen < 1 {
			maxLen = 1
		}
		if len([]rune(text)) > maxLen {
			runes := []rune(text)
			text = string(runes[:maxLen]) + "…"
		}
		entry := fmt.Sprintf("  %s", text)
		if p.focused && i == cursorInView {
			lines = append(lines, styleHistorySelected.Render("> "+strings.TrimPrefix(entry, "  ")))
		} else {
			lines = append(lines, styleValue.Render(entry))
		}
	}
	return strings.Join(lines, "\n")
}

func (p *HistoryPanel) capacity() int {
	usable := p.height - 1 // reserve 1 row for title
	if usable < 1 {
		return 1
	}
	return usable
}

func (p *HistoryPanel) viewport() (visible []types.HistoryEntry, cursorInView int) {
	cap := p.capacity()
	total := len(p.history)
	if total == 0 {
		return nil, 0
	}
	start := p.cursor - cap + 1
	if start < 0 {
		start = 0
	}
	if start > total-cap {
		start = total - cap
	}
	if start < 0 {
		start = 0
	}
	end := start + cap
	if end > total {
		end = total
	}
	return p.history[start:end], p.cursor - start
}
