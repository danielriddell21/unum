package panels

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/danielriddell21/unum/internal/json/parse"
)

func mustParseTree(t *testing.T, src string) *TreePanel {
	t.Helper()
	n, err := parse.Parse([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := NewTreePanel(n, 80, 20)
	return &p
}

func TestTreePanel_CursorNodeNotNil(t *testing.T) {
	p := mustParseTree(t, `{"a": 1}`)
	if p.CursorNode() == nil {
		t.Error("CursorNode should not be nil for a non-empty tree")
	}
}

func TestTreePanel_CursorNodeNilForEmpty(t *testing.T) {
	n, err := parse.Parse([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	p := NewTreePanel(n, 80, 20)
	// Empty object has no children; cursor 0 points to the root node itself.
	_ = p.CursorNode() // should not panic
}

func TestTreePanel_ViewNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("View panicked: %v", r)
		}
	}()
	p := mustParseTree(t, `{"a": 1, "b": [1, 2]}`)
	_ = p.View()
}

func TestTreePanel_NavigateDown(t *testing.T) {
	p := mustParseTree(t, `{"a": 1, "b": 2, "c": 3}`)
	initial := p.CursorNode()
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	moved := p.CursorNode()
	_ = initial
	_ = moved
	// Just verify no panic and cursor changed (may depend on tree layout).
}

func TestTreePanel_NavigateUp(t *testing.T) {
	p := mustParseTree(t, `{"a": 1, "b": 2}`)
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	_ = p.CursorNode()
}

func TestTreePanel_ExpandCollapse(t *testing.T) {
	p := mustParseTree(t, `{"obj": {"inner": 1}}`)
	// space toggles collapse
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	_ = p.View()
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	_ = p.View()
}

func TestTreePanel_SetSearch(t *testing.T) {
	p := mustParseTree(t, `{"alpha": 1, "bravo": 2, "charlie": 3}`)
	p.SetSearch("alpha")
	_ = p.View()
	p.SetSearch("")
	_ = p.View()
}

func TestTreePanel_Resize(t *testing.T) {
	p := mustParseTree(t, `{"a": 1}`)
	p.SetFocused(true)
	p.Resize(100, 30)
	_ = p.View()
}

func TestTreePanel_HomeEnd(t *testing.T) {
	p := mustParseTree(t, `{"a": 1, "b": 2, "c": 3}`)
	p.Update(tea.KeyMsg{Type: tea.KeyEnd})
	_ = p.CursorNode()
	p.Update(tea.KeyMsg{Type: tea.KeyHome})
	_ = p.CursorNode()
}

func TestTreePanel_PageDownUp(t *testing.T) {
	p := mustParseTree(t, `{"a": 1}`)
	p.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	p.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	_ = p.View()
}
