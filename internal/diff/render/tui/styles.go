package tui

import (
	diffpanels "github.com/danielriddell21/unum/internal/diff/render/tui/panels"
	"github.com/danielriddell21/unum/internal/tui/panels"
)

// ApplyPalette rebuilds all tui-level and panel-level styles from the palette.
func ApplyPalette(p panels.Palette) {
	panels.ApplyBaseStyles(p)
	diffpanels.ApplyPalette(p)
}
