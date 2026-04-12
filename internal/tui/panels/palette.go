package panels

// Palette holds all themeable hex color strings for the TUI.
type Palette struct {
	AccentPrimary   string // active border, tab active, status type
	AccentSecondary string // path, hash
	AccentTertiary  string // stats, search highlight
	BG              string // terminal background
	BGSelected      string // cursor row background
	BorderActive    string // focused panel border
	BorderDim       string // unfocused panel border and dim connectors
	ObjectKey       string
	ArrayIndex      string
	StringVal       string
	NumberVal       string
	BoolTrue        string
	BoolFalse       string
	NullVal         string
	Text            string // cursor foreground
	Muted           string // muted / inactive text
	Error           string
	Path            string
	Hash            string
	Stats           string
	Search          string
	Added           string // diff line added
	Removed         string // diff line removed
}

// PaletteCyber is the default cyber/neural-interface theme.
var PaletteCyber = Palette{
	AccentPrimary:   "#00D4FF",
	AccentSecondary: "#C678DD",
	AccentTertiary:  "#FFD700",
	BG:              "#0D0D0D",
	BGSelected:      "#1A1A2E",
	BorderActive:    "#00D4FF",
	BorderDim:       "#1E1E1E",
	ObjectKey:       "#00D4FF",
	ArrayIndex:      "#FF6B6B",
	StringVal:       "#98C379",
	NumberVal:       "#E5C07B",
	BoolTrue:        "#56B6C2",
	BoolFalse:       "#E06C75",
	NullVal:         "#5C6370",
	Text:            "#FFFFFF",
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
	BGSelected:      "#001A00",
	BorderActive:    "#00FF41",
	BorderDim:       "#003300",
	ObjectKey:       "#00FF41",
	ArrayIndex:      "#33FF33",
	StringVal:       "#00CC22",
	NumberVal:       "#88FF44",
	BoolTrue:        "#00FF41",
	BoolFalse:       "#FF3300",
	NullVal:         "#005500",
	Text:            "#00FF41",
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
	BGSelected:      "#44475A",
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
	BGSelected:      "#3B4252",
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

// ResolvePalette returns the named palette, defaulting to Cyber.
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

