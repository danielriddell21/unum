package panels

import (
	"github.com/charmbracelet/lipgloss"

	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

func ApplyPalette(p tuipanels.Palette) {
	styleMuted = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	styleObjectKey = lipgloss.NewStyle().Foreground(lipgloss.Color(p.ObjectKey))
	styleArrayIdx = lipgloss.NewStyle().Foreground(lipgloss.Color(p.ArrayIndex))
	styleString = lipgloss.NewStyle().Foreground(lipgloss.Color(p.StringVal))
	styleNumber = lipgloss.NewStyle().Foreground(lipgloss.Color(p.NumberVal))
	styleBoolTrue = lipgloss.NewStyle().Foreground(lipgloss.Color(p.BoolTrue))
	styleBoolFalse = lipgloss.NewStyle().Foreground(lipgloss.Color(p.BoolFalse))
	styleNull = lipgloss.NewStyle().Foreground(lipgloss.Color(p.NullVal))
	styleCursor = lipgloss.NewStyle().
		Background(lipgloss.Color(p.BGSelected)).
		Foreground(lipgloss.Color(p.Text))
	connectorMid = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render("├─▶ ")
	connectorLast = lipgloss.NewStyle().Foreground(lipgloss.Color(p.BorderDim)).Render("└─▷ ")
	connectorOpen = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render("▼  ")
	connectorClosed = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render("▶  ")
	styleTabActive = lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary)).Bold(true)
	styleTabInactive = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	styleHash = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Hash))
	styleStats = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Stats))
	styleError = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Error))
	styleSearchPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Search)).Bold(true)
	stylePath = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Path)).Bold(true)
	styleSBType = lipgloss.NewStyle().Foreground(lipgloss.Color(p.AccentPrimary))
	styleSBSep = lipgloss.NewStyle().Foreground(lipgloss.Color(p.BorderDim)).Render(" · ")
}

func init() { ApplyPalette(tuipanels.PaletteCyber) }

var (
	styleMuted      lipgloss.Style
	styleObjectKey  lipgloss.Style
	styleArrayIdx   lipgloss.Style
	styleString     lipgloss.Style
	styleNumber     lipgloss.Style
	styleBoolTrue   lipgloss.Style
	styleBoolFalse  lipgloss.Style
	styleNull       lipgloss.Style
	styleCursor     lipgloss.Style
	connectorMid    string
	connectorLast   string
	connectorOpen   string
	connectorClosed string

	styleTabActive    lipgloss.Style
	styleTabInactive  lipgloss.Style
	styleHash         lipgloss.Style
	styleStats        lipgloss.Style
	styleError        lipgloss.Style
	styleSearchPrompt lipgloss.Style
	stylePath         lipgloss.Style
	styleSBType       lipgloss.Style
	styleSBSep        string
)
