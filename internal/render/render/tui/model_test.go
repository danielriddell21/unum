package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func sizedModel() Model {
	m := NewModel("1.2.3", Info{
		File:   "sample.d2",
		Lang:   "d2",
		Source: "client -> api\napi -> db",
		Width:  155,
		Height: 458,
		Bytes:  1800,
		Shapes: 3,
		Conns:  2,
	})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 30})
	return next.(Model)
}

func TestViewShowsPanelsAndInfo(t *testing.T) {
	view := sizedModel().View()
	for _, want := range []string{"SOURCE", "RENDER", "d2", "155 × 458", "client -> api"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q", want)
		}
	}
}

func TestHelpToggle(t *testing.T) {
	m := sizedModel()
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	if !strings.Contains(next.(Model).View(), "keyboard reference") {
		t.Error("help overlay not shown after '?'")
	}
}

func TestQuitKey(t *testing.T) {
	m := sizedModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("expected tea.QuitMsg from quit key")
	}
}

func TestScrollClamp(t *testing.T) {
	m := sizedModel()
	up, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if up.(Model).scroll != 0 {
		t.Error("scroll should clamp at 0")
	}
	down, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if down.(Model).scroll != 1 {
		t.Errorf("scroll = %d, want 1", down.(Model).scroll)
	}
}

func TestErrorInfo(t *testing.T) {
	m := NewModel("1.0", Info{File: "x.mmd", Lang: "mermaid", Source: "flowchart", Shapes: -1, Err: errStub{}})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if !strings.Contains(next.(Model).View(), "render failed") {
		t.Error("expected error message in render panel")
	}
}

type errStub struct{}

func (errStub) Error() string { return "boom" }
