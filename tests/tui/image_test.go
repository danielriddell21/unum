package tui_test

import (
	"bytes"
	"image"
	"image/color"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/danielriddell21/unum/internal/image/optimize"
	imagetui "github.com/danielriddell21/unum/internal/image/render/tui"
)

func imageFixture(t *testing.T) optimize.Source {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 80, 60))
	for y := range 60 {
		for x := range 80 {
			img.Set(x, y, color.RGBA{uint8(x * 3), uint8(y * 4), 90, 255})
		}
	}
	data, err := optimize.Encode(img, optimize.FormatJPEG, 90)
	if err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	src, err := optimize.Decode("fixture.jpg", data)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return src
}

func newImageModel(t *testing.T) imagetui.Model {
	t.Helper()
	deps := imagetui.DefaultDeps()
	deps.Save = func(string, []byte) error { return nil }
	return imagetui.NewModel(imageFixture(t), optimize.Options{Quality: 80}, "out.jpg", "dev", deps)
}

func TestTUIImage_RendersPanels(t *testing.T) {
	tm := teatest.NewTestModel(t, newImageModel(t), teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("PREVIEW")) && bytes.Contains(b, []byte("SETTINGS"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestTUIImage_ShowsSourceDetails(t *testing.T) {
	tm := teatest.NewTestModel(t, newImageModel(t), teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("fixture.jpg")) && bytes.Contains(b, []byte("80 × 60"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestTUIImage_EncodesAndReportsASize(t *testing.T) {
	tm := teatest.NewTestModel(t, newImageModel(t), teatest.WithInitialTermSize(200, 50))

	// The result line replaces the "encoding…" placeholder once the first
	// encode lands.
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("result"))
	}, teatest.WithDuration(10*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestTUIImage_QualityKeyChangesTheSetting(t *testing.T) {
	tm := teatest.NewTestModel(t, newImageModel(t), teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("SETTINGS"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("75"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestTUIImage_HelpOverlay(t *testing.T) {
	tm := teatest.NewTestModel(t, newImageModel(t), teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("SETTINGS"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("keyboard reference"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestTUIImage_QuitsOnEsc(t *testing.T) {
	tm := teatest.NewTestModel(t, newImageModel(t), teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("PREVIEW"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}
