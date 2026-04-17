package panels

import (
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/diff/node"
)

func TestInfoPanel_NilDiff(t *testing.T) {
	p := NewInfoPanel(nil, 80, 20)
	out := p.View()
	if !strings.Contains(out, "no diff") {
		t.Errorf("nil diff: want 'no diff' in %q", out)
	}
}

func TestInfoPanel_WithFileNames(t *testing.T) {
	d := &node.Diff{FileA: "a.txt", FileB: "b.txt", Added: 1, Removed: 1}
	p := NewInfoPanel(d, 80, 20)
	out := p.View()
	if !strings.Contains(out, "a.txt") || !strings.Contains(out, "b.txt") {
		t.Errorf("file names missing in info panel: %q", out)
	}
}

func TestInfoPanel_WithHunks(t *testing.T) {
	d := &node.Diff{
		Added: 1,
		Hunks: []node.Hunk{{OldStart: 1, OldCount: 3, NewStart: 1, NewCount: 4}},
	}
	p := NewInfoPanel(d, 80, 20)
	out := p.View()
	if !strings.Contains(out, "Hunk 1 of 1") {
		t.Errorf("hunk info missing: %q", out)
	}
}

func TestInfoPanel_SetHunkIdx(t *testing.T) {
	d := &node.Diff{
		Hunks: []node.Hunk{
			{OldStart: 1, NewStart: 1},
			{OldStart: 10, NewStart: 10},
		},
	}
	p := NewInfoPanel(d, 80, 20)
	p.SetHunkIdx(1)
	out := p.View()
	if !strings.Contains(out, "Hunk 2 of 2") {
		t.Errorf("SetHunkIdx(1): want 'Hunk 2 of 2' in %q", out)
	}
}

func TestInfoPanel_WithModifiedCount(t *testing.T) {
	d := &node.Diff{Added: 1, Removed: 0, Modified: 2}
	p := NewInfoPanel(d, 80, 20)
	out := p.View()
	if !strings.Contains(out, "modified") {
		t.Errorf("modified count: want 'modified' in %q", out)
	}
}

func TestInfoPanel_WithRoot(t *testing.T) {
	d := &node.Diff{Format: node.FormatJSON, Root: &node.DiffNode{}}
	p := NewInfoPanel(d, 80, 20)
	out := p.View()
	if !strings.Contains(out, "json") {
		t.Errorf("structured diff: want format badge 'json' in %q", out)
	}
}

func TestInfoPanel_Resize(t *testing.T) {
	d := makeDiff(0, 0)
	p := NewInfoPanel(d, 80, 20)
	p.SetFocused(true)
	p.Resize(100, 30)
	// Should not panic.
	_ = p.View()
}
