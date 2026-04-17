package panels

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestPanelTitle_ContainsTitle(t *testing.T) {
	for _, active := range []bool{true, false} {
		got := PanelTitle("MyTitle", active)
		if !strings.Contains(got, "MyTitle") {
			t.Errorf("PanelTitle(active=%v) missing title: %q", active, got)
		}
	}
}

func TestWrapPanel_ClampsTinyDimensions(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("WrapPanel panicked on tiny dimensions: %v", r)
		}
	}()
	out := WrapPanel("hi", true, 1, 1)
	if lipgloss.Width(out) == 0 {
		t.Error("WrapPanel should produce non-empty output even for clamped dimensions")
	}
}

func TestWrapPanel_HasBorder(t *testing.T) {
	out := WrapPanel("body", false, 20, 5)
	// Rounded border characters appear in the output.
	if !strings.ContainsAny(out, "╭╮╰╯─│") {
		t.Errorf("WrapPanel output should include border chars: %q", out)
	}
}

func TestApplyBaseStyles_RebuildsColors(t *testing.T) {
	original := ColorBG
	defer func() {
		// Restore default cyber palette to keep package state clean for other tests.
		ApplyBaseStyles(PaletteCyber)
		if ColorBG != original {
			t.Errorf("failed to restore palette state: ColorBG=%q want %q", ColorBG, original)
		}
	}()

	ApplyBaseStyles(PaletteMatrix)
	if ColorBG != PaletteMatrix.BG {
		t.Errorf("ColorBG=%q, want %q", ColorBG, PaletteMatrix.BG)
	}
	if ColorActiveBorder != PaletteMatrix.BorderActive {
		t.Errorf("ColorActiveBorder=%q, want %q", ColorActiveBorder, PaletteMatrix.BorderActive)
	}
}
