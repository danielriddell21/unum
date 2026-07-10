package static_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	sampleD2  = filepath.Join("testdata", "sample.d2")
	sampleMMD = filepath.Join("testdata", "sample.mmd")
)

func chromiumAvailable() bool {
	if bin := os.Getenv("UNUM_CHROMIUM_BIN"); bin != "" {
		if _, err := os.Stat(bin); err == nil {
			return true
		}
	}
	for _, name := range []string{"chromium", "chromium-browser", "google-chrome", "google-chrome-stable", "chrome"} {
		if _, err := exec.LookPath(name); err == nil {
			return true
		}
	}
	return false
}

func TestRenderD2SVG(t *testing.T) {
	stdout, _, code := run("render", sampleD2, flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if !strings.Contains(stdout, "<svg") {
		t.Errorf("expected <svg in output:\n%s", stdout)
	}
}

func TestRenderD2Drawio(t *testing.T) {
	stdout, _, code := run("render", sampleD2, "--format", "drawio", flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if !strings.Contains(stdout, "mxGraphModel") {
		t.Errorf("expected mxGraphModel in drawio output:\n%s", stdout)
	}
}

func TestRenderNoColor(t *testing.T) {
	stdout, stderr, code := run("render", sampleD2, flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if strings.Contains(stdout, "\x1b[") || strings.Contains(stderr, "\x1b[") {
		t.Error("--no-color output contains ANSI escape codes")
	}
}

func TestRenderOutputFile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.svg")
	_, _, code := run("render", sampleD2, "-o", out, flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !strings.Contains(string(data), "<svg") {
		t.Errorf("output file missing <svg:\n%s", data)
	}
}

func TestRenderUnknownLanguage(t *testing.T) {
	_, stderr, code := run("render", filepath.Join("testdata", "sample.json"), flagNoColor)
	if code == 0 {
		t.Fatal("expected non-zero exit for unknown diagram language")
	}
	if stderr == "" {
		t.Error("expected error message in stderr")
	}
}

func TestRenderBadFormat(t *testing.T) {
	_, stderr, code := run("render", sampleD2, "--format", "pdf")
	if code == 0 {
		t.Fatal("expected non-zero exit for unknown format")
	}
	if stderr == "" {
		t.Error("expected error message in stderr")
	}
}

func TestRenderMissingFile(t *testing.T) {
	_, stderr, code := run("render", "nonexistent.d2")
	if code == 0 {
		t.Fatal("expected non-zero exit for missing file")
	}
	if stderr == "" {
		t.Error("expected error message in stderr")
	}
}

func TestRenderMermaidSVG(t *testing.T) {
	if !chromiumAvailable() {
		t.Skip("no chromium available; skipping mermaid render")
	}
	stdout, _, code := run("render", sampleMMD, flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if !strings.Contains(stdout, "<svg") {
		t.Errorf("expected <svg in mermaid output:\n%s", stdout)
	}
}

func TestRenderD2PNG(t *testing.T) {
	if !chromiumAvailable() {
		t.Skip("no chromium available; skipping png render")
	}
	out := filepath.Join(t.TempDir(), "out.png")
	_, _, code := run("render", sampleD2, "--format", "png", "-o", out, flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read png: %v", err)
	}
	if len(data) < 8 || string(data[1:4]) != "PNG" {
		t.Errorf("output is not a PNG (first bytes: %v)", data[:min(8, len(data))])
	}
}
