package panels

import (
	"github.com/charmbracelet/lipgloss"

	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

// ApplyPalette updates all panel style variables to match the given palette.
func ApplyPalette(p tuipanels.Palette) {
	diffAdded = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Added))
	diffRemoved = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Removed))
	diffUnchanged = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	diffHunkHdr = lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary))
	diffLineNum = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	diffSearchHL = lipgloss.NewStyle().Background(lipgloss.Color(p.Search)).Foreground(lipgloss.Color("#000000"))
	sbAdded = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Added)).Bold(true)
	sbRemoved = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Removed)).Bold(true)
	sbMuted = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	sbAccent = lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary))
}

// Package-level style variables — initialized with the cyber palette defaults.
// ApplyPalette() reassigns all of these; do not set them elsewhere.
var (
	diffAdded     = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Added))
	diffRemoved   = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Removed))
	diffUnchanged = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Muted))
	diffHunkHdr   = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.AccentPrimary))
	diffLineNum   = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Muted))
	diffSearchHL  = lipgloss.NewStyle().Background(lipgloss.Color(tuipanels.PaletteCyber.Search)).Foreground(lipgloss.Color("#000000"))

	sbAdded   = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Added)).Bold(true)
	sbRemoved = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Removed)).Bold(true)
	sbMuted   = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Muted))
	sbAccent  = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.AccentPrimary))
)
