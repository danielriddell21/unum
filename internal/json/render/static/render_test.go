package static

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/danielriddell21/unum/internal/json/lens/merkle"
	"github.com/danielriddell21/unum/internal/json/lens/stats"
	"github.com/danielriddell21/unum/internal/json/node"
	"github.com/danielriddell21/unum/internal/json/parse"
)

func TestResolveTheme_KnownNames(t *testing.T) {
	cases := []struct {
		name  string
		theme Theme
	}{
		{"matrix", Matrix},
		{"dracula", Dracula},
		{"nord", Nord},
	}
	for _, tc := range cases {
		got := ResolveTheme(tc.name)
		want := tc.theme.ObjectKey.Render("x")
		if got.ObjectKey.Render("x") != want {
			t.Errorf("ResolveTheme(%q) ObjectKey style mismatch", tc.name)
		}
	}
}

func TestResolveTheme_DefaultIsCyber(t *testing.T) {
	for _, name := range []string{"", "unknown", "cyber"} {
		got := ResolveTheme(name)
		if got.ObjectKey.Render("x") != Cyber.ObjectKey.Render("x") {
			t.Errorf("ResolveTheme(%q) should default to Cyber", name)
		}
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.ShowLineNums != true {
		t.Error("DefaultOptions: ShowLineNums should be true")
	}
}

func TestRender_ProducesOutput(t *testing.T) {
	root, err := parse.Parse([]byte(`{"key": "value"}`))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	opts := Options{Theme: Cyber, NoColor: true}
	if err := Render(&buf, root, opts); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("Render produced no output")
	}
}

func TestRender_ContainsKey(t *testing.T) {
	root, err := parse.Parse([]byte(`{"mykey": "myval"}`))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = Render(&buf, root, Options{Theme: Cyber, NoColor: true})
	if !strings.Contains(buf.String(), "mykey") {
		t.Errorf("output does not contain key 'mykey':\n%s", buf.String())
	}
}

func TestRender_Compact(t *testing.T) {
	root, err := parse.Parse([]byte(`{"a": 1, "b": 2}`))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = Render(&buf, root, Options{Theme: Cyber, Compact: true})
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("compact render: expected 1 line, got %d", len(lines))
	}
}

func TestRender_Array(t *testing.T) {
	root, err := parse.Parse([]byte(`[1, "two", true, null]`))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = Render(&buf, root, Options{Theme: Cyber, NoColor: true})
	out := buf.String()
	for _, want := range []string{"1", "two", "true", "null"} {
		if !strings.Contains(out, want) {
			t.Errorf("array render missing %q in:\n%s", want, out)
		}
	}
}

func TestRender_EmptyObject(t *testing.T) {
	root, err := parse.Parse([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = Render(&buf, root, Options{Theme: Cyber, NoColor: true})
	if !strings.Contains(buf.String(), "{}") {
		t.Errorf("empty object render: want '{}' in %q", buf.String())
	}
}

func TestRender_EmptyArray(t *testing.T) {
	root, err := parse.Parse([]byte(`[]`))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = Render(&buf, root, Options{Theme: Cyber, NoColor: true})
	if !strings.Contains(buf.String(), "[]") {
		t.Errorf("empty array render: want '[]' in %q", buf.String())
	}
}

func TestRender_NestedObjectInObject(t *testing.T) {
	root, err := parse.Parse([]byte(`{"a": {"b": 1}}`))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = Render(&buf, root, Options{Theme: Cyber, NoColor: true})
	out := buf.String()
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Errorf("nested object render missing keys in:\n%s", out)
	}
}

func TestRender_NestedArrayInObject(t *testing.T) {
	root, err := parse.Parse([]byte(`{"items": [1, 2, 3]}`))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = Render(&buf, root, Options{Theme: Cyber, NoColor: true})
	out := buf.String()
	if !strings.Contains(out, "items") {
		t.Errorf("nested array render missing key in:\n%s", out)
	}
}

func TestRender_NestedObjectInArray(t *testing.T) {
	root, err := parse.Parse([]byte(`[{"x": 1}, {"x": 2}]`))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = Render(&buf, root, Options{Theme: Cyber, NoColor: true})
	out := buf.String()
	if !strings.Contains(out, "x") {
		t.Errorf("nested object in array render missing key in:\n%s", out)
	}
}

func TestRender_WithLineNums(t *testing.T) {
	root, err := parse.Parse([]byte(`{"a": 1}`))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = Render(&buf, root, Options{Theme: Cyber, ShowLineNums: true, NoColor: true})
	out := buf.String()
	if !strings.Contains(out, "1") {
		t.Errorf("line number render: missing '1' in:\n%s", out)
	}
}

func TestRender_AllLeafKinds(t *testing.T) {
	src := `{"s": "str", "n": 3.14, "b": false, "z": null}`
	root, err := parse.Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = Render(&buf, root, Options{Theme: Cyber, NoColor: true})
	out := buf.String()
	for _, want := range []string{"str", "3.14", "false", "null"} {
		if !strings.Contains(out, want) {
			t.Errorf("leaf kind render missing %q in:\n%s", want, out)
		}
	}
}

func TestBoot_WritesOnFinish(t *testing.T) {
	var buf bytes.Buffer
	finish := Boot(&buf, "test.json", Options{Theme: Cyber, NoColor: true})
	finish(10, 5*time.Millisecond)
	if buf.Len() == 0 {
		t.Error("Boot finish func produced no output")
	}
}

func TestBoot_QuietSuppresses(t *testing.T) {
	var buf bytes.Buffer
	finish := Boot(&buf, "test.json", Options{Theme: Cyber, Quiet: true})
	finish(10, 5*time.Millisecond)
	if buf.Len() != 0 {
		t.Error("Boot with Quiet should produce no output")
	}
}

func TestBoot_ColoredMode(t *testing.T) {
	var buf bytes.Buffer
	finish := Boot(&buf, "test.json", Options{Theme: Cyber})
	finish(5, 10*time.Millisecond)
	if buf.Len() == 0 {
		t.Error("Boot colored mode should write output")
	}
}

func TestRenderError_NoColor(t *testing.T) {
	var buf bytes.Buffer
	RenderError(&buf, errors.New("something went wrong"), Options{Theme: Cyber, NoColor: true})
	out := buf.String()
	if !strings.Contains(out, "error:") {
		t.Errorf("RenderError NoColor: %q, want 'error:' prefix", out)
	}
	if !strings.Contains(out, "something went wrong") {
		t.Errorf("RenderError: %q, want error message", out)
	}
}

func TestRenderError_WithColor(t *testing.T) {
	var buf bytes.Buffer
	RenderError(&buf, errors.New("boom"), Options{Theme: Cyber})
	if !strings.Contains(buf.String(), "boom") {
		t.Errorf("RenderError with color missing error message")
	}
}

func TestRender_ShowMerkle(t *testing.T) {
	// Merkle hash is shown via renderInline on nested children.
	root, err := parse.Parse([]byte(`{"nested": {"k": 1}}`))
	if err != nil {
		t.Fatal(err)
	}
	// Annotate the nested child object with a fake merkle hash.
	nestedNode := root.Children[0]
	nestedNode.Annotate(node.Annotation{
		Key:   node.AnnotationKey{Lens: merkle.Lens, Name: "hash"},
		Value: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
	})
	var buf bytes.Buffer
	_ = Render(&buf, root, Options{Theme: Cyber, ShowMerkle: true, NoColor: true})
	if !strings.Contains(buf.String(), "#") {
		t.Errorf("ShowMerkle render should include '#' hash tag, got:\n%s", buf.String())
	}
}

func TestRender_ShowStats(t *testing.T) {
	root, err := parse.Parse([]byte(`{"nums": [1, 2, 3]}`))
	if err != nil {
		t.Fatal(err)
	}
	arrNode := root.Children[0]
	arrNode.Annotate(node.Annotation{Key: node.AnnotationKey{Lens: stats.Lens, Name: "numeric_count"}, Value: 3})
	arrNode.Annotate(node.Annotation{Key: node.AnnotationKey{Lens: stats.Lens, Name: "count"}, Value: 3})
	arrNode.Annotate(node.Annotation{Key: node.AnnotationKey{Lens: stats.Lens, Name: "min"}, Value: 1.0})
	arrNode.Annotate(node.Annotation{Key: node.AnnotationKey{Lens: stats.Lens, Name: "max"}, Value: 3.0})
	arrNode.Annotate(node.Annotation{Key: node.AnnotationKey{Lens: stats.Lens, Name: "mean"}, Value: 2.0})

	var buf bytes.Buffer
	_ = Render(&buf, root, Options{Theme: Cyber, ShowStats: true, NoColor: true})
	out := buf.String()
	if !strings.Contains(out, "n=3") {
		t.Errorf("ShowStats render should include 'n=3', got:\n%s", out)
	}
}

func TestEscapeForDisplay(t *testing.T) {
	tests := []struct{ in, want string }{
		{"hello", "hello"},
		{"line\nbreak", `line\nbreak`},
		{"carriage\rreturn", `carriage\rreturn`},
		{"tab\there", `tab\there`},
	}
	for _, tc := range tests {
		if got := escapeForDisplay(tc.in); got != tc.want {
			t.Errorf("escapeForDisplay(%q)=%q, want %q", tc.in, got, tc.want)
		}
	}
}
