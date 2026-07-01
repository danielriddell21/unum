package tui

import (
	jsonpanels "github.com/danielriddell21/unum/internal/json/render/tui/panels"
	"github.com/danielriddell21/unum/internal/tui/panels"
)

func ApplyPalette(p panels.Palette) {
	panels.ApplyBaseStyles(p)
	jsonpanels.ApplyPalette(p)
}
