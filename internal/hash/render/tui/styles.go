package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorAccent = "#00D4FF"
	colorDim    = "#3A3A3A"
	colorValue  = "#E5E5E5"
	colorBorder = "#1A1A2E"
	colorActive = "#00D4FF"

	styleTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent)).
			Bold(true)

	styleLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent))

	styleValue = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorValue))

	styleDim = lipgloss.NewStyle().
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

// ApplyTheme updates styles to match a named theme.
func ApplyTheme(accent, dim, value, border string) {
	colorAccent = accent
	colorDim = dim
	colorValue = value
	colorBorder = border
	colorActive = accent

	styleTitle = lipgloss.NewStyle().Foreground(lipgloss.Color(accent)).Bold(true)
	styleLabel = lipgloss.NewStyle().Foreground(lipgloss.Color(accent))
	styleValue = lipgloss.NewStyle().Foreground(lipgloss.Color(value))
	styleDim = lipgloss.NewStyle().Foreground(lipgloss.Color(dim))
	styleHistorySelected = lipgloss.NewStyle().Foreground(lipgloss.Color(accent)).Bold(true)
	borderActive = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(accent))
	borderDim = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(border))
}
