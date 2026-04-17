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

func TestJSONModel_TabCyclesFocus(t *testing.T) {
	root := mustParse(t, `{"a": 1}`)
	m := NewModel(root, testJSONFile, "dev")
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	initialFocus := nm.focused

	nm2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyTab})
	nm = nm2.(Model)
	if nm.focused == initialFocus {
		t.Error("tab should cycle focus")
	}
}

func TestJSONModel_QuestionMarkTogglesHelp(t *testing.T) {
	root := mustParse(t, `{"a": 1}`)
	m := NewModel(root, testJSONFile, "dev")
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	nm2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	nm = nm2.(Model)
	if nm.currentMode != modeHelp {
		t.Error("? should enter help mode")
	}

	nm3, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	nm = nm3.(Model)
	if nm.currentMode == modeHelp {
		t.Error("? in help mode should exit help")
	}
}

func TestJSONModel_SlashEntersSearchMode(t *testing.T) {
	root := mustParse(t, `{"a": 1}`)
	m := NewModel(root, testJSONFile, "dev")
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	nm2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	nm = nm2.(Model)
	if nm.currentMode != modeSearch {
		t.Error("/ should enter search mode")
	}
}

func TestJSONModel_EscExitsSearchMode(t *testing.T) {
	root := mustParse(t, `{"a": 1}`)
	m := NewModel(root, testJSONFile, "dev")
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	nm2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	nm = nm2.(Model)

	nm3, _ := nm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	nm = nm3.(Model)
	if nm.currentMode != modeNormal {
		t.Error("esc should exit search mode")
	}
}

func TestJSONModel_OKeySetsReloadRequest(t *testing.T) {
	root := mustParse(t, `{"a": 1}`)
	m := NewModel(root, testJSONFile, "dev")
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	nm2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	nm = nm2.(Model)
	if !nm.ReloadRequest {
		t.Error("o key should set ReloadRequest")
	}
}

func TestJSONModel_LensDigitKeys(t *testing.T) {
	root := mustParse(t, `{"a": 1}`)
	m := NewModel(root, testJSONFile, "dev")
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	// Digit keys should not panic.
	for _, digit := range []string{"1", "2", "3", "4", "5", "6"} {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("digit key %q panicked: %v", digit, r)
			}
		}()
		nm2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(digit)})
		nm = nm2.(Model)
	}
}

func TestJSONModel_ShiftTabCyclesBackward(t *testing.T) {
	root := mustParse(t, `{"a": 1}`)
	m := NewModel(root, testJSONFile, "dev")
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)
	initialFocus := nm.focused

	nm2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	nm = nm2.(Model)
	if nm.focused == initialFocus {
		t.Error("shift+tab should cycle focus backward")
	}
}
