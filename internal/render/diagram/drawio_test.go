package diagram_test

import (
	"bytes"
	"testing"

	"github.com/danielriddell21/unum/internal/render/diagram"
)

func TestDrawioFromSVG(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="320" height="240"></svg>`)
	out, err := diagram.DrawioFromSVG(svg)
	if err != nil {
		t.Fatalf("DrawioFromSVG: %v", err)
	}
	for _, want := range []string{"<mxGraphModel", "shape=image", "data:image/svg+xml;base64,", `width="320"`, `height="240"`} {
		if !bytes.Contains(out, []byte(want)) {
			t.Errorf("drawio output missing %q:\n%s", want, out)
		}
	}
}

func TestMermaidDrawioFlowchartEditable(t *testing.T) {
	src := "flowchart TD\n    A[Client] --> B(API Server)\n"
	out, err := diagram.MermaidDrawio(src, nil, nil)
	if err != nil {
		t.Fatalf("MermaidDrawio: %v", err)
	}
	for _, want := range []string{"<mxGraphModel", `vertex="1"`, `edge="1"`, "Client", "API Server"} {
		if !bytes.Contains(out, []byte(want)) {
			t.Errorf("editable flowchart drawio missing %q:\n%s", want, out)
		}
	}
	if bytes.Contains(out, []byte("shape=image")) {
		t.Errorf("flowchart drawio should be an editable node graph, not an image cell:\n%s", out)
	}
}

func TestMermaidDrawioNonFlowchartImageCell(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="320" height="240"></svg>`)
	out, err := diagram.MermaidDrawio("sequenceDiagram\n    Alice->>Bob: hi\n", svg, nil)
	if err != nil {
		t.Fatalf("MermaidDrawio: %v", err)
	}
	if !bytes.Contains(out, []byte("shape=image")) {
		t.Errorf("non-flowchart drawio should fall back to an image cell:\n%s", out)
	}
}

func TestDrawioFromSVGDefaultSize(t *testing.T) {
	out, err := diagram.DrawioFromSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`))
	if err != nil {
		t.Fatalf("DrawioFromSVG: %v", err)
	}
	if !bytes.Contains(out, []byte(`width="640"`)) {
		t.Errorf("expected default width 640 when svg has no size:\n%s", out)
	}
}
