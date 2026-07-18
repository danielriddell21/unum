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
	for _, want := range []string{"__start__ ->", "Idle -> Run", `"go"`} {
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

func TestMermaidToD2ReservedIDs(t *testing.T) {
	// Node IDs that are d2 reserved keywords must be quoted so the generated
	// source still compiles.
	out, err := diagram.MermaidToD2("flowchart TD\n    style[Styling] --> shape[Shapes]\n")
	if err != nil {
		t.Fatalf("MermaidToD2: %v", err)
	}
	if _, err := diagram.RenderD2(out, nil); err != nil {
		t.Errorf("generated d2 with reserved-keyword IDs does not compile: %v\n%s", err, out)
	}
}

func TestMermaidToD2StateNamedStart(t *testing.T) {
	// A user state named "start" must stay distinct from the synthetic
	// start terminal.
	out, err := diagram.MermaidToD2("stateDiagram-v2\n    [*] --> start\n    start --> done\n")
	if err != nil {
		t.Fatalf("MermaidToD2: %v", err)
	}
	for _, want := range []string{`__start__: "●"`, "__start__ -> start", "start -> done"} {
		if !strings.Contains(out, want) {
			t.Errorf("state conversion missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, `start: "●"`+"\n") && !strings.Contains(out, `__start__: "●"`) {
		t.Errorf("user state was merged with the synthetic terminal:\n%s", out)
	}
	if _, err := diagram.RenderD2(out, nil); err != nil {
		t.Errorf("generated d2 does not compile: %v\n%s", err, out)
	}
}

func TestMermaidToD2Unsupported(t *testing.T) {
	// Chart/timeline types have no d2 equivalent; expect empty (note in TUI).
	out, _ := diagram.MermaidToD2("pie title T\n    \"A\" : 1\n")
	if out != "" {
		t.Errorf("expected empty conversion for a pie chart, got:\n%s", out)
	}
}
