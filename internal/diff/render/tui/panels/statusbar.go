// Package panels contains Bubble Tea sub-models for the diff TUI.
package panels

import (
	"fmt"

	"github.com/danielriddell21/unum/internal/diff/node"
	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

// StatusBar renders the bottom status bar for the diff TUI.
func StatusBar(d *node.Diff, view string, width int, searchMode bool, searchQuery string, matchCount int, version string) string {
	if searchMode {
		var matchHint string
		if searchQuery != "" {
			if matchCount == 0 {
				matchHint = sbMuted.Render("  no matches")
			} else {
				matchHint = sbMuted.Render(fmt.Sprintf("  %d matches", matchCount))
			}
		}
		hint := sbMuted.Render("  type to filter · enter:confirm · esc:clear")
		return tuipanels.SearchBar("SEARCH", searchQuery, matchHint+hint, width, sbAccent)
	}

	added := sbAdded.Render(fmt.Sprintf("+%d", d.Added))
	removed := sbRemoved.Render(fmt.Sprintf("-%d", d.Removed))
	viewLabel := sbMuted.Render("[" + view + "]")

	left := " " + added + "  " + removed + "  " + viewLabel
	right := sbMuted.Render("v:toggle  j/k:scroll  /:search  n/N:match  ?:help  q:quit") + "  " + sbMuted.Render(version) + " "
	return tuipanels.Bar(left, right, width)
}
