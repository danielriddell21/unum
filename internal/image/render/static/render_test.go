package static

import (
	"bytes"
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/image/optimize"
)

func testOptions() Options {
	return Options{Theme: ResolveTheme("cyber"), NoColor: true}
}

func testSource() optimize.Source {
	return optimize.Source{
		Name:   "photo.jpg",
		Format: optimize.FormatJPEG,
		Width:  1200,
		Height: 800,
		Bytes:  118000,
	}
}

func TestBoot(t *testing.T) {
	tests := []struct {
		name   string
		opts   Options
		want   string
		absent bool
	}{
		{"no-color", Options{NoColor: true}, "[ UNUM ] image  photo.jpg", false},
		{"quiet prints nothing", Options{Quiet: true}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			Boot(&buf, "photo.jpg", tt.opts)
			if tt.absent {
				if buf.Len() != 0 {
					t.Errorf("quiet should print nothing, got %q", buf.String())
				}
				return
			}
			if !strings.Contains(buf.String(), tt.want) {
				t.Errorf("Boot() = %q, want it to contain %q", buf.String(), tt.want)
			}
		})
	}
}

func TestBootNoColorHasNoEscapeCodes(t *testing.T) {
	var buf bytes.Buffer
	Boot(&buf, "photo.jpg", Options{NoColor: true})

	if strings.Contains(buf.String(), "\x1b[") {
		t.Errorf("no-color output contains ANSI escapes: %q", buf.String())
	}
}

func TestRenderReport(t *testing.T) {
	var buf bytes.Buffer
	RenderReport(&buf, Report{
		Source: testSource(),
		Format: optimize.FormatJPEG,
		Width:  600,
		Height: 400,
		Steps: []optimize.Step{
			{Quality: 90, Bytes: 100000},
			{Quality: 80, Bytes: 50000},
		},
		Active: 80,
	}, testOptions())

	out := buf.String()
	for _, want := range []string{
		"photo.jpg", "jpeg", "1200 × 800", "118 kB",
		"600 × 400", "quality", "saving", "90", "80", "← default",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing %q:\n%s", want, out)
		}
	}
}

func TestRenderReportShowsSavingSign(t *testing.T) {
	var buf bytes.Buffer
	RenderReport(&buf, Report{
		Source: testSource(),
		Format: optimize.FormatJPEG,
		Width:  1200,
		Height: 800,
		Steps: []optimize.Step{
			{Quality: 95, Bytes: 150000}, // larger than the source
			{Quality: 40, Bytes: 20000},  // much smaller
		},
	}, testOptions())

	out := buf.String()
	if !strings.Contains(out, "+27%") {
		t.Errorf("expected a positive percentage for a file that grew:\n%s", out)
	}
	if !strings.Contains(out, "-83%") {
		t.Errorf("expected a negative percentage for a file that shrank:\n%s", out)
	}
}

func TestRenderReportWithoutSteps(t *testing.T) {
	var buf bytes.Buffer
	RenderReport(&buf, Report{Source: testSource(), Format: optimize.FormatPNG, Width: 10, Height: 10}, testOptions())

	if strings.Contains(buf.String(), "quality") {
		t.Errorf("no ladder should be rendered when there are no steps:\n%s", buf.String())
	}
}

func TestRenderSummary(t *testing.T) {
	var buf bytes.Buffer
	RenderSummary(&buf, Summary{
		Source: testSource(),
		Result: optimize.Result{Format: optimize.FormatJPEG, Quality: 70, Width: 600, Height: 400, Bytes: 30000},
		Output: "small.jpg",
	}, testOptions())

	out := buf.String()
	for _, want := range []string{"photo.jpg", "118 kB", "small.jpg", "30 kB", "-75%", "jpeg q70", "600 × 400"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary missing %q:\n%s", want, out)
		}
	}
}

func TestRenderNote(t *testing.T) {
	var buf bytes.Buffer
	RenderNote(&buf, "try --scale", testOptions())

	if !strings.Contains(buf.String(), "try --scale") {
		t.Errorf("note missing from output: %q", buf.String())
	}
}

// The colored and no-color paths are separate format strings, so they can
// drift apart. Both must report the same facts.
func TestColorAndNoColorCarryTheSameFacts(t *testing.T) {
	sum := Summary{
		Source: testSource(),
		Result: optimize.Result{Format: optimize.FormatJPEG, Quality: 70, Width: 600, Height: 400, Bytes: 30000},
		Output: "small.jpg",
	}

	var plain, colored bytes.Buffer
	RenderSummary(&plain, sum, Options{Theme: ResolveTheme("cyber"), NoColor: true})
	RenderSummary(&colored, sum, Options{Theme: ResolveTheme("cyber")})

	for _, want := range []string{"photo.jpg", "118 kB", "small.jpg", "30 kB", "-75%", "jpeg q70"} {
		if !strings.Contains(plain.String(), want) {
			t.Errorf("no-color summary missing %q:\n%s", want, plain.String())
		}
		if !strings.Contains(colored.String(), want) {
			t.Errorf("colored summary missing %q:\n%s", want, colored.String())
		}
	}
}

func TestWrite(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, []byte{1, 2, 3}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), []byte{1, 2, 3}) {
		t.Errorf("Write wrote %v, want [1 2 3]", buf.Bytes())
	}
}

func TestResolveTheme(t *testing.T) {
	for _, name := range []string{"cyber", "matrix", "dracula", "nord", "clean", "solarized"} {
		if got := ResolveTheme(name); got.Banner.GetForeground() == nil {
			t.Errorf("theme %q resolved without a banner color", name)
		}
	}

	// Distinct themes must not collapse onto the same palette.
	if ResolveTheme("cyber").Banner.GetForeground() == ResolveTheme("matrix").Banner.GetForeground() {
		t.Error("cyber and matrix resolved to the same banner color")
	}

	// An unknown name still has to yield something usable.
	if ResolveTheme("nonsense").Banner.GetForeground() == nil {
		t.Error("an unknown theme name should fall back to a usable palette")
	}
}
