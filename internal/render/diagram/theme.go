package diagram

import (
	"oss.terrastruct.com/d2/d2target"

	"github.com/danielriddell21/unum/internal/theme"
)

func ThemeByName(name string) *Theme {
	switch name {
	case "clean", "solarized":
		return themeFromPalette(theme.ResolveLightPalette(name), false)
	default:
		return themeFromPalette(theme.ResolvePalette(name), true)
	}
}

func themeFromPalette(p theme.Palette, dark bool) *Theme {
	return &Theme{
		Dark:       dark,
		Background: p.BG,
		Surface:    p.BGPanel,
		SurfaceAlt: p.BGSelected,
		Border:     p.BorderDim,
		Text:       p.Text,
		Muted:      p.Muted,
		Accent:     p.AccentPrimary,
		AccentAlt:  p.AccentSecondary,
	}
}

type Theme struct {
	Dark       bool
	Background string
	Surface    string
	SurfaceAlt string
	Border     string
	Text       string
	Muted      string
	Accent     string
	AccentAlt  string
}

func p(s string) *string { return &s }

func (t *Theme) d2Base() int64 {
	if t.Dark {
		return 200
	}
	return 0
}

func (t *Theme) d2Overrides() *d2target.ThemeOverrides {
	return &d2target.ThemeOverrides{
		N1: p(t.Text), N2: p(t.Text), N3: p(t.Muted), N4: p(t.Muted),
		N5: p(t.Border), N6: p(t.Surface), N7: p(t.Background),
		B1: p(t.Accent), B2: p(t.Accent), B3: p(t.SurfaceAlt),
		B4: p(t.SurfaceAlt), B5: p(t.Surface), B6: p(t.Background),
		AA2: p(t.AccentAlt), AA4: p(t.SurfaceAlt), AA5: p(t.Surface),
		AB4: p(t.SurfaceAlt), AB5: p(t.Surface),
	}
}
