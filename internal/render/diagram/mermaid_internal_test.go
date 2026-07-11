package diagram

import "testing"

func TestGraphToD2(t *testing.T) {
	g := mermaidGraph{
		Nodes: []struct{ ID, Label string }{{"Client", "Client"}, {"API", "API Server"}},
		Edges: []struct{ Src, Dst, Label string }{{"Client", "API", "request"}, {"API", "Client", ""}},
	}
	out := graphToD2(g)
	for _, want := range []string{`Client: "Client"`, `API: "API Server"`, `Client -> API: "request"`, "API -> Client\n"} {
		if !contains(out, want) {
			t.Errorf("graphToD2 missing %q in:\n%s", want, out)
		}
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

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
