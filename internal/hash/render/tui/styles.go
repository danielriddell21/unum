package tui

import (
	hashpanels "github.com/danielriddell21/unum/internal/hash/render/tui/panels"
	"github.com/danielriddell21/unum/internal/tui/panels"
)

func ApplyPalette(p panels.Palette) {
	panels.ApplyBaseStyles(p)
	hashpanels.ApplyPalette(p)
}
