package tui

import (
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func stubImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for i := range img.Pix {
		img.Pix[i] = 200
	}
	return img
}

func sizedModel() Model {
	m := NewModel("1.2.3", Info{
		File:   "sample.d2",
		Lang:   "d2",
		Source: "client -> api\napi -> db",
		SVG:    []byte("<svg/>"),
		PNG:    []byte("\x89PNG..."),
		Drawio: []byte("<mxGraphModel/>"),
		Img:    stubImage(),
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
	if !strings.Contains(view, "▀") {
		t.Error("expected rendered image art (half blocks) in view")
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

func TestTabTogglesFocus(t *testing.T) {
	m := sizedModel()
	if m.focused != focusSource {
		t.Fatalf("initial focus = %v, want focusSource", m.focused)
	}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if next.(Model).focused != focusRender {
		t.Error("tab should move focus to render panel")
	}
	back, _ := next.(Model).Update(tea.KeyMsg{Type: tea.KeyTab})
	if back.(Model).focused != focusSource {
		t.Error("tab should cycle focus back to source")
	}
}

func TestRenderFocusPanAndZoom(t *testing.T) {
	m := sizedModel()
	m.focused = focusRender

	right, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if right.(Model).offX != panStep {
		t.Errorf("offX = %d after pan right, want %d", right.(Model).offX, panStep)
	}
	if right.(Model).scroll != 0 {
		t.Error("source must not scroll while render is focused")
	}

	zin, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	if zin.(Model).zoom <= 1 {
		t.Errorf("zoom = %v after '+', want > 1", zin.(Model).zoom)
	}

	reset, _ := zin.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("0")})
	rm := reset.(Model)
	if rm.zoom != 1 || rm.offX != 0 || rm.offY != 0 {
		t.Errorf("reset gave zoom=%v offX=%d offY=%d, want 1/0/0", rm.zoom, rm.offX, rm.offY)
	}
}

func TestZoomReRendersAscii(t *testing.T) {
	calls := 0
	m := NewModel("dev", Info{
		File:        "a.d2",
		Lang:        "d2",
		Source:      "x -> y",
		ASCII:       "small",
		ASCIIRender: func(scale float64) string { calls++; return "zoomed" },
		Width:       40,
		Height:      30,
		Shapes:      -1,
	})
	sized, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 30})
	sm := sized.(Model)
	sm.focused = focusRender
	next, _ := sm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	if calls == 0 {
		t.Fatal("zoom should call ASCIIRender to re-render at scale")
	}
	if next.(Model).ascii != "zoomed" {
		t.Errorf("ascii = %q after zoom, want re-rendered %q", next.(Model).ascii, "zoomed")
	}
}

func TestZoomClamps(t *testing.T) {
	m := sizedModel()
	m.focused = focusRender
	m.zoom = zoomMin
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	if out.(Model).zoom < zoomMin {
		t.Errorf("zoom = %v, must not drop below %v", out.(Model).zoom, zoomMin)
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
