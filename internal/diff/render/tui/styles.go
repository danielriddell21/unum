package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/danielriddell21/unum/internal/json/render/tui/panels"
)

var (
	styleTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(panels.PaletteCyber.AccentPrimary)).
			Bold(true)

	styleTitleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color(panels.PaletteCyber.Muted))

	borderActive = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(panels.PaletteCyber.BorderActive))

	borderDim = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(panels.PaletteCyber.BorderDim))

	styleHint = lipgloss.NewStyle().
			Foreground(lipgloss.Color(panels.PaletteCyber.Muted))

	styleAdded = lipgloss.NewStyle().
			Foreground(lipgloss.Color(panels.PaletteCyber.Added))

	styleRemoved = lipgloss.NewStyle().
			Foreground(lipgloss.Color(panels.PaletteCyber.Removed))

	styleHunkHeader = lipgloss.NewStyle().
			Foreground(lipgloss.Color(panels.PaletteCyber.AccentPrimary))

	styleLineNum = lipgloss.NewStyle().
			Foreground(lipgloss.Color(panels.PaletteCyber.Muted))

	colorActiveBorder = panels.PaletteCyber.BorderActive
)

// ApplyPalette rebuilds all tui-level style vars from the palette.
func ApplyPalette(p panels.Palette) {
	styleTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.AccentPrimary)).Bold(true)
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
	styleAdded = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Added))
	styleRemoved = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Removed))
	styleHunkHeader = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.AccentPrimary))
	styleLineNum = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Muted))
	colorActiveBorder = p.BorderActive
}
