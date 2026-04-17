package panels

import (
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/json/node"
	"github.com/danielriddell21/unum/internal/json/parse"
)

func mustParseStatusbar(t *testing.T, src string) *node.Node {
	t.Helper()
	n, err := parse.Parse([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return n
}

func TestStatusBar_NilNode(t *testing.T) {
	out := StatusBar(nil, 80, false, "", "", "v1")
	if !strings.Contains(out, "no selection") {
		t.Errorf("nil node: want 'no selection' in %q", out)
	}
}

func TestStatusBar_SearchMode(t *testing.T) {
	out := StatusBar(nil, 80, true, "foo", "", "v1")
	if !strings.Contains(out, "foo") {
		t.Errorf("search mode: want 'foo' in %q", out)
	}
}

func TestStatusBar_WithNode(t *testing.T) {
	n := mustParseStatusbar(t, `{"a": 1}`)
	out := StatusBar(n, 80, false, "", "", "v1")
	if !strings.Contains(out, "OBJECT") {
		t.Errorf("object node: want 'OBJECT' in %q", out)
	}
}

func TestStatusBar_ArrayNode(t *testing.T) {
	n := mustParseStatusbar(t, `[1, 2, 3]`)
	out := StatusBar(n, 80, false, "", "", "v1")
	if !strings.Contains(out, "ARRAY") {
		t.Errorf("array node: want 'ARRAY' in %q", out)
	}
	if !strings.Contains(out, "3 items") {
		t.Errorf("array node: want '3 items' in %q", out)
	}
}

func TestStatusBar_YankFeedback(t *testing.T) {
	n := mustParseStatusbar(t, `{"k": 1}`)
	out := StatusBar(n, 80, false, "", "path yanked!", "v1")
	if !strings.Contains(out, "yanked") {
		t.Errorf("yank feedback: want 'yanked' in %q", out)
	}
}
