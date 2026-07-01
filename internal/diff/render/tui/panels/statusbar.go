package panels

import (
	"fmt"

	"github.com/danielriddell21/unum/internal/diff/node"
	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

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

	added := sbAdded.Render(fmt.Sprintf("Add: %d", d.Added))
	changed := sbModified.Render(fmt.Sprintf("Change: %d", d.Modified))
	removed := sbRemoved.Render(fmt.Sprintf("Delete: %d", d.Removed))
	viewLabel := sbMuted.Render("[" + view + "]")

	left := " " + added + "  " + changed + "  " + removed + "  " + viewLabel
	right := sbMuted.Render("v:toggle  j/k:scroll  /:search  n/N:match  ?:help  q:quit") + "  " + sbMuted.Render(version) + " "
	return tuipanels.Bar(left, right, width)
}
