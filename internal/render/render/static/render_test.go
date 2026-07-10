package static_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/render/render/static"
)

func TestBootNoColor(t *testing.T) {
	var buf bytes.Buffer
	static.Boot(&buf, "d2", "svg", static.Options{NoColor: true})
	out := buf.String()
	if !strings.Contains(out, "render") || !strings.Contains(out, "d2 → svg") {
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

func TestWrite(t *testing.T) {
	var buf bytes.Buffer
	if err := static.Write(&buf, []byte("<svg/>")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if buf.String() != "<svg/>" {
		t.Errorf("Write = %q, want %q", buf.String(), "<svg/>")
	}
}
