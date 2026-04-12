package static

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

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
