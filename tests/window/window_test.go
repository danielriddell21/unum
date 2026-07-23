package window_test

import (
	"path/filepath"
	"strings"
	"testing"
)

var sample = filepath.Join("testdata", "sample.json")

// The functional binary is built without -tags ebiten, so `window` reports
// the stub error; the windowed happy path needs a display and OpenGL.

func TestWindowUnavailableWithoutTag(t *testing.T) {
	_, stderr, code := run("window", sample)
	if code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	if !strings.Contains(stderr, "-tags ebiten") {
		t.Errorf("expected rebuild hint in stderr:\n%s", stderr)
	}
}

func TestWindowNoColor(t *testing.T) {
	_, stderr, code := run("window", sample, "--no-color")
	if code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	if !strings.Contains(stderr, "-tags ebiten") {
		t.Errorf("expected rebuild hint in stderr:\n%s", stderr)
	}
}

func TestWindowHelp(t *testing.T) {
	stdout, _, code := run("window", "--help")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	for _, tool := range []string{"json", "diff", "hash"} {
		if !strings.Contains(stdout, tool) {
			t.Errorf("help should mention %q:\n%s", tool, stdout)
		}
	}
}

func TestWindowTooManyArgs(t *testing.T) {
	_, _, code := run("window", sample, sample, sample)
	if code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
}
