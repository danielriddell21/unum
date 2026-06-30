package panels

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func Bar(left, right string, width int) string {
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	mid := width - leftW - rightW
	if mid < 1 {
		mid = 1
	}
	return left + strings.Repeat(" ", mid) + right
}

func SearchBar(label, query, suffix string, width int, accentStyle lipgloss.Style) string {
	prompt := accentStyle.Render("[ " + label + " > ")
	cursor := accentStyle.Render("█")
	q := accentStyle.Render(query)
	closeBracket := accentStyle.Render(" ]")
	line := prompt + q + cursor + closeBracket + suffix
	return Pad(line, width)
}

func Pad(s string, width int) string {
	w := lipgloss.Width(s)
	if w < width {
		s += strings.Repeat(" ", width-w)
	}
	return s
}
