package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/danielriddell21/unum/internal/json/node"
	"github.com/danielriddell21/unum/internal/json/parse"
)

const testJSONFile = "test.json"

func mustParse(t *testing.T, src string) *node.Node {
	t.Helper()
	root, err := parse.Parse([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return root
}

func windowMsg(w, h int) tea.WindowSizeMsg {
	return tea.WindowSizeMsg{Width: w, Height: h}
}

func TestJSONModel_InitNoPanic(t *testing.T) {
	root := mustParse(t, `{"a": 1}`)
	m := NewModel(root, testJSONFile, "dev")
	next, _ := m.Update(windowMsg(120, 40))
	if next == nil {
		t.Fatal("Update returned nil model")
	}
}

func TestJSONModel_SecondWindowSizeResizes(t *testing.T) {
	root := mustParse(t, `{"a": 1}`)
	m := NewModel(root, testJSONFile, "dev")
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)
	next2, _ := nm.Update(windowMsg(160, 50))
	nm2 := next2.(Model)
	if nm2.width != 160 || nm2.height != 50 {
		t.Errorf("after resize: width=%d height=%d, want 160 50", nm2.width, nm2.height)
	}
}

func TestJSONModel_PanelsInitialisedAfterWindowSize(t *testing.T) {
	root := mustParse(t, `{"x": 42}`)
	m := NewModel(root, testJSONFile, "dev")
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)
	if nm.tree.CursorNode() == nil {
		t.Error("tree panel should have a cursor node after init")
	}
}

func TestJSONModel_ViewNoPanic(t *testing.T) {
	root := mustParse(t, `{"hello": "world"}`)
	m := NewModel(root, testJSONFile, "dev")
	next, _ := m.Update(windowMsg(120, 40))
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("View panicked: %v", r)
		}
	}()
	_ = next.(Model).View()
}

func TestJSONModel_EmptyObjectNoPanic(t *testing.T) {
	root := mustParse(t, `{}`)
	m := NewModel(root, "empty.json", "dev")
	next, _ := m.Update(windowMsg(120, 40))
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("View panicked on empty object: %v", r)
		}
	}()
	_ = next.(Model).View()
}
