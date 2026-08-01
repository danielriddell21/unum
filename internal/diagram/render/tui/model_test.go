package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func sizedModel() Model {
	m := NewModel("1.2.3", Info{
		File:   "sample.d2",
		Lang:   "d2",
		Source: "client -> api\napi -> db",
		ASCII:  "┌────────┐\n│ client │\n└────────┘",
		SVG:    []byte("<svg/>"),
		PNG:    []byte("\x89PNG..."),
		Drawio: []byte("<mxGraphModel/>"),
		Width:  155,
		Height: 458,
		Shapes: 3,
		Conns:  2,
	})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 30})
	return next.(Model)
}

func TestViewShowsPanelsAndPreview(t *testing.T) {
	view := sizedModel().View()
	for _, want := range []string{"SOURCE", "RENDER", "155×458", "client -> api"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q", want)
		}
	}
	if !strings.Contains(view, "┌") {
		t.Error("expected native ascii preview in view")
	}
}

func TestNoPreviewNote(t *testing.T) {
	// A diagram with no ascii (chart/timeline types) shows a can't-render note.
	m := NewModel("dev", Info{File: "p.mmd", Lang: "mermaid", Source: "pie", Shapes: -1})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 30})
	if !strings.Contains(next.(Model).View(), "no terminal preview") {
		t.Error("expected a can't-render note when no ascii preview is available")
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

func TestSaveKeyWritesFile(t *testing.T) {
	dir := t.TempDir()
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	m := sizedModel()
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if !strings.Contains(next.(Model).status, "diagram.svg") {
		t.Errorf("status = %q, want mention of diagram.svg", next.(Model).status)
	}
	if _, err := os.Stat(filepath.Join(dir, "diagram.svg")); err != nil {
		t.Errorf("diagram.svg not written: %v", err)
	}
}

func TestAsciiPreview(t *testing.T) {
	m := NewModel("dev", Info{
		File:   "a.d2",
		Lang:   "d2",
		Source: "x -> y",
		ASCII:  "┌───┐\n│ x │\n└───┘",
		Width:  40,
		Height: 30,
		Shapes: -1,
	})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 30})
	view := next.(Model).View()
	if !strings.Contains(view, "┌") || !strings.Contains(view, "x") {
		t.Errorf("expected native ascii diagram in view:\n%s", view)
	}
}

func TestErrorInfo(t *testing.T) {
	m := NewModel("1.0", Info{File: "x.mmd", Lang: "mermaid", Source: "flowchart", Shapes: -1, Err: errStub{}})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if !strings.Contains(next.(Model).View(), "preview unavailable") {
		t.Error("expected error message in render panel")
	}
}

type errStub struct{}

func (errStub) Error() string { return "boom" }
