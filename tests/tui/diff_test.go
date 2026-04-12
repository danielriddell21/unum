package tui_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	diffnode "github.com/danielriddell21/unum/internal/diff/node"
	diffparse "github.com/danielriddell21/unum/internal/diff/parse"
	difftui "github.com/danielriddell21/unum/internal/diff/render/tui"
)

var (
	diffA     = filepath.Join("testdata", "diff-a.txt")
	diffB     = filepath.Join("testdata", "diff-b.txt")
	diffAJSON = filepath.Join("testdata", "diff-a.json")
	diffBJSON = filepath.Join("testdata", "diff-b.json")
	diffATF   = filepath.Join("testdata", "diff-a.tfplan.json")
)

func mustParseDiffFiles(t *testing.T, pathA, pathB string) *diffnode.Diff {
	t.Helper()
	dataA, err := os.ReadFile(pathA)
	if err != nil {
		t.Fatalf("read %s: %v", pathA, err)
	}
	dataB, err := os.ReadFile(pathB)
	if err != nil {
		t.Fatalf("read %s: %v", pathB, err)
	}
	d, err := diffparse.Text(dataA, dataB, 3)
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	d.FileA = pathA
	d.FileB = pathB
	return d
}

func mustParseJSONDiffFiles(t *testing.T, pathA, pathB string) *diffnode.Diff {
	t.Helper()
	dataA, err := os.ReadFile(pathA)
	if err != nil {
		t.Fatalf("read %s: %v", pathA, err)
	}
	dataB, err := os.ReadFile(pathB)
	if err != nil {
		t.Fatalf("read %s: %v", pathB, err)
	}
	d, err := diffparse.JSON(dataA, dataB)
	if err != nil {
		t.Fatalf("json diff: %v", err)
	}
	d.FileA = pathA
	d.FileB = pathB
	d.Format = diffnode.FormatJSON
	return d
}

// TestTUIDiff_RendersHunkMarkers checks the initial render contains the DIFF
// panel title.
func TestTUIDiff_RendersHunkMarkers(t *testing.T) {
	d := mustParseDiffFiles(t, diffA, diffB)
	m := difftui.NewModel(d)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("DIFF"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

// TestTUIDiff_HelpOverlay checks that '?' shows new help text.
// Note: 'q' in help mode exits the overlay only; a second 'q' quits.
func TestTUIDiff_HelpOverlay(t *testing.T) {
	d := mustParseDiffFiles(t, diffA, diffB)
	m := difftui.NewModel(d)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("DIFF"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("keyboard reference"))
	}, teatest.WithDuration(3*time.Second))

	// First 'q' closes the overlay; second 'q' quits the program.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

// TestTUIDiff_ViewToggle sends 'v' twice and quits cleanly — verifies no
// panic during view cycling (unified → split → unified).
func TestTUIDiff_ViewToggle(t *testing.T) {
	d := mustParseDiffFiles(t, diffA, diffB)
	m := difftui.NewModel(d)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("DIFF"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})

	// Just quit — goal is no panic, not post-toggle content assertion.
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

// TestTUIDiff_SemanticViewRendersJSON checks the semantic diff view for JSON
// files shows the DIFF panel.
func TestTUIDiff_SemanticViewRendersJSON(t *testing.T) {
	d := mustParseJSONDiffFiles(t, diffAJSON, diffBJSON)
	m := difftui.NewModel(d)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("DIFF"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func mustParseTerraformDiff(t *testing.T, path string) *diffnode.Diff {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	d, err := diffparse.Terraform(data)
	if err != nil {
		t.Fatalf("terraform diff: %v", err)
	}
	d.FileA = path
	d.FileB = path
	d.Format = diffnode.FormatTerraform
	return d
}

// TestTUIDiff_SemanticViewRendersTerraform checks the semantic diff view for a
// Terraform plan file shows the DIFF panel.
func TestTUIDiff_SemanticViewRendersTerraform(t *testing.T) {
	d := mustParseTerraformDiff(t, diffATF)
	m := difftui.NewModel(d)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("DIFF"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}
