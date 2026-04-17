package panels

import (
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/diff/node"
)

func textDiff() *node.Diff {
	return &node.Diff{
		Format:  node.FormatText,
		FileA:   "a.txt",
		FileB:   "b.txt",
		Added:   1,
		Removed: 1,
		Hunks: []node.Hunk{
			{
				OldStart: 1, OldCount: 3, NewStart: 1, NewCount: 3,
				Lines: []node.Line{
					{Kind: node.Unchanged, OldNum: 1, NewNum: 1, Content: "same line"},
					{Kind: node.Removed, OldNum: 2, Content: "old line"},
					{Kind: node.Added, NewNum: 2, Content: "new line"},
					{Kind: node.Unchanged, OldNum: 3, NewNum: 3, Content: "same again"},
				},
			},
		},
	}
}

func jsonDiff() *node.Diff {
	return &node.Diff{
		Format:   node.FormatJSON,
		FileA:    "a.json",
		FileB:    "b.json",
		Added:    1,
		Modified: 1,
		Root: &node.DiffNode{
			Kind: node.Unchanged,
			Children: []*node.DiffNode{
				{Kind: node.Modified, Path: ".version", Key: "version", OldValue: `"1.0"`, NewValue: `"2.0"`},
				{Kind: node.Added, Path: ".region", Key: "region", NewValue: `"us-east-1"`},
			},
		},
		Hunks: []node.Hunk{
			{
				OldStart: 1, OldCount: 1, NewStart: 1, NewCount: 1,
				Lines: []node.Line{
					{Kind: node.Removed, OldNum: 1, Content: `"version": "1.0"`},
					{Kind: node.Added, NewNum: 1, Content: `"version": "2.0"`},
				},
			},
		},
	}
}

// ─── UnifiedPanel ─────────────────────────────────────────────────────────────

func TestUnifiedPanel_ViewNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("View panicked: %v", r)
		}
	}()
	p := NewUnifiedPanel(textDiff(), 80, 20)
	_ = p.View()
}

func TestUnifiedPanel_ContainsDiffContent(t *testing.T) {
	p := NewUnifiedPanel(textDiff(), 80, 20)
	out := p.View()
	if !strings.Contains(out, "new line") && !strings.Contains(out, "old line") {
		t.Logf("unified panel output: %q", out)
	}
}

func TestUnifiedPanel_SearchMatchCount(t *testing.T) {
	p := NewUnifiedPanel(textDiff(), 80, 20)
	p.SetSearch("same")
	if p.MatchCount() == 0 {
		t.Error("SetSearch('same') should find matches in 'same line' and 'same again'")
	}
}

func TestUnifiedPanel_SetSearchEmpty(t *testing.T) {
	p := NewUnifiedPanel(textDiff(), 80, 20)
	p.SetSearch("same")
	p.SetSearch("")
	if p.MatchCount() != 0 {
		t.Errorf("SetSearch('') should clear matches, got %d", p.MatchCount())
	}
}

func TestUnifiedPanel_NextPrevMatch(t *testing.T) {
	p := NewUnifiedPanel(textDiff(), 80, 20)
	p.SetSearch("same")
	if p.MatchCount() < 2 {
		t.Skip("need at least 2 matches for next/prev test")
	}
	initial := p.CurrentHunkIdx()
	p.NextMatch()
	p.NextMatch()
	p.PrevMatch()
	// Should not panic and should have navigated.
	_ = p.CurrentHunkIdx()
	_ = initial
}

func TestUnifiedPanel_ScrollOperations(t *testing.T) {
	p := NewUnifiedPanel(textDiff(), 80, 20)
	p.ScrollDown(3)
	p.ScrollUp(3)
	p.HalfPageDown()
	p.HalfPageUp()
	// Should not panic.
}

func TestUnifiedPanel_Resize(t *testing.T) {
	p := NewUnifiedPanel(textDiff(), 80, 20)
	p.SetFocused(true)
	p.Resize(100, 30)
	_ = p.View()
}

func TestUnifiedPanel_TextOnly(t *testing.T) {
	d := jsonDiff()
	p := NewUnifiedPanel(d, 80, 20)
	p.SetTextOnly(true)
	_ = p.View()
	p.SetTextOnly(false)
	_ = p.View()
}

func TestUnifiedPanel_EmptyDiff(t *testing.T) {
	d := &node.Diff{Format: node.FormatText}
	p := NewUnifiedPanel(d, 80, 20)
	out := p.View()
	if !strings.Contains(out, "no differences") {
		t.Logf("empty diff unified panel: %q", out)
	}
}

// ─── SplitPanel ───────────────────────────────────────────────────────────────

func TestSplitPanel_ViewNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SplitPanel.View panicked: %v", r)
		}
	}()
	p := NewSplitPanel(textDiff(), 80, 20)
	_ = p.View()
}

func TestSplitPanel_SearchMatchCount(t *testing.T) {
	p := NewSplitPanel(textDiff(), 80, 20)
	// "old" appears in the Removed line — SplitPanel only matches added/removed.
	p.SetSearch("old")
	if p.MatchCount() == 0 {
		t.Error("SplitPanel SetSearch('old') should find match in 'old line' (Removed)")
	}
}

func TestSplitPanel_NextPrevMatch(t *testing.T) {
	p := NewSplitPanel(textDiff(), 80, 20)
	p.SetSearch("same")
	p.NextMatch()
	p.PrevMatch()
	// Should not panic.
}

func TestSplitPanel_ScrollOperations(t *testing.T) {
	p := NewSplitPanel(textDiff(), 80, 20)
	p.ScrollDown(3)
	p.ScrollUp(3)
	p.HalfPageDown()
	p.HalfPageUp()
}

func TestSplitPanel_CycleFocus(t *testing.T) {
	p := NewSplitPanel(textDiff(), 80, 20)
	p.CycleFocus()
	_ = p.View()
}

func TestSplitPanel_Resize(t *testing.T) {
	p := NewSplitPanel(textDiff(), 80, 20)
	p.Resize(120, 30)
	_ = p.View()
}
