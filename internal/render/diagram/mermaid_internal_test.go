package diagram

import (
	"strings"
	"testing"
)

func TestMermaidToD2(t *testing.T) {
	src := "flowchart TD\n    A[Client] -->|go| B(API Server)\n    B --> A\n"
	out, err := MermaidToD2(src)
	if err != nil {
		t.Fatalf("MermaidToD2: %v", err)
	}
	for _, want := range []string{`A: "Client"`, `B: "API Server"`, `A -> B: "go"`, "B -> A\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("MermaidToD2 missing %q in:\n%s", want, out)
		}
	}
}

func TestMermaidToD2NonFlowchart(t *testing.T) {
	// A sequence diagram is not a node/edge flowchart; expect empty (fall back).
	out, _ := MermaidToD2("sequenceDiagram\n    Alice->>Bob: hi\n")
	if out != "" {
		t.Errorf("expected empty conversion for non-flowchart, got:\n%s", out)
	}
}

func TestD2ID(t *testing.T) {
	if got := d2ID("Client_1"); got != "Client_1" {
		t.Errorf("d2ID(Client_1) = %q, want unquoted", got)
	}
	if got := d2ID("has space"); got != `"has space"` {
		t.Errorf("d2ID(has space) = %q, want quoted", got)
	}
}
