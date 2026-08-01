package engine_test

import (
	"bytes"
	"testing"

	"github.com/danielriddell21/unum/internal/diagram/engine"
)

const (
	d2Src  = "a -> b\nb -> c\n"
	mmdSrc = "flowchart TD\n A --> B\n"
)

func TestRenderDispatch(t *testing.T) {
	cases := []struct {
		lang, format, wantContentType, wantContains string
	}{
		{"d2", "svg", "image/svg+xml", "<svg"},
		{"d2", "drawio", "application/xml", "mxGraphModel"},
		{"mermaid", "svg", "image/svg+xml", "<svg"},
		{"mermaid", "drawio", "application/xml", "mxGraphModel"},
		{"d2", "", "image/svg+xml", "<svg"}, // default format
	}
	for _, tc := range cases {
		src := d2Src
		if tc.lang == "mermaid" {
			src = mmdSrc
		}
		out, ct, err := engine.Render(tc.lang, tc.format, src, engine.ThemeByName("cyber"))
		if err != nil {
			t.Fatalf("Render(%s,%s): %v", tc.lang, tc.format, err)
		}
		if ct != tc.wantContentType {
			t.Errorf("Render(%s,%s) content-type = %q, want %q", tc.lang, tc.format, ct, tc.wantContentType)
		}
		if !bytes.Contains(out, []byte(tc.wantContains)) {
			t.Errorf("Render(%s,%s) missing %q", tc.lang, tc.format, tc.wantContains)
		}
	}
}

func TestRenderPNG(t *testing.T) {
	for _, lang := range []string{"d2", "mermaid"} {
		src := d2Src
		if lang == "mermaid" {
			src = mmdSrc
		}
		out, ct, err := engine.Render(lang, "png", src, engine.ThemeByName("cyber"))
		if err != nil {
			t.Fatalf("Render(%s,png): %v", lang, err)
		}
		if ct != "image/png" {
			t.Errorf("Render(%s,png) content-type = %q, want image/png", lang, ct)
		}
		if len(out) < 8 || string(out[1:4]) != "PNG" {
			t.Errorf("Render(%s,png) is not a PNG", lang)
		}
	}
}

func TestRenderUnknownLang(t *testing.T) {
	if _, _, err := engine.Render("cobol", "svg", "x", nil); err == nil {
		t.Error("expected an error for an unknown language")
	}
}

func TestSVGSize(t *testing.T) {
	if w, h := engine.SVGSize([]byte(`<svg width="320" height="240">`)); w != 320 || h != 240 {
		t.Errorf("SVGSize width/height = %d/%d, want 320/240", w, h)
	}
	if w, h := engine.SVGSize([]byte(`<svg viewBox="0 0 100 50">`)); w != 100 || h != 50 {
		t.Errorf("SVGSize from viewBox = %d/%d, want 100/50", w, h)
	}
	if w, h := engine.SVGSize([]byte("<svg>")); w != 0 || h != 0 {
		t.Errorf("SVGSize with no dimensions = %d/%d, want 0/0", w, h)
	}
}

func TestD2Counts(t *testing.T) {
	d, err := engine.RenderD2(d2Src, engine.ThemeByName("cyber"))
	if err != nil {
		t.Fatalf("RenderD2: %v", err)
	}
	if d.NumShapes() != 3 {
		t.Errorf("NumShapes = %d, want 3", d.NumShapes())
	}
	if d.NumConnections() != 2 {
		t.Errorf("NumConnections = %d, want 2", d.NumConnections())
	}
}

func TestRenderD2ASCII(t *testing.T) {
	out, err := engine.RenderD2ASCII(d2Src)
	if err != nil {
		t.Fatalf("RenderD2ASCII: %v", err)
	}
	if out == "" {
		t.Error("RenderD2ASCII produced no output")
	}
}

func TestSVGToPNG(t *testing.T) {
	d, err := engine.RenderD2(d2Src, engine.ThemeByName("cyber"))
	if err != nil {
		t.Fatalf("RenderD2: %v", err)
	}
	png, err := engine.SVGToPNG(d.SVG)
	if err != nil {
		t.Fatalf("SVGToPNG: %v", err)
	}
	if len(png) < 8 || string(png[1:4]) != "PNG" {
		t.Error("SVGToPNG did not produce a PNG")
	}
}

func TestMermaidSVGToPNG(t *testing.T) {
	svg, err := engine.RenderMermaid(mmdSrc, engine.ThemeByName("cyber"))
	if err != nil {
		t.Fatalf("RenderMermaid: %v", err)
	}
	png, err := engine.MermaidSVGToPNG(svg)
	if err != nil {
		t.Fatalf("MermaidSVGToPNG: %v", err)
	}
	if len(png) < 8 || string(png[1:4]) != "PNG" {
		t.Error("MermaidSVGToPNG did not produce a PNG")
	}
}
