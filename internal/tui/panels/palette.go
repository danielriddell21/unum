package panels

// Re-export from internal/theme for backwards compatibility within TUI code.
// New code should import internal/theme directly.

import "github.com/danielriddell21/unum/internal/theme"

// Palette is an alias for theme.Palette.
type Palette = theme.Palette

var (
	PaletteCyber   = theme.PaletteCyber
	PaletteMatrix  = theme.PaletteMatrix
	PaletteDracula = theme.PaletteDracula
	PaletteNord    = theme.PaletteNord
)

// ResolvePalette delegates to theme.ResolvePalette.
func ResolvePalette(name string) Palette {
	return theme.ResolvePalette(name)
}
