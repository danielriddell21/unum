package diagram_test

import (
	"bytes"
	"testing"

	"github.com/danielriddell21/unum/internal/render/diagram"
)

func TestRenderD2SVG(t *testing.T) {
	d, err := diagram.RenderD2("x -> y")
	if err != nil {
		t.Fatalf("RenderD2: %v", err)
	}
	if !bytes.Contains(d.SVG, []byte("<svg")) {
		t.Errorf("expected <svg in output, got:\n%s", d.SVG)
	}
}

func TestRenderD2Invalid(t *testing.T) {
	if _, err := diagram.RenderD2("x -> "); err == nil {
		t.Error("expected error for invalid d2 source")
	}
}

func TestD2Drawio(t *testing.T) {
	d, err := diagram.RenderD2("alpha -> beta")
	if err != nil {
		t.Fatalf("RenderD2: %v", err)
	}
	out, err := d.Drawio()
	if err != nil {
		t.Fatalf("Drawio: %v", err)
	}
	for _, want := range []string{"<mxGraphModel", `vertex="1"`, `edge="1"`, "alpha", "beta"} {
		if !bytes.Contains(out, []byte(want)) {
			t.Errorf("drawio output missing %q:\n%s", want, out)
		}
	}
}
