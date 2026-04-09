package tui

import "github.com/charmbracelet/lipgloss"

// palette — cyber/neural-interface color scheme
const (
	colorBG          = "#0D0D0D"
	colorActiveBorder = "#00D4FF" // cyan — focused panel border
	colorDimBorder   = "#1E1E1E" // very dim — unfocused panel border
	colorObjectKey   = "#00D4FF" // cyan — object keys
	colorArrayIndex  = "#FF6B6B" // coral — array indices
	colorString      = "#98C379" // soft green — string values
	colorNumber      = "#E5C07B" // amber — numbers
	colorBoolTrue    = "#56B6C2" // teal — true
	colorBoolFalse   = "#E06C75" // red — false
	colorNull        = "#5C6370" // muted gray — null
	colorSelected    = "#1A1A2E" // deep navy — selected row bg
	colorSelectedFG  = "#FFFFFF"
	colorPath        = "#C678DD" // purple — jq path in statusbar
	colorSearch      = "#FFD700" // gold — search highlight
	colorHash        = "#C678DD" // purple — merkle hashes
	colorStats       = "#FFD700" // gold — stat annotations
	colorTabActive   = "#00D4FF"
	colorTabInactive = "#3A3A3A"
	colorMuted       = "#3A3A3A"
	colorError       = "#E06C75"
	colorGreen       = "#00FF41" // matrix green — boot/status
)

var (
	// Panel titles
	styleTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorActiveBorder)).
			Bold(true)

	styleTitleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted))

	// Panel borders — active vs inactive
	borderActive = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorActiveBorder))

	borderDim = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorDimBorder))
)
