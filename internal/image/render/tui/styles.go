package tui

import (
	imagepanels "github.com/danielriddell21/unum/internal/image/render/tui/panels"
	"github.com/danielriddell21/unum/internal/tui/panels"
)

func ApplyPalette(p panels.Palette) {
	panels.ApplyBaseStyles(p)
	imagepanels.ApplyPalette(p)
}
