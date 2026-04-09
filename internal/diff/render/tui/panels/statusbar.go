// Package panels contains Bubble Tea sub-models for the diff TUI.
package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/danielriddell21/unum/internal/diff/node"
	jsonpanels "github.com/danielriddell21/unum/internal/json/render/tui/panels"
)

var (
	sbAdded   = lipgloss.NewStyle().Foreground(lipgloss.Color(jsonpanels.PaletteCyber.Added)).Bold(true)
	sbRemoved = lipgloss.NewStyle().Foreground(lipgloss.Color(jsonpanels.PaletteCyber.Removed)).Bold(true)
	sbMuted   = lipgloss.NewStyle().Foreground(lipgloss.Color(jsonpanels.PaletteCyber.Muted))
	sbAccent  = lipgloss.NewStyle().Foreground(lipgloss.Color(jsonpanels.PaletteCyber.AccentPrimary))
)

// ApplyPalette updates status bar styles from a palette.
func ApplyPalette(p jsonpanels.Palette) {
	sbAdded = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Added)).Bold(true)
	sbRemoved = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Removed)).Bold(true)
	sbMuted = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	sbAccent = lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary))
}

// StatusBar renders the bottom status bar for the diff TUI.
func StatusBar(d *node.Diff, view string, width int, searchMode bool, searchQuery string) string {
	if searchMode {
		prompt := sbAccent.Render("[ SEARCH > ")
		cursor := sbAccent.Render("█")
		q := ""
		if searchQuery != "" {
			q = sbAccent.Render(searchQuery)
		}
		close := sbAccent.Render(" ]")
		hint := sbMuted.Render("  type to filter · enter:confirm · esc:clear")
		line := prompt + q + cursor + close + hint
		return pad(line, width)
	}

	added := sbAdded.Render(fmt.Sprintf("+%d", d.Added))
	removed := sbRemoved.Render(fmt.Sprintf("-%d", d.Removed))
	viewLabel := sbMuted.Render("[" + view + "]")

	left := " " + added + "  " + removed + "  " + viewLabel
	hints := sbMuted.Render("v:toggle  j/k:scroll  /:search  n/N:match  q:quit") + " "

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(hints)
	mid := width - leftW - rightW
	if mid < 1 {
		mid = 1
	}
	return left + strings.Repeat(" ", mid) + hints
}

func pad(s string, width int) string {
	w := lipgloss.Width(s)
	if w < width {
		s += strings.Repeat(" ", width-w)
	}
	return s
}
