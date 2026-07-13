package static_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	sampleD2     = filepath.Join("testdata", "sample.d2")
	sampleMMD    = filepath.Join("testdata", "sample.mmd")
	samplePieMMD = filepath.Join("testdata", "sample-pie.mmd")
)

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

func TestRenderMermaidDrawioEditable(t *testing.T) {
	// A mermaid flowchart exports as an editable node graph, not an image cell.
	stdout, _, code := run("render", sampleMMD, "--format", "drawio", flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if !strings.Contains(stdout, `vertex="1"`) || !strings.Contains(stdout, `edge="1"`) {
		t.Errorf("expected editable vertex/edge cells in mermaid drawio:\n%s", stdout)
	}
	if strings.Contains(stdout, "shape=image") {
		t.Errorf("mermaid flowchart drawio should not be an embedded image:\n%s", stdout)
	}
}

func TestRenderMermaidPieSVG(t *testing.T) {
	// A non-flowchart mermaid type still renders natively to SVG.
	stdout, _, code := run("render", samplePieMMD, flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if !strings.Contains(stdout, "<svg") {
		t.Errorf("expected <svg for pie chart:\n%s", stdout)
	}
}

func TestRenderMermaidPieDrawioImageCell(t *testing.T) {
	// Non-flowchart types fall back to an embedded-image drawio cell.
	stdout, _, code := run("render", samplePieMMD, "--format", "drawio", flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if !strings.Contains(stdout, "shape=image") {
		t.Errorf("expected image cell for non-flowchart drawio:\n%s", stdout)
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
	// mermaid renders natively (go-mermaid) — no browser required.
	stdout, _, code := run("render", sampleMMD, flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if !strings.Contains(stdout, "<svg") {
		t.Errorf("expected <svg in mermaid output:\n%s", stdout)
	}
}

func TestRenderMermaidPNG(t *testing.T) {
	// mermaid png also renders natively (pure-Go raster) — no browser required.
	out := filepath.Join(t.TempDir(), "m.png")
	_, _, code := run("render", sampleMMD, "--format", "png", "-o", out, flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read png: %v", err)
	}
	if len(data) < 8 || string(data[1:4]) != "PNG" {
		t.Errorf("output is not a PNG")
	}
}

func TestRenderD2PNG(t *testing.T) {
	// d2 png rasterises in pure Go (resvg) — no browser required.
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
