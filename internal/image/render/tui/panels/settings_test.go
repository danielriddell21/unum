package panels

import (
	"errors"
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/image/optimize"
)

func testView() SettingsView {
	return SettingsView{
		Source: optimize.Source{
			Name:   "photo.jpg",
			Format: optimize.FormatJPEG,
			Width:  1200,
			Height: 800,
			Bytes:  118000,
		},
		Format:  optimize.FormatJPEG,
		Quality: 80,
		Scale:   1,
		Width:   1200,
		Height:  800,
	}
}

func TestSettingsPanelShowsSourceAndSettings(t *testing.T) {
	p := NewSettingsPanel(40, 20)

	got := p.View(testView())

	for _, want := range []string{"photo.jpg", "jpeg", "1200 × 800", "118 kB", "80", "100%"} {
		if !strings.Contains(got, want) {
			t.Errorf("settings panel missing %q:\n%s", want, got)
		}
	}
}

func TestSettingsPanelResultStates(t *testing.T) {
	p := NewSettingsPanel(40, 20)

	t.Run("encoding", func(t *testing.T) {
		v := testView()
		v.Encoding = true
		if got := p.View(v); !strings.Contains(got, "encoding") {
			t.Errorf("expected an encoding note:\n%s", got)
		}
	})

	t.Run("error", func(t *testing.T) {
		v := testView()
		v.Err = errors.New("cannot write webp")
		if got := p.View(v); !strings.Contains(got, "cannot write webp") {
			t.Errorf("expected the error to surface:\n%s", got)
		}
	})

	t.Run("result", func(t *testing.T) {
		v := testView()
		v.Result = &optimize.Result{Bytes: 30000}
		got := p.View(v)
		if !strings.Contains(got, "30 kB") {
			t.Errorf("expected the result size:\n%s", got)
		}
		if !strings.Contains(got, "-75%") {
			t.Errorf("expected the saving:\n%s", got)
		}
	})

	t.Run("result that grew", func(t *testing.T) {
		v := testView()
		v.Result = &optimize.Result{Bytes: 150000}
		if got := p.View(v); !strings.Contains(got, "+27%") {
			t.Errorf("expected a positive percentage when the file grew:\n%s", got)
		}
	})
}

func TestSettingsPanelLadder(t *testing.T) {
	p := NewSettingsPanel(40, 20)
	v := testView()
	v.Steps = []optimize.Step{{Quality: 90, Bytes: 100000}, {Quality: 80, Bytes: 50000}}

	got := p.View(v)

	for _, want := range []string{"quality", "90", "80", "100 kB", "50 kB"} {
		if !strings.Contains(got, want) {
			t.Errorf("ladder missing %q:\n%s", want, got)
		}
	}
}

func TestSettingsPanelShowsSavedPath(t *testing.T) {
	p := NewSettingsPanel(40, 20)
	v := testView()
	v.Saved = "photo-small.jpg"

	if got := p.View(v); !strings.Contains(got, "photo-small.jpg") {
		t.Errorf("expected the saved path:\n%s", got)
	}
}

func TestSettingsPanelResize(t *testing.T) {
	p := NewSettingsPanel(10, 10)
	p.Resize(60, 30)

	// The separator rules follow the panel width.
	if got := p.View(testView()); !strings.Contains(got, strings.Repeat("─", 60)) {
		t.Error("separator should span the resized panel width")
	}
}
