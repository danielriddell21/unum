package panels

import "github.com/danielriddell21/unum/internal/theme"

type Palette = theme.Palette

var (
	PaletteCyber   = theme.PaletteCyber
	PaletteMatrix  = theme.PaletteMatrix
	PaletteDracula = theme.PaletteDracula
	PaletteNord    = theme.PaletteNord
)

func ResolvePalette(name string) Palette {
	return theme.ResolvePalette(name)
}
