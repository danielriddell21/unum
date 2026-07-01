package panels

import (
	"github.com/charmbracelet/lipgloss"

	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

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

func init() { ApplyPalette(tuipanels.PaletteCyber) }

var (
	diffAdded     lipgloss.Style
	diffRemoved   lipgloss.Style
	diffUnchanged lipgloss.Style
	diffHunkHdr   lipgloss.Style
	diffLineNum   lipgloss.Style
	diffSearchHL  lipgloss.Style

	sbAdded   lipgloss.Style
	sbRemoved lipgloss.Style
	sbMuted   lipgloss.Style
	sbAccent  lipgloss.Style
)
