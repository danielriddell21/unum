// Package theme holds the colour palette shared by all render modes (TUI, static, web).
package theme

// Palette holds all themeable hex color strings.
type Palette struct {
	AccentPrimary   string // active border, tab active, status type
	AccentSecondary string // path, hash
	AccentTertiary  string // stats, search highlight
	BG              string // terminal background
	BGPanel         string // panel/sidebar background (slightly different from BG)
	BGSelected      string // cursor row background (web: --bg-hover)
	BGAdded         string // faint added-row background (web diff)
	BGRemoved       string // faint removed-row background (web diff)
	BorderActive    string // focused panel border
	BorderDim       string // unfocused panel border and dim connectors
	ObjectKey       string
	ArrayIndex      string
	StringVal       string
	NumberVal       string
	BoolTrue        string
	BoolFalse       string
	NullVal         string
	Text            string // primary text colour
	Muted           string // muted / inactive text
	Error           string
	Path            string
	Hash            string
	Stats           string
	Search          string
	Added           string // diff line added
	Removed         string // diff line removed
}

// ToCSSVars returns the palette as a map of CSS custom property names to values.
// The map covers every variable used by shared.js and the web tool stylesheets.
func (p Palette) ToCSSVars() map[string]string {
	return map[string]string{
		"--bg":            p.BG,
		"--bg-panel":      p.BGPanel,
		"--bg-hover":      p.BGSelected,
		"--border":        p.BorderDim,
		"--border-active": p.BorderActive,
		"--text":          p.Text,
		"--muted":         p.Muted,
		"--array-idx":     p.ArrayIndex,
		"--string":        p.StringVal,
		"--number":        p.NumberVal,
		"--bool-true":     p.BoolTrue,
		"--bool-false":    p.BoolFalse,
		"--null":          p.NullVal,
		"--path":          p.Path,
		"--search":        p.Search,
		"--hash":          p.Hash,
		"--stats":         p.Stats,
		"--bg-added":      p.BGAdded,
		"--bg-removed":    p.BGRemoved,
		"--diff-added":    p.Added,
		"--diff-removed":  p.Removed,
		"--diff-hunk":     p.AccentPrimary,
	}
}

// PaletteCyber is the default cyber/neural-interface theme.
var PaletteCyber = Palette{
	AccentPrimary:   "#00D4FF",
	AccentSecondary: "#C678DD",
	AccentTertiary:  "#FFD700",
	BG:              "#0D0D0D",
	BGPanel:         "#111111",
	BGSelected:      "#1A1A2E",
	BGAdded:         "#0d1a0d",
	BGRemoved:       "#1a0d0d",
	BorderActive:    "#00D4FF",
	BorderDim:       "#1E1E1E",
	ObjectKey:       "#00D4FF",
	ArrayIndex:      "#FF6B6B",
	StringVal:       "#98C379",
	NumberVal:       "#E5C07B",
	BoolTrue:        "#56B6C2",
	BoolFalse:       "#E06C75",
	NullVal:         "#5C6370",
	Text:            "#C0C0C0",
	Muted:           "#3A3A3A",
	Error:           "#E06C75",
	Path:            "#C678DD",
	Hash:            "#C678DD",
	Stats:           "#FFD700",
	Search:          "#FFD700",
	Added:           "#98C379",
	Removed:         "#E06C75",
}

// PaletteMatrix is the matrix green theme.
var PaletteMatrix = Palette{
	AccentPrimary:   "#00FF41",
	AccentSecondary: "#39FF14",
	AccentTertiary:  "#88FF44",
	BG:              "#0D0D0D",
	BGPanel:         "#0A1A0A",
	BGSelected:      "#001A00",
	BGAdded:         "#001800",
	BGRemoved:       "#180000",
	BorderActive:    "#00FF41",
	BorderDim:       "#003300",
	ObjectKey:       "#00FF41",
	ArrayIndex:      "#33FF33",
	StringVal:       "#00CC22",
	NumberVal:       "#88FF44",
	BoolTrue:        "#00FF41",
	BoolFalse:       "#FF3300",
	NullVal:         "#005500",
	Text:            "#00CC22",
	Muted:           "#005500",
	Error:           "#FF3300",
	Path:            "#39FF14",
	Hash:            "#39FF14",
	Stats:           "#88FF44",
	Search:          "#FFFFFF",
	Added:           "#00FF41",
	Removed:         "#FF3300",
}

// PaletteDracula is the Dracula theme.
var PaletteDracula = Palette{
	AccentPrimary:   "#BD93F9",
	AccentSecondary: "#FF79C6",
	AccentTertiary:  "#F1FA8C",
	BG:              "#282A36",
	BGPanel:         "#21222C",
	BGSelected:      "#44475A",
	BGAdded:         "#1e3128",
	BGRemoved:       "#3a1a1e",
	BorderActive:    "#BD93F9",
	BorderDim:       "#3D4050",
	ObjectKey:       "#BD93F9",
	ArrayIndex:      "#FF5555",
	StringVal:       "#50FA7B",
	NumberVal:       "#F1FA8C",
	BoolTrue:        "#8BE9FD",
	BoolFalse:       "#FF5555",
	NullVal:         "#6272A4",
	Text:            "#F8F8F2",
	Muted:           "#6272A4",
	Error:           "#FF5555",
	Path:            "#FF79C6",
	Hash:            "#FF79C6",
	Stats:           "#F1FA8C",
	Search:          "#F1FA8C",
	Added:           "#50FA7B",
	Removed:         "#FF5555",
}

// PaletteNord is the Nord theme.
var PaletteNord = Palette{
	AccentPrimary:   "#88C0D0",
	AccentSecondary: "#B48EAD",
	AccentTertiary:  "#EBCB8B",
	BG:              "#2E3440",
	BGPanel:         "#272C36",
	BGSelected:      "#3B4252",
	BGAdded:         "#2c3a2c",
	BGRemoved:       "#3a2c2e",
	BorderActive:    "#88C0D0",
	BorderDim:       "#3B4252",
	ObjectKey:       "#88C0D0",
	ArrayIndex:      "#BF616A",
	StringVal:       "#A3BE8C",
	NumberVal:       "#EBCB8B",
	BoolTrue:        "#81A1C1",
	BoolFalse:       "#BF616A",
	NullVal:         "#4C566A",
	Text:            "#ECEFF4",
	Muted:           "#4C566A",
	Error:           "#BF616A",
	Path:            "#B48EAD",
	Hash:            "#B48EAD",
	Stats:           "#EBCB8B",
	Search:          "#EBCB8B",
	Added:           "#A3BE8C",
	Removed:         "#BF616A",
}

// PaletteClean is the default light theme.
var PaletteClean = Palette{
	AccentPrimary:   "#007ACC",
	AccentSecondary: "#8E44AD",
	AccentTertiary:  "#E67E22",
	BG:              "#F5F7FA",
	BGPanel:         "#EAECF0",
	BGSelected:      "#DDE3EE",
	BGAdded:         "#E8F5E9",
	BGRemoved:       "#FFEBEE",
	BorderActive:    "#007ACC",
	BorderDim:       "#C8D0DC",
	ObjectKey:       "#007ACC",
	ArrayIndex:      "#C0392B",
	StringVal:       "#27882B",
	NumberVal:       "#B07D00",
	BoolTrue:        "#2980B9",
	BoolFalse:       "#C0392B",
	NullVal:         "#7F8C8D",
	Text:            "#1A2332",
	Muted:           "#95A5A6",
	Error:           "#C0392B",
	Path:            "#8E44AD",
	Hash:            "#8E44AD",
	Stats:           "#E67E22",
	Search:          "#E67E22",
	Added:           "#27882B",
	Removed:         "#C0392B",
}

// PaletteSolarized is the Solarized light theme.
var PaletteSolarized = Palette{
	AccentPrimary:   "#268BD2",
	AccentSecondary: "#D33682",
	AccentTertiary:  "#CB4B16",
	BG:              "#FDF6E3",
	BGPanel:         "#EEE8D5",
	BGSelected:      "#E0D9C5",
	BGAdded:         "#EBF5EB",
	BGRemoved:       "#FBE9E9",
	BorderActive:    "#268BD2",
	BorderDim:       "#D0C9B5",
	ObjectKey:       "#268BD2",
	ArrayIndex:      "#DC322F",
	StringVal:       "#859900",
	NumberVal:       "#B58900",
	BoolTrue:        "#2AA198",
	BoolFalse:       "#DC322F",
	NullVal:         "#93A1A1",
	Text:            "#657B83",
	Muted:           "#93A1A1",
	Error:           "#DC322F",
	Path:            "#D33682",
	Hash:            "#D33682",
	Stats:           "#CB4B16",
	Search:          "#CB4B16",
	Added:           "#859900",
	Removed:         "#DC322F",
}

// ResolvePalette returns the named dark palette, defaulting to Cyber.
func ResolvePalette(name string) Palette {
	switch name {
	case "matrix":
		return PaletteMatrix
	case "dracula":
		return PaletteDracula
	case "nord":
		return PaletteNord
	default:
		return PaletteCyber
	}
}

// ResolveLightPalette returns the named light palette, defaulting to Clean.
func ResolveLightPalette(name string) Palette {
	switch name {
	case "solarized":
		return PaletteSolarized
	default:
		return PaletteClean
	}
}
