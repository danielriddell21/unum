package panels

import (
	"fmt"
	"strings"

	"github.com/danielriddell21/unum/internal/json/node"
	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

// StatusBar renders the bottom status bar for the JSON TUI.
func StatusBar(n *node.Node, width int, searchMode bool, searchQuery string, yankFeedback string) string {
	if searchMode {
		hint := styleMuted.Render("  type to filter · enter:confirm · esc:clear")
		return tuipanels.SearchBar("SCANNING", searchQuery, hint, width, styleSearchPrompt)
	}

	if n == nil {
		return tuipanels.Pad(styleMuted.Render("  no selection"), width)
	}

	path := stylePath.Render("[ " + n.Path() + " ]")
	kind := styleSBType.Render(strings.ToUpper(n.Kind.String()))

	extra := ""
	switch n.Kind {
	case node.KindObject:
		extra = styleMuted.Render(fmt.Sprintf("(%d keys)", len(n.Children)))
	case node.KindArray:
		extra = styleMuted.Render(fmt.Sprintf("(%d items)", len(n.Children)))
	}

	yank := ""
	if yankFeedback != "" {
		yank = styleSBType.Render(" ✓ " + yankFeedback + " ")
	}

	hints := styleMuted.Render("tab:panel  1-6:lens  /:search  y:yank  ?:help  q:quit")

	parts := []string{path, kind}
	if extra != "" {
		parts = append(parts, extra)
	}
	if yank != "" {
		parts = append(parts, yank)
	}

	left := " " + strings.Join(parts, styleSBSep)
	right := hints + " "
	return tuipanels.Bar(left, right, width)
}
