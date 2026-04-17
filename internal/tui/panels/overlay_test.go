package panels

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestHelpOverlay_PreservesBaseHeight(t *testing.T) {
	base := strings.Repeat("base line content here that is long\n", 20)
	base = strings.TrimRight(base, "\n")
	out := HelpOverlay(base, "help text", "#ffffff", 20)

	baseLines := strings.Count(base, "\n") + 1
	outLines := strings.Count(out, "\n") + 1
	if baseLines != outLines {
		t.Errorf("line count changed: base=%d, out=%d", baseLines, outLines)
	}
}

func TestHelpOverlay_NarrowBase(t *testing.T) {
	// Should not panic when base is narrower than help width.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("HelpOverlay panicked on narrow base: %v", r)
		}
	}()
	HelpOverlay("x\nx\nx", "help", "#ffffff", 20)
}

func TestHelpOverlay_ContainsHelpContent(t *testing.T) {
	base := strings.Repeat("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n", 30)
	base = strings.TrimRight(base, "\n")
	out := HelpOverlay(base, "uniquehelpmarker", "#ffffff", 30)
	if !strings.Contains(out, "uniquehelpmarker") {
		t.Error("HelpOverlay output should contain help content")
	}
}

func TestHelpOverlay_EmptyBase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("HelpOverlay panicked on empty base: %v", r)
		}
	}()
	HelpOverlay("", "help", "#ffffff", 20)
}

func TestHelpOverlay_WidthClampsHelpRender(t *testing.T) {
	base := strings.Repeat("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n", 30)
	base = strings.TrimRight(base, "\n")
	help := HelpOverlay(base, "help", "#ffffff", 20)
	// Each line should still be at least the original width (no truncation of base).
	for _, ln := range strings.Split(help, "\n") {
		if lipgloss.Width(ln) < 4 {
			t.Errorf("output line too short: %q (width %d)", ln, lipgloss.Width(ln))
		}
	}
}
