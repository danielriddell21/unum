package panels

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/danielriddell21/unum/internal/diff/node"
)

func makeDiff(added, removed int) *node.Diff {
	return &node.Diff{
		Format:  node.FormatText,
		Added:   added,
		Removed: removed,
	}
}

func TestStatusBar_FillsWidth(t *testing.T) {
	d := makeDiff(2, 1)
	out := StatusBar(d, "unified", 80, false, "", 0, "v1")
	if lipgloss.Width(out) < 10 {
		t.Errorf("StatusBar too short: %q", out)
	}
}

func TestStatusBar_SearchMode(t *testing.T) {
	d := makeDiff(0, 0)
	out := StatusBar(d, "unified", 80, true, "foo", 3, "v1")
	if !strings.Contains(out, "foo") {
		t.Errorf("search mode: want 'foo' in %q", out)
	}
}

func TestStatusBar_NoSearchMode(t *testing.T) {
	d := makeDiff(1, 0)
	out := StatusBar(d, "unified", 80, false, "", 0, "v1")
	// Should include format/view hints.
	if out == "" {
		t.Error("status bar should not be empty")
	}
}

func TestStatusBar_ViewNames(t *testing.T) {
	d := makeDiff(0, 0)
	for _, view := range []string{"unified", "split", "semantic"} {
		out := StatusBar(d, view, 80, false, "", 0, "v1")
		if out == "" {
			t.Errorf("StatusBar view=%q returned empty", view)
		}
	}
}
