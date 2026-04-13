package tui

import (
	"github.com/charmbracelet/lipgloss"

	jsonpanels "github.com/danielriddell21/unum/internal/json/render/tui/panels"
	"github.com/danielriddell21/unum/internal/tui/panels"
)

var (
	// Panel titles
	styleTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(panels.PaletteCyber.AccentPrimary)).
			Bold(true)

	styleTitleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color(panels.PaletteCyber.Muted))

	// Panel borders — active vs inactive
	borderActive = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(panels.PaletteCyber.BorderActive))

	borderDim = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(panels.PaletteCyber.BorderDim))

	// Muted hint text — used by picker and other non-panel chrome
	styleHint = lipgloss.NewStyle().
			Foreground(lipgloss.Color(panels.PaletteCyber.Muted))

	// colorActiveBorder is kept for the help overlay border reference
	colorActiveBorder = panels.PaletteCyber.BorderActive

	colorBG = panels.PaletteCyber.BG
)

// ApplyPalette rebuilds all style vars to match the given palette.
// Call this before running any TUI program.
func ApplyPalette(p panels.Palette) {
	jsonpanels.ApplyPalette(p)
	styleTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.AccentPrimary)).
		Bold(true)
	styleTitleDim = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Muted))
	borderActive = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(p.BorderActive))
	borderDim = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(p.BorderDim))
	styleHint = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Muted))
	colorActiveBorder = p.BorderActive
	colorBG = p.BG
}
