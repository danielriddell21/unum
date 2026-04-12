package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/danielriddell21/unum/internal/tui/panels"
)

var (
	colorAccent = "#00D4FF"
	colorDim    = "#3A3A3A"
	colorValue  = "#FFFFFF"
	colorBorder = "#1E1E1E"
	colorActive = "#00D4FF"
	colorBG     = "#0D0D0D"

	styleTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent)).
			Bold(true)

	styleTitleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDim))

	styleLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent))

	styleValue = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorValue))

	styleHint = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDim))

	styleHistorySelected = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorAccent)).
				Bold(true)

	borderActive = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorActive))

	borderDim = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorBorder))
)

// ApplyPalette updates styles to match the given palette.
func ApplyPalette(p panels.Palette) {
	colorAccent = p.AccentPrimary
	colorDim = p.Muted
	colorValue = p.Text
	colorBorder = p.BorderDim
	colorActive = p.BorderActive
	colorBG = p.BG

	styleTitle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Bold(true)
	styleTitleDim = lipgloss.NewStyle().Foreground(lipgloss.Color(colorDim))
	styleLabel = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent))
	styleValue = lipgloss.NewStyle().Foreground(lipgloss.Color(colorValue))
	styleHint = lipgloss.NewStyle().Foreground(lipgloss.Color(colorDim))
	styleHistorySelected = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Bold(true)
	borderActive = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(colorActive))
	borderDim = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(colorBorder))
}
