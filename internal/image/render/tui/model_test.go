package tui

import (
	"errors"
	"image"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/danielriddell21/unum/internal/image/optimize"
)

func testSource() optimize.Source {
	return optimize.Source{
		Name:   "photo.jpg",
		Image:  image.NewRGBA(image.Rect(0, 0, 40, 20)),
		Format: optimize.FormatJPEG,
		Width:  40,
		Height: 20,
		Bytes:  10000,
	}
}

func stubDeps() Deps {
	return Deps{
		Optimize: func(src optimize.Source, opts optimize.Options) (optimize.Result, error) {
			return optimize.Result{
				Format:  opts.Format,
				Quality: opts.Quality,
				Width:   src.Width,
				Height:  src.Height,
				Bytes:   opts.Quality * 10,
			}, nil
		},
		Ladder: func(_ optimize.Source, _ optimize.Options, qualities []int) ([]optimize.Step, error) {
			steps := make([]optimize.Step, len(qualities))
			for i, q := range qualities {
				steps[i] = optimize.Step{Quality: q, Bytes: q * 10}
			}
			return steps, nil
		},
		Save: func(string, []byte) error { return nil },
	}
}

func newTestModel() Model {
	m := NewModel(testSource(), optimize.Options{Quality: 80}, "out.jpg", "dev", stubDeps())
	sized, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return sized.(Model)
}

func TestNewModelNormalizesOptions(t *testing.T) {
	m := NewModel(testSource(), optimize.Options{Quality: 500, Scale: 0}, "out.jpg", "dev", stubDeps())

	if m.opts.Quality != 100 {
		t.Errorf("Quality = %d, want it clamped to 100", m.opts.Quality)
	}
	if m.opts.Scale != 1 {
		t.Errorf("Scale = %v, want an unset scale to become 1", m.opts.Scale)
	}
	if m.opts.Format != optimize.FormatJPEG {
		t.Errorf("Format = %v, want it resolved from the source", m.opts.Format)
	}
}

func TestQualityKeys(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want int
	}{
		{"right raises quality", "right", 85},
		{"l raises quality", "l", 85},
		{"left lowers quality", "left", 75},
		{"h lowers quality", "h", 75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := newTestModel().Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			if q := got.(Model).opts.Quality; q != tt.want {
				t.Errorf("quality = %d, want %d", q, tt.want)
			}
		})
	}
}

func TestQualityClampsAtBounds(t *testing.T) {
	m := newTestModel()
	m.opts.Quality = 98
	raised, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if q := raised.(Model).opts.Quality; q != 100 {
		t.Errorf("quality = %d, want it capped at 100", q)
	}

	m.opts.Quality = 3
	lowered, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if q := lowered.(Model).opts.Quality; q != 1 {
		t.Errorf("quality = %d, want it floored at 1", q)
	}
}

func TestScaleKeys(t *testing.T) {
	lowered, _ := newTestModel().Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if s := lowered.(Model).opts.Scale; s < 0.94 || s > 0.96 {
		t.Errorf("scale = %v, want about 0.95", s)
	}

	// Scale starts at 1 and must not go above it.
	raised, _ := newTestModel().Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if s := raised.(Model).opts.Scale; s != 1 {
		t.Errorf("scale = %v, want it capped at 1", s)
	}
}

func TestScaleFloor(t *testing.T) {
	m := newTestModel()
	m.opts.Scale = 0.06

	got, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	if s := got.(Model).opts.Scale; s < minScale-0.001 {
		t.Errorf("scale = %v, want it floored at %v", s, minScale)
	}
}

func TestCycleFormatAlsoRenamesOutput(t *testing.T) {
	m := newTestModel()

	got, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	next := got.(Model)

	if next.opts.Format == optimize.FormatJPEG {
		t.Error("format should have advanced past jpeg")
	}
	if !strings.HasSuffix(next.output, next.opts.Format.Ext()) {
		t.Errorf("output %q should carry the %s extension", next.output, next.opts.Format.Ext())
	}
}

func TestCycleFormatWrapsAround(t *testing.T) {
	m := newTestModel()

	for range len(optimize.EncodableFormats()) {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
		m = updated.(Model)
	}

	if m.opts.Format != optimize.FormatJPEG {
		t.Errorf("format = %v, want a full cycle back to jpeg", m.opts.Format)
	}
}

func TestResetRestoresDefaults(t *testing.T) {
	m := newTestModel()
	m.opts.Quality = 20
	m.opts.Scale = 0.3

	got, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})

	reset := got.(Model)
	if reset.opts.Quality != 80 || reset.opts.Scale != 1 {
		t.Errorf("after reset quality = %d, scale = %v, want 80 and 1", reset.opts.Quality, reset.opts.Scale)
	}
}

func TestAdjustingBumpsSequenceToDropStaleResults(t *testing.T) {
	m := newTestModel()
	before := m.seq

	got, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})

	if got.(Model).seq == before {
		t.Error("changing a setting should bump the sequence so in-flight encodes are discarded")
	}
}

func TestStaleOptimizedMessageIsIgnored(t *testing.T) {
	m := newTestModel()
	m.seq = 5

	got, _ := m.Update(optimizedMsg{seq: 2, result: optimize.Result{Bytes: 999}})

	if got.(Model).result != nil {
		t.Error("a result from an earlier sequence should be discarded")
	}
}

func TestOptimizedMessageStoresResult(t *testing.T) {
	m := newTestModel()

	got, _ := m.Update(optimizedMsg{seq: m.seq, result: optimize.Result{Bytes: 4321}})

	updated := got.(Model)
	if updated.result == nil || updated.result.Bytes != 4321 {
		t.Errorf("result = %v, want the message's result stored", updated.result)
	}
	if updated.encoding {
		t.Error("encoding should be finished once a result arrives")
	}
}

func TestOptimizedMessageStoresError(t *testing.T) {
	m := newTestModel()

	got, _ := m.Update(optimizedMsg{seq: m.seq, err: errors.New("boom")})

	if updated := got.(Model); updated.err == nil {
		t.Error("expected the encode error to be kept for display")
	}
}

func TestLadderMessageStoresSteps(t *testing.T) {
	m := newTestModel()

	got, _ := m.Update(ladderMsg{seq: m.seq, steps: []optimize.Step{{Quality: 80, Bytes: 100}}})

	if len(got.(Model).steps) != 1 {
		t.Error("expected the ladder steps to be stored")
	}
}

func TestLadderErrorIsNotFatal(t *testing.T) {
	m := newTestModel()

	got, _ := m.Update(ladderMsg{seq: m.seq, err: errors.New("nope")})

	if len(got.(Model).steps) != 0 {
		t.Error("a failed ladder should leave the steps empty rather than store junk")
	}
}

func TestSaveRecordsPath(t *testing.T) {
	m := newTestModel()
	m.result = &optimize.Result{Bytes: 100}

	got, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})

	if saved := got.(Model).saved; saved != "out.jpg" {
		t.Errorf("saved = %q, want out.jpg", saved)
	}
}

func TestSaveSurfacesErrors(t *testing.T) {
	deps := stubDeps()
	deps.Save = func(string, []byte) error { return errors.New("disk full") }
	m := NewModel(testSource(), optimize.Options{Quality: 80}, "out.jpg", "dev", deps)
	m.result = &optimize.Result{Bytes: 100}

	got, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})

	updated := got.(Model)
	if updated.err == nil {
		t.Error("expected the save error to surface")
	}
	if updated.saved != "" {
		t.Error("a failed save should not report a saved path")
	}
}

func TestSaveWithoutResultDoesNothing(t *testing.T) {
	got, _ := newTestModel().Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})

	if saved := got.(Model).saved; saved != "" {
		t.Errorf("saved = %q, want nothing saved before an encode finishes", saved)
	}
}

func TestQuitKeys(t *testing.T) {
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyCtrlC},
		{Type: tea.KeyEsc},
		{Type: tea.KeyRunes, Runes: []rune("q")},
	} {
		if _, cmd := newTestModel().Update(key); cmd == nil {
			t.Errorf("%v should quit", key)
		}
	}
}

func TestHelpToggles(t *testing.T) {
	m := newTestModel()

	shown, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	if !shown.(Model).showHelp {
		t.Error("? should open help")
	}

	hidden, _ := shown.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	if hidden.(Model).showHelp {
		t.Error("? should close help again")
	}
}

func TestViewBeforeSizing(t *testing.T) {
	m := NewModel(testSource(), optimize.Options{Quality: 80}, "out.jpg", "dev", stubDeps())

	if got := m.View(); !strings.Contains(got, "Loading") {
		t.Errorf("expected a loading placeholder before the first window size, got %q", got)
	}
}

func TestViewRendersPanelsAndStatus(t *testing.T) {
	m := newTestModel()
	m.result = &optimize.Result{Bytes: 5000, Format: optimize.FormatJPEG, Quality: 80, Width: 40, Height: 20}

	got := m.View()

	for _, want := range []string{"PREVIEW", "SETTINGS", "photo.jpg", "quality"} {
		if !strings.Contains(got, want) {
			t.Errorf("view missing %q", want)
		}
	}
}

func TestViewShowsHelpOverlay(t *testing.T) {
	m := newTestModel()
	m.showHelp = true

	if got := m.View(); !strings.Contains(got, "keyboard reference") {
		t.Error("help overlay should be rendered")
	}
}

func TestReplaceExt(t *testing.T) {
	tests := []struct {
		name      string
		path, ext string
		want      string
	}{
		{"swap", "photo.jpg", ".png", "photo.png"},
		{"no extension", "photo", ".png", "photo.png"},
		{"path with dots", "my.photo.jpg", ".gif", "my.photo.gif"},
		{"empty", "", ".png", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := replaceExt(tt.path, tt.ext); got != tt.want {
				t.Errorf("replaceExt(%q, %q) = %q, want %q", tt.path, tt.ext, got, tt.want)
			}
		})
	}
}
