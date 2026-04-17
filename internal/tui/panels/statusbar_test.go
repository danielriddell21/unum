package panels

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestPad_Extends(t *testing.T) {
	got := Pad("hi", 5)
	if got != "hi   " {
		t.Errorf("Pad(\"hi\", 5)=%q, want %q", got, "hi   ")
	}
}

func TestPad_NoOpWhenWiderThanWidth(t *testing.T) {
	got := Pad("hello", 3)
	if got != "hello" {
		t.Errorf("Pad(\"hello\", 3)=%q, want unchanged", got)
	}
}

func TestPad_EqualWidth(t *testing.T) {
	got := Pad("abc", 3)
	if got != "abc" {
		t.Errorf("Pad equal width=%q, want %q", got, "abc")
	}
}

func TestBar_FillsToWidth(t *testing.T) {
	got := Bar("L", "R", 10)
	if lipgloss.Width(got) != 10 {
		t.Errorf("Bar width=%d, want 10 (got %q)", lipgloss.Width(got), got)
	}
	if !strings.HasPrefix(got, "L") {
		t.Errorf("Bar should start with left: %q", got)
	}
	if !strings.HasSuffix(got, "R") {
		t.Errorf("Bar should end with right: %q", got)
	}
}

func TestBar_MinSpacingWhenTooNarrow(t *testing.T) {
	// Total content (3+3) > width (4) — should still leave at least one space.
	got := Bar("AAA", "BBB", 4)
	if !strings.Contains(got, " ") {
		t.Errorf("Bar should contain at least one space, got %q", got)
	}
}

func TestSearchBar_ContainsLabelQuerySuffix(t *testing.T) {
	style := lipgloss.NewStyle()
	got := SearchBar("FIND", "foo", " (1/3)", 40, style)
	for _, want := range []string{"FIND", "foo", "(1/3)"} {
		if !strings.Contains(got, want) {
			t.Errorf("SearchBar missing %q in %q", want, got)
		}
	}
}

func TestSearchBar_PaddedToWidth(t *testing.T) {
	style := lipgloss.NewStyle()
	got := SearchBar("FIND", "x", "", 40, style)
	if lipgloss.Width(got) != 40 {
		t.Errorf("SearchBar width=%d, want 40", lipgloss.Width(got))
	}
}
