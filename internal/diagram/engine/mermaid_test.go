package engine_test

import (
	"bytes"
	"testing"

	"github.com/danielriddell21/unum/internal/diagram/engine"
)

// Every diagram kind go-mermaid v0.1.3 recognises must render to an SVG through
// RenderMermaid. Samples are the library's own golden fixtures, minimised.
func TestRenderMermaidAllKinds(t *testing.T) {
	cases := []struct {
		kind string
		src  string
	}{
		{"flowchart", "graph TD\nA[Start] --> B{Decision}\nB -->|yes| C([Done])\nB -->|no| D((Retry))\nD --> A\n"},
		{"sequence", "sequenceDiagram\nparticipant A as Alice\nparticipant B as Bob\nA->>B: Request\nB-->>A: Response\n"},
		{"class", "classDiagram\nclass Animal {\n+int age\n+isMammal() bool\n}\nclass Dog\nAnimal <|-- Dog\n"},
		{"state", "stateDiagram-v2\n[*] --> Still\nStill --> Moving : start\nMoving --> Still : stop\nMoving --> [*]\n"},
		{"er", "erDiagram\nCUSTOMER ||--o{ ORDER : places\nORDER ||--|{ LINE_ITEM : contains\n"},
		{"pie", "pie title Pets adopted\n\"Dogs\" : 386\n\"Cats\" : 85\n\"Rats\" : 15\n"},
		{"journey", "journey\ntitle My working day\nsection Go to work\nMake tea: 5: Me\nDo work: 1: Me, Cat\n"},
		{"quadrant", "quadrantChart\ntitle Reach and engagement\nx-axis Low Reach --> High Reach\ny-axis Low Engagement --> High Engagement\nCampaign A: [0.3, 0.6]\nCampaign B: [0.45, 0.23]\n"},
		{"gitgraph", "gitGraph\ncommit\ncommit id: \"init\"\nbranch develop\ncommit\ncheckout main\nmerge develop tag: \"v1.0\"\n"},
		{"timeline", "timeline\ntitle History of Social Media\nsection Pioneers\n2002 : LinkedIn\n2004 : Facebook : Google\n"},
		{"mindmap", "mindmap\n  root((Mindmap))\n    Origins\n      Long history\n    Research\n      On effectiveness\n"},
		{"gantt", "gantt\ntitle A Gantt Diagram\ndateFormat YYYY-MM-DD\nsection Design\nResearch : a1, 2024-01-01, 10d\nMockups : after a1, 7d\n"},
		{"c4", "C4Context\ntitle System Context diagram\nPerson(customerA, \"Banking Customer\", \"A customer\")\nSystem(systemAA, \"Internet Banking System\", \"Allows viewing\")\nRel(customerA, systemAA, \"Uses\")\n"},
		{"requirement", "requirementDiagram\nrequirement test_req {\nid: 1\ntext: the test text\nrisk: high\nverifymethod: test\n}\nelement test_entity {\ntype: simulation\n}\ntest_entity - satisfies -> test_req\n"},
		{"sankey", "sankey-beta\nAgricultural,Bio-conversion,124\nBio-conversion,Liquid,32\nBio-conversion,Solid,66\nSolid,Electricity,40\n"},
		{"xychart", "xychart-beta\ntitle \"Sales Revenue\"\nx-axis [jan, feb, mar, apr, may]\ny-axis \"Revenue (USD)\" 4000 --> 11000\nbar [5000, 6000, 7500, 8200, 9500]\nline [5000, 6000, 7500, 8200, 9500]\n"},
		{"block", "block-beta\ncolumns 3\na[\"API\"] b[\"Worker\"] c[\"DB\"]\nd[\"Load Balancer\"]:3\ne f g\n"},
		{"kanban", "kanban\n  Backlog\n    [Research spike]\n  In Progress\n    [Implement parser]\n  Done\n    [Setup CI]\n"},
		{"packet", "packet-beta\n0-15: \"Source Port\"\n16-31: \"Destination Port\"\n32-63: \"Sequence Number\"\n"},
		{"radar", "radar-beta\ntitle Team Skills\naxis speed[\"Speed\"], power[\"Power\"], range[\"Range\"]\ncurve a[\"Team A\"]{80, 60, 90}\ncurve b[\"Team B\"]{50, 90, 40}\n"},
	}

	if len(cases) != 20 {
		t.Fatalf("expected 20 diagram kinds, have %d", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			out, err := engine.RenderMermaid(tc.src, nil)
			if err != nil {
				t.Fatalf("RenderMermaid(%s): %v", tc.kind, err)
			}
			if !bytes.Contains(out, []byte("<svg")) {
				t.Errorf("RenderMermaid(%s) produced no <svg:\n%s", tc.kind, out)
			}
		})
	}
}

func TestRenderMermaidUnsupported(t *testing.T) {
	// An unrecognised header must surface a clean error, never a panic.
	if _, err := engine.RenderMermaid("bogusDiagram\n  x --> y\n", nil); err == nil {
		t.Error("expected an error for an unrecognised diagram type")
	}
}
