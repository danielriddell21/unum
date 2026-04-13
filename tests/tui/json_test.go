package tui_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/danielriddell21/unum/internal/json/node"
	"github.com/danielriddell21/unum/internal/json/parse"
	jsontui "github.com/danielriddell21/unum/internal/json/render/tui"
)

var sampleFile = filepath.Join("testdata", "sample.json")

func mustParseJSONFile(t *testing.T, path string) *node.Node {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	root, err := parse.Parse(data)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return root
}

// TestTUIJSON_RendersContent checks the initial render contains top-level JSON
// keys from sample.json and that the program quits cleanly on 'q'.
func TestTUIJSON_RendersContent(t *testing.T) {
	root := mustParseJSONFile(t, sampleFile)
	m := jsontui.NewModel(root, sampleFile, "dev")
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("corpus"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

// TestTUIJSON_HelpOverlay checks that '?' opens the help overlay, then sends
// 'q' twice: first to close the overlay, second to quit the program.
func TestTUIJSON_HelpOverlay(t *testing.T) {
	root := mustParseJSONFile(t, sampleFile)
	m := jsontui.NewModel(root, sampleFile, "dev")
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("corpus"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("keyboard reference"))
	}, teatest.WithDuration(3*time.Second))

	// 'q' in help mode closes the overlay; second 'q' quits the program.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

// TestTUIJSON_ArrowNavigation sends arrow keys and quits cleanly — verifies
// no panic during navigation.
func TestTUIJSON_ArrowNavigation(t *testing.T) {
	root := mustParseJSONFile(t, sampleFile)
	m := jsontui.NewModel(root, sampleFile, "dev")
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("corpus"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyUp})

	// Just quit — the goal is no panic, not to assert post-navigation content.
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

// TestTUIJSON_TabCyclesPanels sends Tab three times and quits cleanly.
func TestTUIJSON_TabCyclesPanels(t *testing.T) {
	root := mustParseJSONFile(t, sampleFile)
	m := jsontui.NewModel(root, sampleFile, "dev")
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("corpus"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})

	// Just quit — goal is no panic during panel cycling.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}
