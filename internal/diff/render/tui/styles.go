package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/danielriddell21/unum/internal/tui/panels"
)

var (
	styleTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(panels.PaletteCyber.AccentPrimary)).
			Bold(true)

	styleTitleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color(panels.PaletteCyber.Muted))

	styleHint = lipgloss.NewStyle().
			Foreground(lipgloss.Color(panels.PaletteCyber.Muted))

	borderActive = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(panels.PaletteCyber.BorderActive))

	borderDim = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(panels.PaletteCyber.BorderDim))

	colorActiveBorder = panels.PaletteCyber.BorderActive

	colorBG = panels.PaletteCyber.BG
)

// ApplyPalette rebuilds all tui-level style vars from the palette.
func ApplyPalette(p panels.Palette) {
	styleTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.AccentPrimary)).Bold(true)
	styleTitleDim = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Muted))
	styleHint = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Muted))
	borderActive = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(p.BorderActive))
	borderDim = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(p.BorderDim))
	colorActiveBorder = p.BorderActive
	colorBG = p.BG
}
