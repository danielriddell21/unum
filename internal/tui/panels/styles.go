package panels

import "github.com/charmbracelet/lipgloss"

// Shared TUI style variables — initialised from PaletteCyber, rebuilt by ApplyBaseStyles.
var (
	StyleTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(PaletteCyber.AccentPrimary)).
			Bold(true)

	StyleTitleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color(PaletteCyber.Muted))

	StyleHint = lipgloss.NewStyle().
			Foreground(lipgloss.Color(PaletteCyber.Muted))

	BorderActive = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(PaletteCyber.BorderActive))

	BorderDim = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(PaletteCyber.BorderDim))

	ColorActiveBorder = PaletteCyber.BorderActive

	ColorBG = PaletteCyber.BG
)

// ApplyBaseStyles rebuilds all shared style vars from the given palette.
func ApplyBaseStyles(p Palette) {
	StyleTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.AccentPrimary)).Bold(true)
	StyleTitleDim = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Muted))
	StyleHint = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Muted))
	BorderActive = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(p.BorderActive))
	BorderDim = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(p.BorderDim))
	ColorActiveBorder = p.BorderActive
	ColorBG = p.BG
}

// PanelTitle renders a panel header, highlighted when active.
func PanelTitle(title string, active bool) string {
	if active {
		return StyleTitle.Render(" " + title + " ")
	}
	return StyleTitleDim.Render(" " + title + " ")
}

// WrapPanel wraps content in a rounded border, highlighted when active.
func WrapPanel(content string, active bool, width, height int) string {
	if width < 4 {
		width = 4
	}
	if height < 4 {
		height = 4
	}
	style := BorderDim.Width(width - 2).Height(height - 2)
	if active {
		style = BorderActive.Width(width - 2).Height(height - 2)
	}
	return style.Render(content)
}
