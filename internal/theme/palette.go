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

const (
	cyberCyan   = "#00D4FF" // accent + border + object-key
	cyberPurple = "#C678DD" // accent + path + hash
	cyberGold   = "#FFD700" // accent + stats + search
	cyberRed    = "#E06C75" // bool-false + error + removed
)

// PaletteCyber is the default cyber/neural-interface theme.
var PaletteCyber = Palette{
	AccentPrimary:   cyberCyan,
	AccentSecondary: cyberPurple,
	AccentTertiary:  cyberGold,
	BG:              "#0D0D0D",
	BGPanel:         "#111111",
	BGSelected:      "#1A1A2E",
	BGAdded:         "#0d1a0d",
	BGRemoved:       "#1a0d0d",
	BorderActive:    cyberCyan,
	BorderDim:       "#1E1E1E",
	ObjectKey:       cyberCyan,
	ArrayIndex:      "#FF6B6B",
	StringVal:       "#98C379",
	NumberVal:       "#E5C07B",
	BoolTrue:        "#56B6C2",
	BoolFalse:       cyberRed,
	NullVal:         "#5C6370",
	Text:            "#C0C0C0",
	Muted:           "#3A3A3A",
	Error:           cyberRed,
	Path:            cyberPurple,
	Hash:            cyberPurple,
	Stats:           cyberGold,
	Search:          cyberGold,
	Added:           "#98C379",
	Removed:         cyberRed,
}

const (
	matrixGreen     = "#00FF41" // accent + border + object-key + bool-true + added
	matrixLime      = "#39FF14" // accent + path + hash
	matrixPaleGreen = "#88FF44" // accent + number + stats
	matrixRed       = "#FF3300" // bool-false + error + removed
)

// PaletteMatrix is the matrix green theme.
var PaletteMatrix = Palette{
	AccentPrimary:   matrixGreen,
	AccentSecondary: matrixLime,
	AccentTertiary:  matrixPaleGreen,
	BG:              "#0D0D0D",
	BGPanel:         "#0A1A0A",
	BGSelected:      "#001A00",
	BGAdded:         "#001800",
	BGRemoved:       "#180000",
	BorderActive:    matrixGreen,
	BorderDim:       "#003300",
	ObjectKey:       matrixGreen,
	ArrayIndex:      "#33FF33",
	StringVal:       "#00CC22",
	NumberVal:       matrixPaleGreen,
	BoolTrue:        matrixGreen,
	BoolFalse:       matrixRed,
	NullVal:         "#005500",
	Text:            "#00CC22",
	Muted:           "#005500",
	Error:           matrixRed,
	Path:            matrixLime,
	Hash:            matrixLime,
	Stats:           matrixPaleGreen,
	Search:          "#FFFFFF",
	Added:           matrixGreen,
	Removed:         matrixRed,
}

const (
	draculaPurple = "#BD93F9" // accent + border + object-key
	draculaPink   = "#FF79C6" // accent + path + hash
	draculaYellow = "#F1FA8C" // accent + number + stats + search
	draculaRed    = "#FF5555" // array-idx + bool-false + error + removed
)

// PaletteDracula is the Dracula theme.
var PaletteDracula = Palette{
	AccentPrimary:   draculaPurple,
	AccentSecondary: draculaPink,
	AccentTertiary:  draculaYellow,
	BG:              "#282A36",
	BGPanel:         "#21222C",
	BGSelected:      "#44475A",
	BGAdded:         "#1e3128",
	BGRemoved:       "#3a1a1e",
	BorderActive:    draculaPurple,
	BorderDim:       "#3D4050",
	ObjectKey:       draculaPurple,
	ArrayIndex:      draculaRed,
	StringVal:       "#50FA7B",
	NumberVal:       draculaYellow,
	BoolTrue:        "#8BE9FD",
	BoolFalse:       draculaRed,
	NullVal:         "#6272A4",
	Text:            "#F8F8F2",
	Muted:           "#6272A4",
	Error:           draculaRed,
	Path:            draculaPink,
	Hash:            draculaPink,
	Stats:           draculaYellow,
	Search:          draculaYellow,
	Added:           "#50FA7B",
	Removed:         draculaRed,
}

const (
	nordFrost  = "#88C0D0" // accent + border + object-key
	nordMauve  = "#B48EAD" // accent + path + hash
	nordYellow = "#EBCB8B" // accent + number + stats + search
	nordRed    = "#BF616A" // array-idx + bool-false + error + removed
)

// PaletteNord is the Nord theme.
var PaletteNord = Palette{
	AccentPrimary:   nordFrost,
	AccentSecondary: nordMauve,
	AccentTertiary:  nordYellow,
	BG:              "#2E3440",
	BGPanel:         "#272C36",
	BGSelected:      "#3B4252",
	BGAdded:         "#2c3a2c",
	BGRemoved:       "#3a2c2e",
	BorderActive:    nordFrost,
	BorderDim:       "#3B4252",
	ObjectKey:       nordFrost,
	ArrayIndex:      nordRed,
	StringVal:       "#A3BE8C",
	NumberVal:       nordYellow,
	BoolTrue:        "#81A1C1",
	BoolFalse:       nordRed,
	NullVal:         "#4C566A",
	Text:            "#ECEFF4",
	Muted:           "#4C566A",
	Error:           nordRed,
	Path:            nordMauve,
	Hash:            nordMauve,
	Stats:           nordYellow,
	Search:          nordYellow,
	Added:           "#A3BE8C",
	Removed:         nordRed,
}

const (
	cleanBlue   = "#007ACC" // accent + border + object-key
	cleanPurple = "#8E44AD" // accent + path + hash
	cleanOrange = "#E67E22" // accent + stats + search
	cleanRed    = "#C0392B" // array-idx + bool-false + error + removed
)

// PaletteClean is the default light theme.
var PaletteClean = Palette{
	AccentPrimary:   cleanBlue,
	AccentSecondary: cleanPurple,
	AccentTertiary:  cleanOrange,
	BG:              "#F5F7FA",
	BGPanel:         "#EAECF0",
	BGSelected:      "#DDE3EE",
	BGAdded:         "#E8F5E9",
	BGRemoved:       "#FFEBEE",
	BorderActive:    cleanBlue,
	BorderDim:       "#C8D0DC",
	ObjectKey:       cleanBlue,
	ArrayIndex:      cleanRed,
	StringVal:       "#27882B",
	NumberVal:       "#B07D00",
	BoolTrue:        "#2980B9",
	BoolFalse:       cleanRed,
	NullVal:         "#7F8C8D",
	Text:            "#1A2332",
	Muted:           "#95A5A6",
	Error:           cleanRed,
	Path:            cleanPurple,
	Hash:            cleanPurple,
	Stats:           cleanOrange,
	Search:          cleanOrange,
	Added:           "#27882B",
	Removed:         cleanRed,
}

const (
	solarBlue    = "#268BD2" // accent + border + object-key
	solarMagenta = "#D33682" // accent + path + hash
	solarOrange  = "#CB4B16" // accent + stats + search
	solarRed     = "#DC322F" // array-idx + bool-false + error + removed
)

// PaletteSolarized is the Solarized light theme.
var PaletteSolarized = Palette{
	AccentPrimary:   solarBlue,
	AccentSecondary: solarMagenta,
	AccentTertiary:  solarOrange,
	BG:              "#FDF6E3",
	BGPanel:         "#EEE8D5",
	BGSelected:      "#E0D9C5",
	BGAdded:         "#EBF5EB",
	BGRemoved:       "#FBE9E9",
	BorderActive:    solarBlue,
	BorderDim:       "#D0C9B5",
	ObjectKey:       solarBlue,
	ArrayIndex:      solarRed,
	StringVal:       "#859900",
	NumberVal:       "#B58900",
	BoolTrue:        "#2AA198",
	BoolFalse:       solarRed,
	NullVal:         "#93A1A1",
	Text:            "#657B83",
	Muted:           "#93A1A1",
	Error:           solarRed,
	Path:            solarMagenta,
	Hash:            solarMagenta,
	Stats:           solarOrange,
	Search:          solarOrange,
	Added:           "#859900",
	Removed:         solarRed,
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
