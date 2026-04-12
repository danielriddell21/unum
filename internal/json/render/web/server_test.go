package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/json/node"
	"github.com/danielriddell21/unum/internal/json/parse"
	"github.com/danielriddell21/unum/internal/web/shared"
)

func mustParse(t *testing.T, src string) *node.Node {
	t.Helper()
	root, err := parse.Parse([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return root
}

func TestBuildPayload_Filename(t *testing.T) {
	root := mustParse(t, `{"key": "value"}`)
	p, err := buildPayload(root, "sample.json")
	if err != nil {
		t.Fatal(err)
	}
	if p.Filename != "sample.json" {
		t.Errorf("filename=%q, want \"sample.json\"", p.Filename)
	}
}

func TestBuildPayload_TreePresent(t *testing.T) {
	root := mustParse(t, `{"a": 1, "b": true}`)
	p, err := buildPayload(root, "test.json")
	if err != nil {
		t.Fatal(err)
	}
	if p.Tree == nil {
		t.Fatal("Tree should not be nil")
	}
}

func TestBuildPayload_NodeCount(t *testing.T) {
	root := mustParse(t, `{"a": 1, "b": 2}`)
	p, err := buildPayload(root, "test.json")
	if err != nil {
		t.Fatal(err)
	}
	// root + 2 children = 3
	if p.NodeCount != 3 {
		t.Errorf("nodeCount=%d, want 3", p.NodeCount)
	}
}

func TestBuildPayload_MaxDepth(t *testing.T) {
	root := mustParse(t, `{"outer": {"inner": 1}}`)
	p, err := buildPayload(root, "test.json")
	if err != nil {
		t.Fatal(err)
	}
	if p.MaxDepth != 2 {
		t.Errorf("maxDepth=%d, want 2", p.MaxDepth)
	}
}

func TestBuildPayload_YAMLPopulated(t *testing.T) {
	root := mustParse(t, `{"name": "alice"}`)
	p, err := buildPayload(root, "test.json")
	if err != nil {
		t.Fatal(err)
	}
	if p.YAML == "" {
		t.Error("YAML should be populated")
	}
}

func TestBuildPayload_TypegenKeys(t *testing.T) {
	root := mustParse(t, `{"count": 1}`)
	p, err := buildPayload(root, "test.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"go", "ts", "jsonschema"} {
		if _, ok := p.Typegen[key]; !ok {
			t.Errorf("typegen missing key %q", key)
		}
	}
}

func TestBuildPayload_MerkleRoot(t *testing.T) {
	root := mustParse(t, `{"x": 1}`)
	p, err := buildPayload(root, "test.json")
	if err != nil {
		t.Fatal(err)
	}
	if p.MerkleRoot == "" {
		t.Error("MerkleRoot should be populated")
	}
	if len(p.MerkleRoot) != 64 {
		t.Errorf("MerkleRoot length=%d, want 64 (SHA256 hex)", len(p.MerkleRoot))
	}
}

func TestBuildPayload_TreeKinds(t *testing.T) {
	root := mustParse(t, `{"s": "hello", "n": 42, "b": true, "arr": [1, 2]}`)
	p, err := buildPayload(root, "test.json")
	if err != nil {
		t.Fatal(err)
	}
	// Tree root should be an object
	if p.Tree.Kind != "object" {
		t.Errorf("root kind=%q, want \"object\"", p.Tree.Kind)
	}
}

func TestBuildPayload_Deterministic(t *testing.T) {
	root := mustParse(t, `{"a": 1, "b": [1, 2, 3]}`)
	p1, _ := buildPayload(root, "test.json")
	p2, _ := buildPayload(root, "test.json")
	if p1.MerkleRoot != p2.MerkleRoot {
		t.Errorf("non-deterministic MerkleRoot: %q vs %q", p1.MerkleRoot, p2.MerkleRoot)
	}
}

func TestServeIndex_InjectsConfig(t *testing.T) {
	d := shared.IndexData{DarkTheme: "nord", LightTheme: "solarized"}
	srv := httptest.NewServer(shared.ServeTemplate(assets, "assets/index.html")(d))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)

	want := `window.UNUM_CONFIG={darkTheme:"nord",lightTheme:"solarized"}`
	if !strings.Contains(string(body), want) {
		t.Errorf("response body does not contain %q\ngot: %s", want, body)
	}
}

func TestServeIndex_DefaultThemes(t *testing.T) {
	d := shared.IndexData{DarkTheme: "cyber", LightTheme: "clean"}
	srv := httptest.NewServer(shared.ServeTemplate(assets, "assets/index.html")(d))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)

	want := `window.UNUM_CONFIG={darkTheme:"cyber",lightTheme:"clean"}`
	if !strings.Contains(string(body), want) {
		t.Errorf("response body does not contain %q\ngot: %s", want, body)
	}
}
