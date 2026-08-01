package tui_test

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	rendertui "github.com/danielriddell21/unum/internal/diagram/render/tui"
)

func stubInfo() rendertui.Info {
	return rendertui.Info{
		File:   "sample.d2",
		Lang:   "d2",
		Source: "client -> api\napi -> db",
		ASCII:  "┌────────┐\n│ client │\n└────────┘",
		SVG:    []byte("<svg/>"),
		PNG:    []byte("\x89PNG\r\n"),
		Drawio: []byte("<mxGraphModel/>"),
		Width:  155,
		Height: 458,
		Shapes: 3,
		Conns:  2,
	}
}

func TestTUIRender_ShowsSourceAndRenderPanels(t *testing.T) {
	m := rendertui.NewModel("dev", stubInfo())
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("SOURCE")) && bytes.Contains(b, []byte("RENDER"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestTUIRender_HelpOverlay(t *testing.T) {
	m := rendertui.NewModel("dev", stubInfo())
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("RENDER"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("keyboard reference"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}
