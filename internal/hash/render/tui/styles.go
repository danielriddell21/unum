package tui

import (
	"github.com/charmbracelet/lipgloss"

	hashpanels "github.com/danielriddell21/unum/internal/hash/render/tui/panels"
	"github.com/danielriddell21/unum/internal/tui/panels"
)

var (
	colorActive = panels.PaletteCyber.BorderActive
	colorBG     = panels.PaletteCyber.BG

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
)

// ApplyPalette updates all style variables to match the given palette.
func ApplyPalette(p panels.Palette) {
	hashpanels.ApplyPalette(p)
	colorActive = p.BorderActive
	colorBG = p.BG
	styleTitle = lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary)).Bold(true)
	styleTitleDim = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	styleHint = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	borderActive = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(p.BorderActive))
	borderDim = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(p.BorderDim))
}
