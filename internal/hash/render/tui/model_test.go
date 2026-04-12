package tui

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/danielriddell21/unum/internal/hash/types"
)

// testDerive is a minimal derive function that doesn't import hash (avoids cycle in test).
func testDerive(input string) types.Result {
	h := sha256.Sum256([]byte(input))
	n := binary.BigEndian.Uint16(h[0:2])
	port := uint16(1024 + n%(65535-1024+1))
	color := "#" + hex.EncodeToString(h[0:3])
	short := hex.EncodeToString(h[0:4])
	return types.Result{
		Input:  input,
		Port:   port,
		UUID:   fmt.Sprintf("test-uuid-%s", short),
		Color:  color,
		Short:  short,
		Emoji:  "🦊",
		Phrase: "alpha-bravo-charlie",
	}
}

func testAppendHistory(_ string) error        { return nil }
func testLoadHistory() []types.HistoryEntry   { return []types.HistoryEntry{{Input: "prev", Time: time.Now()}} }

func init() {
	SetFuncs(testDerive, testAppendHistory, testLoadHistory)
}

func windowMsg(w, h int) tea.WindowSizeMsg {
	return tea.WindowSizeMsg{Width: w, Height: h}
}

func TestHashTUI_InitNoPanic(t *testing.T) {
	m := NewModel()
	next, _ := m.Update(windowMsg(120, 40))
	if next == nil {
		t.Fatal("Update returned nil model")
	}
}

func TestHashTUI_SecondWindowSizeResizes(t *testing.T) {
	m := NewModel()
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)
	next2, _ := nm.Update(windowMsg(160, 50))
	nm2 := next2.(Model)
	if nm2.width != 160 || nm2.height != 50 {
		t.Errorf("after resize: width=%d height=%d, want 160 50", nm2.width, nm2.height)
	}
}

func TestHashTUI_ViewNoPanic(t *testing.T) {
	m := NewModel()
	next, _ := m.Update(windowMsg(120, 40))
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("View panicked: %v", r)
		}
	}()
	_ = next.(Model).View()
}

func TestHashTUI_EnterKeyDerives(t *testing.T) {
	m := NewModel()
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	// Type "hello"
	for _, ch := range "hello" {
		nm2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		nm = nm2.(Model)
	}

	// Press enter
	nm2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	nm = nm2.(Model)

	if nm.result == nil {
		t.Fatal("result should be non-nil after Enter")
	}
	if nm.result.Input != "hello" {
		t.Errorf("result.Input = %q, want hello", nm.result.Input)
	}
}

func TestHashTUI_EmptyEnterNoResult(t *testing.T) {
	m := NewModel()
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	nm2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	nm = nm2.(Model)

	if nm.result != nil {
		t.Error("empty enter should not set result")
	}
}

func TestHashTUI_TabSwitchesFocus(t *testing.T) {
	m := NewModel()
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	if nm.focused != focusInput {
		t.Fatal("initial focus should be input")
	}

	nm2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyTab})
	nm = nm2.(Model)

	if nm.focused != focusHistory {
		t.Error("after tab, focus should be history")
	}

	nm3, _ := nm.Update(tea.KeyMsg{Type: tea.KeyTab})
	nm = nm3.(Model)

	if nm.focused != focusInput {
		t.Error("after second tab, focus should return to input")
	}
}

func TestHashTUI_QTypesInInputMode(t *testing.T) {
	m := NewModel()
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	// In input mode, 'q' should be typed into the input, not quit
	nm2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	nm = nm2.(Model)

	if nm.input.Value() != "q" {
		t.Errorf("q should type into input, got %q", nm.input.Value())
	}
}

func TestHashTUI_HelpToggle(t *testing.T) {
	m := NewModel()
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	if nm.showHelp {
		t.Fatal("showHelp should be false initially")
	}

	nm2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	nm = nm2.(Model)
	if !nm.showHelp {
		t.Error("after ?: showHelp should be true")
	}

	nm3, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	nm = nm3.(Model)
	if nm.showHelp {
		t.Error("after second ?: showHelp should be false")
	}
}

func TestHashTUI_HistoryNavigation(t *testing.T) {
	m := NewModel()
	next, _ := m.Update(windowMsg(120, 40))
	nm := next.(Model)

	// Switch to history mode
	nm2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyTab})
	nm = nm2.(Model)

	if len(nm.history) == 0 {
		t.Skip("no history entries to navigate")
	}

	initial := nm.historyCursor

	// down
	nm3, _ := nm.Update(tea.KeyMsg{Type: tea.KeyDown})
	nm = nm3.(Model)

	// If there's only 1 entry cursor stays at 0; otherwise it moves
	if len(nm.history) > 1 && nm.historyCursor == initial {
		t.Error("cursor should have moved down")
	}
}
