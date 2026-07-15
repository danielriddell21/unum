package diagram_test

import (
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/render/diagram"
)

func TestMermaidToD2Flowchart(t *testing.T) {
	src := "flowchart TD\n    A[Client] -->|go| B(API Server)\n    B --> A\n"
	out, err := diagram.MermaidToD2(src)
	if err != nil {
		t.Fatalf("MermaidToD2: %v", err)
	}
	for _, want := range []string{`A: "Client"`, `B: "API Server"`, `A -> B: "go"`, "B -> A\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("MermaidToD2 missing %q in:\n%s", want, out)
		}
	}
}

func TestMermaidToD2Sequence(t *testing.T) {
	src := "sequenceDiagram\n    participant A as Alice\n    A->>B: hi\n"
	out, err := diagram.MermaidToD2(src)
	if err != nil {
		t.Fatalf("MermaidToD2: %v", err)
	}
	for _, want := range []string{"shape: sequence_diagram", `A: "Alice"`, `A -> B: "hi"`} {
		if !strings.Contains(out, want) {
			t.Errorf("sequence conversion missing %q in:\n%s", want, out)
		}
	}
}

func TestMermaidToD2State(t *testing.T) {
	out, err := diagram.MermaidToD2("stateDiagram-v2\n    [*] --> Idle\n    Idle --> Run : go\n")
	if err != nil {
		t.Fatalf("MermaidToD2: %v", err)
	}
	for _, want := range []string{"start ->", "Idle -> Run", `"go"`} {
		if !strings.Contains(out, want) {
			t.Errorf("state conversion missing %q in:\n%s", want, out)
		}
	}
}

func TestMermaidToD2ER(t *testing.T) {
	out, err := diagram.MermaidToD2("erDiagram\n    CUSTOMER ||--o{ ORDER : places\n")
	if err != nil {
		t.Fatalf("MermaidToD2: %v", err)
	}
	if !strings.Contains(out, `CUSTOMER -> ORDER: "places"`) {
		t.Errorf("er conversion missing relationship in:\n%s", out)
	}
}

func TestMermaidToD2Unsupported(t *testing.T) {
	// Chart/timeline types have no d2 equivalent; expect empty (note in TUI).
	out, _ := diagram.MermaidToD2("pie title T\n    \"A\" : 1\n")
	if out != "" {
		t.Errorf("expected empty conversion for a pie chart, got:\n%s", out)
	}
}
