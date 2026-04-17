package panels

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Bar lays out left + spaces + right to fill width.
func Bar(left, right string, width int) string {
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	mid := width - leftW - rightW
	if mid < 1 {
		mid = 1
	}
	return left + strings.Repeat(" ", mid) + right
}

// SearchBar renders a search-mode prompt line padded to width.
// label is the prompt label (e.g. "SEARCH", "SCANNING").
// query is the current search text.
// suffix is appended after the closing bracket (match counts, hints, etc.).
// accentStyle is applied to the brackets, cursor, and query text.
func SearchBar(label, query, suffix string, width int, accentStyle lipgloss.Style) string {
	prompt := accentStyle.Render("[ " + label + " > ")
	cursor := accentStyle.Render("█")
	q := accentStyle.Render(query)
	closeBracket := accentStyle.Render(" ]")
	line := prompt + q + cursor + closeBracket + suffix
	return Pad(line, width)
}

// Pad right-pads s with spaces to fill width.
func Pad(s string, width int) string {
	w := lipgloss.Width(s)
	if w < width {
		s += strings.Repeat(" ", width-w)
	}
	return s
}
