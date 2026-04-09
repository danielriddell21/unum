package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/danielriddell21/unum/internal/diff/node"
	jsonpanels "github.com/danielriddell21/unum/internal/json/render/tui/panels"
	diffpanels "github.com/danielriddell21/unum/internal/diff/render/tui/panels"
)

// Start launches the diff TUI.
func Start(d *node.Diff, theme string) error {
	applyTheme(theme)
	m := NewModel(d)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func applyTheme(theme string) {
	p := jsonpanels.ResolvePalette(theme)
	ApplyPalette(p)
	diffpanels.ApplyPalette(p)
	diffpanels.ApplyDiffPalette(p)
}
