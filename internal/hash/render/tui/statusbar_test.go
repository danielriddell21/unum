package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/danielriddell21/unum/internal/hash/types"
)

func TestStatusBar_NilResult(t *testing.T) {
	out := statusBar(nil, focusInput, 80, "v1.0")
	if !strings.Contains(out, "no input yet") {
		t.Errorf("nil result: want 'no input yet' in %q", out)
	}
	if lipgloss.Width(out) != 80 {
		t.Errorf("width=%d, want 80", lipgloss.Width(out))
	}
}

func TestStatusBar_WithResult(t *testing.T) {
	r := &types.Result{Input: "my-service", Port: 8080}
	out := statusBar(r, focusInput, 80, "v2.0")
	if !strings.Contains(out, "my-service") {
		t.Errorf("result: want 'my-service' in %q", out)
	}
}

func TestStatusBar_HistoryFocusHints(t *testing.T) {
	out := statusBar(nil, focusHistory, 80, "v1.0")
	if !strings.Contains(out, "navigate") {
		t.Errorf("history focus: want 'navigate' hint in %q", out)
	}
}

func TestStatusBar_InputFocusHints(t *testing.T) {
	out := statusBar(nil, focusInput, 80, "v1.0")
	if !strings.Contains(out, "derive") {
		t.Errorf("input focus: want 'derive' hint in %q", out)
	}
}
