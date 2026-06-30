package panels

import (
	"fmt"
	"strings"

	"github.com/danielriddell21/unum/internal/hash/types"
)

type HashPanel struct {
	width  int
	height int
}

func NewHashPanel(w, h int) HashPanel {
	return HashPanel{width: w, height: h}
}

func (p *HashPanel) Resize(w, h int) {
	p.width = w
	p.height = h
}

func (p *HashPanel) View(inputView string, result *types.Result) string {
	innerW := p.width
	if innerW < 1 {
		innerW = 1
	}
	sep := styleHint.Render(strings.Repeat("─", innerW))

	var resultsBlock string
	if result != nil {
		rows := []struct{ label, value string }{
			{"input", result.Input},
			{"port", fmt.Sprintf("%d", result.Port)},
			{"uuid", result.UUID},
			{"color", result.Color},
			{"short", result.Short},
			{"emoji", result.Emoji},
			{"phrase", result.Phrase},
		}
		lines := make([]string, 0, len(rows)+2)
		lines = append(lines, sep)
		for _, row := range rows {
			label := styleLabel.Render(fmt.Sprintf("  %-7s", row.label))
			value := styleValue.Render(row.value)
			lines = append(lines, fmt.Sprintf("%s  %s", label, value))
		}
		lines = append(lines, sep)
		resultsBlock = strings.Join(lines, "\n")
	} else {
		resultsBlock = styleHint.Render("  type something and press enter")
	}

	return strings.Join([]string{"", inputView, "", resultsBlock}, "\n")
}
