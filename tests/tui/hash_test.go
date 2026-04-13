package tui_test

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/danielriddell21/unum/internal/hash/derive"
	hashtui "github.com/danielriddell21/unum/internal/hash/render/tui"
	"github.com/danielriddell21/unum/internal/hash/types"
)

func init() {
	// Wire up real derive + no-op history so the TUI doesn't need disk access.
	hashtui.SetFuncs(
		derive.Derive,
		func(_ string) error { return nil },
		func() []types.HistoryEntry { return nil },
	)
}

// TestTUIHash_RendersInputPlaceholder checks the initial render shows the
// input field placeholder text.
func TestTUIHash_RendersInputPlaceholder(t *testing.T) {
	m := hashtui.NewModel("dev")
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("enter text to hash"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

// TestTUIHash_TypeAndDerive types text and presses enter, then checks that
// derived values (port label) appear in the render.
func TestTUIHash_TypeAndDerive(t *testing.T) {
	m := hashtui.NewModel("dev")
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("enter text to hash"))
	}, teatest.WithDuration(5*time.Second))

	tm.Type("hello-world")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// After deriving, the hash panel renders the result — look for "port" label.
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("port"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

// TestTUIHash_TabSwitchesFocus checks that Tab moves focus to the history
// panel. The status bar changes text when history is focused — we check for
// "tab:back" which only appears in the history-focused status bar.
func TestTUIHash_TabSwitchesFocus(t *testing.T) {
	m := hashtui.NewModel("dev")
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("HASH"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyTab})

	// "tab:back" appears in the history-mode status bar — it changes on focus switch.
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("tab:back"))
	}, teatest.WithDuration(3*time.Second))

	// esc quits from history focus mode.
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

// TestTUIHash_HelpOverlay checks '?' shows the help overlay.
func TestTUIHash_HelpOverlay(t *testing.T) {
	m := hashtui.NewModel("dev")
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("HASH"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("keyboard reference"))
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}
