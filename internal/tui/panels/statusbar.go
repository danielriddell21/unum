package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/danielriddell21/unum/internal/json/node"
)


// StatusBar renders the bottom status bar.
// It's a pure render function — no sub-model state.
func StatusBar(n *node.Node, width int, searchMode bool, searchQuery string, yankFeedback string) string {
	if searchMode {
		prompt := styleSearchPrompt.Render("[ SCANNING > ")
		cursor := styleSearchPrompt.Render("█")
		q := ""
		if searchQuery != "" {
			q = styleSearchPrompt.Render(searchQuery)
		}
		close := styleSearchPrompt.Render(" ]")
		hint := styleSBHint.Render("  type to filter · enter:confirm · esc:clear")
		line := prompt + q + cursor + close + hint
		return Pad(line, width)
	}

	if n == nil {
		return Pad(styleSBHint.Render("  no selection"), width)
	}

	path := styleSBPath.Render("[ " + n.Path() + " ]")
	kind := styleSBType.Render(strings.ToUpper(n.Kind.String()))

	extra := ""
	switch n.Kind {
	case node.KindObject:
		extra = styleSBHint.Render(fmt.Sprintf("(%d keys)", len(n.Children)))
	case node.KindArray:
		extra = styleSBHint.Render(fmt.Sprintf("(%d items)", len(n.Children)))
	}

	yank := ""
	if yankFeedback != "" {
		yank = styleSBType.Render(" ✓ " + yankFeedback + " ")
	}

	hints := styleSBHint.Render("tab:panel  1-6:lens  /:search  y:yank  ?:help  q:quit")

	parts := []string{path, kind}
	if extra != "" {
		parts = append(parts, extra)
	}
	if yank != "" {
		parts = append(parts, yank)
	}

	left := " " + strings.Join(parts, styleSBSep)
	right := hints + " "

	// Fill middle space
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	mid := width - leftW - rightW
	if mid < 1 {
		mid = 1
	}
	return left + strings.Repeat(" ", mid) + right
}

// Pad right-pads s with spaces to fill width.
func Pad(s string, width int) string {
	w := lipgloss.Width(s)
	if w < width {
		s += strings.Repeat(" ", width-w)
	}
	return s
}
