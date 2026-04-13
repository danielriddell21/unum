package tui

import (
	"github.com/danielriddell21/unum/internal/hash/types"
	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

func statusBar(result *types.Result, focused focus, width int, version string) string {
	var left string
	if result != nil {
		left = " " + tuipanels.StyleTitle.Render("[ "+result.Input+" ]")
	} else {
		left = " " + tuipanels.StyleHint.Render("no input yet")
	}

	var right string
	if focused == focusInput {
		right = tuipanels.StyleHint.Render("enter:derive  tab:history  ?:help  esc:quit") + "  " + tuipanels.StyleHint.Render(version) + " "
	} else {
		right = tuipanels.StyleHint.Render("↑↓:navigate  enter:re-derive  tab:back  ?:help  q:quit") + "  " + tuipanels.StyleHint.Render(version) + " "
	}

	return tuipanels.Bar(left, right, width)
}
