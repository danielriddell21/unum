package panels

import (
	"github.com/charmbracelet/lipgloss"

	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

// ApplyPalette updates all panel style variables to match the given palette.
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

// Package-level style variables — initialized with the cyber palette defaults.
// ApplyPalette() reassigns all of these; do not set them elsewhere.
var (
	styleMuted      = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Muted))
	styleObjectKey  = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.ObjectKey))
	styleArrayIdx   = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.ArrayIndex))
	styleString     = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.StringVal))
	styleNumber     = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.NumberVal))
	styleBoolTrue   = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.BoolTrue))
	styleBoolFalse  = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.BoolFalse))
	styleNull       = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.NullVal))
	styleCursor     = lipgloss.NewStyle().Background(lipgloss.Color(tuipanels.PaletteCyber.BGSelected)).Foreground(lipgloss.Color(tuipanels.PaletteCyber.Text))
	connectorMid    = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Muted)).Render("├─▶ ")
	connectorLast   = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.BorderDim)).Render("└─▷ ")
	connectorOpen   = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Muted)).Render("▼  ")
	connectorClosed = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Muted)).Render("▶  ")

	styleTabActive    = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.AccentPrimary)).Bold(true)
	styleTabInactive  = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Muted))
	styleHash         = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Hash))
	styleStats        = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Stats))
	styleError        = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Error))
	styleSearchPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Search)).Bold(true)
	stylePath         = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.Path)).Bold(true)
	styleSBType       = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.AccentPrimary))
	styleSBSep        = lipgloss.NewStyle().Foreground(lipgloss.Color(tuipanels.PaletteCyber.BorderDim)).Render(" · ")
)
