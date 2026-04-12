package panels

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpOverlay centers helpContent over base as a rounded-border modal.
func HelpOverlay(base, helpContent, borderColor string, width int) string {
	help := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(1, 2).
		Width(width).
		Render(helpContent)

	w := lipgloss.Width(base)
	h := lipgloss.Height(base)
	hw := lipgloss.Width(help)
	hh := lipgloss.Height(help)
	x := (w - hw) / 2
	y := (h - hh) / 2

	lines := strings.Split(base, "\n")
	helpLines := strings.Split(help, "\n")
	for i, hl := range helpLines {
		row := y + i
		if row >= 0 && row < len(lines) {
			line := lines[row]
			if x >= 0 && x < lipgloss.Width(line) {
				lines[row] = line[:x] + hl
			}
		}
	}
	return strings.Join(lines, "\n")
}
