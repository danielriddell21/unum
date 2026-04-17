package theme

import "testing"

func TestResolvePalette(t *testing.T) {
	tests := []struct {
		name    string
		want    Palette
		wantTag string
	}{
		{"matrix", PaletteMatrix, "matrix"},
		{"dracula", PaletteDracula, "dracula"},
		{"nord", PaletteNord, "nord"},
		{"", PaletteCyber, "default-empty"},
		{"unknown", PaletteCyber, "default-unknown"},
		{"cyber", PaletteCyber, "default-cyber"},
	}
	for _, tc := range tests {
		t.Run(tc.wantTag, func(t *testing.T) {
			got := ResolvePalette(tc.name)
			if got.AccentPrimary != tc.want.AccentPrimary {
				t.Errorf("ResolvePalette(%q).AccentPrimary=%q, want %q",
					tc.name, got.AccentPrimary, tc.want.AccentPrimary)
			}
			if got.BG != tc.want.BG {
				t.Errorf("ResolvePalette(%q).BG=%q, want %q", tc.name, got.BG, tc.want.BG)
			}
		})
	}
}

func TestResolveLightPalette(t *testing.T) {
	tests := []struct {
		name    string
		want    Palette
		wantTag string
	}{
		{"solarized", PaletteSolarized, "solarized"},
		{"", PaletteClean, "default-empty"},
		{"unknown", PaletteClean, "default-unknown"},
		{"clean", PaletteClean, "default-clean"},
	}
	for _, tc := range tests {
		t.Run(tc.wantTag, func(t *testing.T) {
			got := ResolveLightPalette(tc.name)
			if got.AccentPrimary != tc.want.AccentPrimary {
				t.Errorf("ResolveLightPalette(%q).AccentPrimary=%q, want %q",
					tc.name, got.AccentPrimary, tc.want.AccentPrimary)
			}
		})
	}
}

func TestToCSSVars_AllKeysPresent(t *testing.T) {
	wantKeys := []string{
		"--bg", "--bg-panel", "--bg-hover", "--border", "--border-active",
		"--text", "--muted", "--array-idx", "--string", "--number",
		"--bool-true", "--bool-false", "--null", "--path", "--search",
		"--hash", "--stats", "--bg-added", "--bg-removed",
		"--diff-added", "--diff-removed", "--diff-hunk",
	}
	got := PaletteCyber.ToCSSVars()
	if len(got) != len(wantKeys) {
		t.Errorf("ToCSSVars returned %d keys, want %d", len(got), len(wantKeys))
	}
	for _, k := range wantKeys {
		v, ok := got[k]
		if !ok {
			t.Errorf("ToCSSVars missing key %q", k)
			continue
		}
		if v == "" {
			t.Errorf("ToCSSVars[%q] is empty", k)
		}
	}
}

func TestToCSSVars_ValueMapping(t *testing.T) {
	p := Palette{
		BG:            "#111",
		BGPanel:       "#222",
		BGSelected:    "#333",
		BorderDim:     "#444",
		BorderActive:  "#555",
		Text:          "#666",
		Muted:         "#777",
		ArrayIndex:    "#888",
		StringVal:     "#999",
		NumberVal:     "#aaa",
		BoolTrue:      "#bbb",
		BoolFalse:     "#ccc",
		NullVal:       "#ddd",
		Path:          "#eee",
		Search:        "#fff",
		Hash:          "#012",
		Stats:         "#345",
		BGAdded:       "#678",
		BGRemoved:     "#9ab",
		Added:         "#cde",
		Removed:       "#f01",
		AccentPrimary: "#234",
	}
	v := p.ToCSSVars()
	checks := map[string]string{
		"--bg":            "#111",
		"--bg-panel":      "#222",
		"--bg-hover":      "#333",
		"--border":        "#444",
		"--border-active": "#555",
		"--text":          "#666",
		"--muted":         "#777",
		"--array-idx":     "#888",
		"--string":        "#999",
		"--number":        "#aaa",
		"--bool-true":     "#bbb",
		"--bool-false":    "#ccc",
		"--null":          "#ddd",
		"--path":          "#eee",
		"--search":        "#fff",
		"--hash":          "#012",
		"--stats":         "#345",
		"--bg-added":      "#678",
		"--bg-removed":    "#9ab",
		"--diff-added":    "#cde",
		"--diff-removed":  "#f01",
		"--diff-hunk":     "#234",
	}
	for k, want := range checks {
		if v[k] != want {
			t.Errorf("ToCSSVars[%q]=%q, want %q", k, v[k], want)
		}
	}
}

func TestPalettes_AllNonEmpty(t *testing.T) {
	palettes := map[string]Palette{
		"cyber":     PaletteCyber,
		"matrix":    PaletteMatrix,
		"dracula":   PaletteDracula,
		"nord":      PaletteNord,
		"clean":     PaletteClean,
		"solarized": PaletteSolarized,
	}
	for name, p := range palettes {
		if p.BG == "" || p.AccentPrimary == "" || p.BorderActive == "" {
			t.Errorf("palette %q has empty required color", name)
		}
	}
}
