package panels

import (
	"github.com/charmbracelet/lipgloss"

	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

func ApplyPalette(p tuipanels.Palette) {
	styleLabel = lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary))
	styleValue = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Text))
	styleHint = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	styleHistorySelected = lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary)).Bold(true)
}

var (
	styleLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color(tuipanels.PaletteCyber.AccentPrimary))

	styleValue = lipgloss.NewStyle().
			Foreground(lipgloss.Color(tuipanels.PaletteCyber.Text))

	styleHint = lipgloss.NewStyle().
			Foreground(lipgloss.Color(tuipanels.PaletteCyber.Muted))

	styleHistorySelected = lipgloss.NewStyle().
				Foreground(lipgloss.Color(tuipanels.PaletteCyber.AccentPrimary)).
				Bold(true)
)
