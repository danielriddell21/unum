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

func TestDiffModel_VKeyNoopForSemanticOnlyDiff(t *testing.T) {
	m := NewModel(treeDiff())
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	if nm.view != viewSemantic {
		t.Fatalf("initial view=%v, want viewSemantic", nm.view)
	}

	next2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	nm2 := next2.(Model)
	if nm2.view != viewSemantic {
		t.Errorf("v on semantic-only diff: view=%v, want viewSemantic (unchanged)", nm2.view)
	}
}

func TestDiffModel_VKeyCyclesThreeModes(t *testing.T) {
	// diff with both Root (semantic) and Hunks (text)
	d := &node.Diff{
		Format: node.FormatJSON,
		Added:  1,
		Root: &node.DiffNode{
			Kind:     node.Unchanged,
			Children: []*node.DiffNode{{Kind: node.Added, Path: ".key", NewValue: `"v"`}},
		},
		Hunks: []node.Hunk{{
			OldStart: 1, OldCount: 1, NewStart: 1, NewCount: 1,
			Lines: []node.Line{{Kind: node.Added, NewNum: 1, Content: `"key": "v"`}},
		}},
	}
	m := NewModel(d)
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	if nm.view != viewSemantic {
		t.Fatalf("initial view=%v, want viewSemantic", nm.view)
	}
	next, _ = nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	nm = next.(Model)
	if nm.view != viewUnified {
		t.Errorf("v #1: view=%v, want viewUnified", nm.view)
	}
	next, _ = nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	nm = next.(Model)
	if nm.view != viewSplit {
		t.Errorf("v #2: view=%v, want viewSplit", nm.view)
	}
	next, _ = nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	nm = next.(Model)
	if nm.view != viewSemantic {
		t.Errorf("v #3: view=%v, want viewSemantic", nm.view)
	}
}

func TestDiffModel_QuestionMarkEntersHelpMode(t *testing.T) {
	m := NewModel(textDiff())
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	if nm.mode != modeNormal {
		t.Fatalf("initial mode=%v, want modeNormal", nm.mode)
	}

	next2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	nm2 := next2.(Model)
	if nm2.mode != modeHelp {
		t.Errorf("after ?: mode=%v, want modeHelp", nm2.mode)
	}
}

func TestDiffModel_EscExitsHelpMode(t *testing.T) {
	m := NewModel(textDiff())
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)
	nm.mode = modeHelp

	next2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	nm2 := next2.(Model)
	if nm2.mode != modeNormal {
		t.Errorf("after esc in help: mode=%v, want modeNormal", nm2.mode)
	}
}

func TestDiffModel_OKeySetsReloadRequest(t *testing.T) {
	m := NewModel(textDiff())
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	next2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	nm2 := next2.(Model)
	if !nm2.ReloadRequest {
		t.Error("o key should set ReloadRequest=true")
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
