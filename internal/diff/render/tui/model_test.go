package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/danielriddell21/unum/internal/diff/node"
)

func textDiff() *node.Diff {
	return &node.Diff{
		Format: node.FormatText,
		Hunks: []node.Hunk{
			{
				OldStart: 1, OldCount: 2, NewStart: 1, NewCount: 2,
				Lines: []node.Line{
					{Kind: node.Removed, OldNum: 1, Content: "old"},
					{Kind: node.Added, NewNum: 1, Content: "new"},
				},
			},
		},
	}
}

func treeDiff() *node.Diff {
	return &node.Diff{
		Format:   node.FormatJSON,
		Added:    1,
		Modified: 0,
		Root: &node.DiffNode{
			Kind: node.Unchanged,
			Children: []*node.DiffNode{
				{Kind: node.Added, Path: ".key", Key: "key", NewValue: `"val"`},
			},
		},
	}
}

func windowMsg(w, h int) tea.WindowSizeMsg {
	return tea.WindowSizeMsg{Width: w, Height: h}
}

func TestDiffModel_InitNoPanic(t *testing.T) {
	m := NewModel(textDiff())
	next, _ := m.Update(windowMsg(120, 40))
	if next == nil {
		t.Fatal("Update returned nil model")
	}
}

func TestDiffModel_Initialized(t *testing.T) {
	m := NewModel(textDiff())
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)
	if !nm.initialized {
		t.Error("model should be initialized after first WindowSizeMsg")
	}
}

func TestDiffModel_SecondWindowSizeResizes(t *testing.T) {
	m := NewModel(textDiff())
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)
	next2, _ := nm.Update(windowMsg(160, 50))
	nm2 := next2.(Model)
	if nm2.width != 160 || nm2.height != 50 {
		t.Errorf("after resize: width=%d height=%d, want 160 50", nm2.width, nm2.height)
	}
}

func TestDiffModel_VKeyTogglesViewForTextDiff(t *testing.T) {
	m := NewModel(textDiff())
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	if nm.view != viewUnified {
		t.Fatalf("initial view=%v, want viewUnified", nm.view)
	}

	next2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	nm2 := next2.(Model)
	if nm2.view != viewSplit {
		t.Errorf("after v: view=%v, want viewSplit", nm2.view)
	}

	next3, _ := nm2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	nm3 := next3.(Model)
	if nm3.view != viewUnified {
		t.Errorf("after second v: view=%v, want viewUnified", nm3.view)
	}
}

func TestDiffModel_VKeyNoOpForTreeDiff(t *testing.T) {
	m := NewModel(treeDiff())
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	next2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	nm2 := next2.(Model)
	if nm2.view != viewUnified {
		t.Errorf("v on tree diff changed view to %v, want viewUnified (no-op)", nm2.view)
	}
}

func TestDiffModel_ViewNoPanic(t *testing.T) {
	m := NewModel(textDiff())
	next, _ := m.Update(windowMsg(120, 40))
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("View panicked: %v", r)
		}
	}()
	_ = next.(Model).View()
}
