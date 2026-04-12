package tui

import (
	"github.com/danielriddell21/unum/internal/hash/types"
	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

func statusBar(result *types.Result, focused focus, width int) string {
	var left string
	if result != nil {
		left = " " + styleTitle.Render("[ "+result.Input+" ]")
	} else {
		left = " " + styleHint.Render("no input yet")
	}

	var right string
	if focused == focusInput {
		right = styleHint.Render("enter:derive  tab:history  ?:help  esc:quit") + " "
	} else {
		right = styleHint.Render("↑↓:navigate  enter:re-derive  tab:back  ?:help  q:quit") + " "
	}

	return tuipanels.Bar(left, right, width)
}
