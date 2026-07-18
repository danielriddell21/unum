package static_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/diagram/render/static"
)

func TestBootNoColor(t *testing.T) {
	var buf bytes.Buffer
	static.Boot(&buf, "d2", "svg", static.Options{NoColor: true})
	out := buf.String()
	if !strings.Contains(out, "diagram") || !strings.Contains(out, "d2 → svg") {
		t.Errorf("unexpected boot line: %q", out)
	}
	if strings.Contains(out, "\x1b[") {
		t.Errorf("no-color boot line contains ANSI escapes: %q", out)
	}
}

func TestBootQuiet(t *testing.T) {
	var buf bytes.Buffer
	static.Boot(&buf, "d2", "svg", static.Options{Quiet: true})
	if buf.Len() != 0 {
		t.Errorf("quiet boot should write nothing, got %q", buf.String())
	}
}

func TestBootColored(t *testing.T) {
	// Exercises the styled (non --no-color) path. lipgloss strips ANSI when the
	// writer is not a TTY, so assert on content rather than escape codes.
	var buf bytes.Buffer
	static.Boot(&buf, "mermaid", "png", static.Options{Theme: static.ResolveTheme("cyber")})
	out := buf.String()
	if !strings.Contains(out, "diagram") || !strings.Contains(out, "mermaid → png") {
		t.Errorf("unexpected boot line: %q", out)
	}
}

func TestResolveTheme(t *testing.T) {
	// A known and an unknown theme both resolve to a usable banner style.
	for _, name := range []string{"cyber", "does-not-exist"} {
		if got := static.ResolveTheme(name); got.Banner.Render("x") == "" {
			t.Errorf("ResolveTheme(%q) produced an empty banner style", name)
		}
	}
}

func TestWrite(t *testing.T) {
	var buf bytes.Buffer
	if err := static.Write(&buf, []byte("<svg/>")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if buf.String() != "<svg/>" {
		t.Errorf("Write = %q, want %q", buf.String(), "<svg/>")
	}
}
