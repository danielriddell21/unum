package panels

import (
	"strings"
	"testing"
	"time"

	"github.com/danielriddell21/unum/internal/hash/types"
)

func entries(inputs ...string) []types.HistoryEntry {
	out := make([]types.HistoryEntry, len(inputs))
	for i, s := range inputs {
		out[i] = types.HistoryEntry{Input: s, Time: time.Now()}
	}
	return out
}

func TestHistoryPanel_EmptyHistory(t *testing.T) {
	p := NewHistoryPanel(80, 20)
	if p.SelectedInput() != "" {
		t.Error("empty history: SelectedInput should be empty")
	}
	out := p.View()
	if !strings.Contains(out, "no history") {
		t.Errorf("empty history: want 'no history' in %q", out)
	}
}

func TestHistoryPanel_SelectedInput(t *testing.T) {
	p := NewHistoryPanel(80, 20)
	p.SetHistory(entries("alpha", "bravo"))
	if got := p.SelectedInput(); got != "alpha" {
		t.Errorf("SelectedInput=%q, want 'alpha'", got)
	}
}

func TestHistoryPanel_ScrollDown(t *testing.T) {
	p := NewHistoryPanel(80, 20)
	p.SetHistory(entries("alpha", "bravo"))
	p.ScrollDown()
	if got := p.SelectedInput(); got != "bravo" {
		t.Errorf("after ScrollDown: %q, want 'bravo'", got)
	}
}

func TestHistoryPanel_ScrollDownBound(t *testing.T) {
	p := NewHistoryPanel(80, 20)
	p.SetHistory(entries("only"))
	p.ScrollDown()
	p.ScrollDown() // at bottom, should not panic
	if got := p.SelectedInput(); got != "only" {
		t.Errorf("ScrollDown at end: %q, want 'only'", got)
	}
}

func TestHistoryPanel_ScrollUp(t *testing.T) {
	p := NewHistoryPanel(80, 20)
	p.SetHistory(entries("alpha", "bravo"))
	p.ScrollDown()
	p.ScrollUp()
	if got := p.SelectedInput(); got != "alpha" {
		t.Errorf("after ScrollUp: %q, want 'alpha'", got)
	}
}

func TestHistoryPanel_ScrollUpBound(t *testing.T) {
	p := NewHistoryPanel(80, 20)
	p.SetHistory(entries("alpha"))
	p.ScrollUp() // at top, should not panic
	if got := p.SelectedInput(); got != "alpha" {
		t.Errorf("ScrollUp at top: %q", got)
	}
}

func TestHistoryPanel_ResetCursor(t *testing.T) {
	p := NewHistoryPanel(80, 20)
	p.SetHistory(entries("a", "b", "c"))
	p.ScrollDown()
	p.ScrollDown()
	p.ResetCursor()
	if got := p.SelectedInput(); got != "a" {
		t.Errorf("after ResetCursor: %q, want 'a'", got)
	}
}

func TestHistoryPanel_ViewWithEntries(t *testing.T) {
	p := NewHistoryPanel(80, 20)
	p.SetHistory(entries("my-service"))
	out := p.View()
	if !strings.Contains(out, "my-service") {
		t.Errorf("View with history: want 'my-service' in %q", out)
	}
}

func TestHistoryPanel_ViewFocused(t *testing.T) {
	p := NewHistoryPanel(80, 20)
	p.SetHistory(entries("svc"))
	p.SetFocused(true)
	out := p.View()
	if !strings.Contains(out, ">") {
		t.Errorf("focused: want '>' cursor in %q", out)
	}
}

func TestHistoryPanel_SetHistoryClampsOversizedCursor(t *testing.T) {
	p := NewHistoryPanel(80, 20)
	p.SetHistory(entries("a", "b", "c"))
	p.ScrollDown()
	p.ScrollDown()
	// Replace with shorter list — cursor should clamp.
	p.SetHistory(entries("only"))
	if got := p.SelectedInput(); got != "only" {
		t.Errorf("cursor clamp after SetHistory: %q", got)
	}
}

func TestHistoryPanel_Resize(t *testing.T) {
	p := NewHistoryPanel(80, 20)
	p.SetHistory(entries("a", "b"))
	p.Resize(100, 30)
	out := p.View()
	if out == "" {
		t.Error("after Resize: View returned empty")
	}
}
