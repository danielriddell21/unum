package tui

import (
	"github.com/danielriddell21/unum/internal/image/optimize"
	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

func statusBar(m Model, width int) string {
	left := " " + tuipanels.StyleTitle.Render("[ "+m.src.Name+" ]")
	if m.result != nil {
		left += "  " + tuipanels.StyleHint.Render(
			optimize.HumanBytes(m.src.Bytes)+" → "+optimize.HumanBytes(m.result.Bytes))
	}

	right := tuipanels.StyleHint.Render("←→:quality  ↑↓:scale  f:format  s:save  ?:help  q:quit") +
		"  " + tuipanels.StyleHint.Render(m.version) + " "

	return tuipanels.Bar(left, right, width)
}
