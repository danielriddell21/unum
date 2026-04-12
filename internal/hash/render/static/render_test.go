package static

import (
	"bytes"
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/hash/types"
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
		if got.Label.Render("x") != tc.theme.Label.Render("x") {
			t.Errorf("ResolveTheme(%q) Label style mismatch", tc.name)
		}
	}
}

func TestResolveTheme_DefaultIsCyber(t *testing.T) {
	for _, name := range []string{"", "unknown", "cyber"} {
		got := ResolveTheme(name)
		if got.Label.Render("x") != Cyber.Label.Render("x") {
			t.Errorf("ResolveTheme(%q) should default to Cyber", name)
		}
	}
}

var sampleResult = types.Result{
	Input:  "my-service",
	Port:   12345,
	UUID:   "abc-uuid-123",
	Color:  "#001122",
	Short:  "abcd1234",
	Emoji:  "🚀",
	Phrase: "alpha-beta-gamma",
}

func TestRenderTable_ContainsAllLabels(t *testing.T) {
	var buf bytes.Buffer
	RenderTable(&buf, sampleResult, Options{Theme: Cyber, NoColor: true})
	out := buf.String()
	for _, label := range []string{"input", "port", "uuid", "color", "short", "emoji", "phrase"} {
		if !strings.Contains(out, label) {
			t.Errorf("RenderTable missing label %q:\n%s", label, out)
		}
	}
}

func TestRenderTable_ContainsValue(t *testing.T) {
	var buf bytes.Buffer
	RenderTable(&buf, sampleResult, Options{Theme: Cyber, NoColor: true})
	if !strings.Contains(buf.String(), "my-service") {
		t.Errorf("RenderTable missing input value:\n%s", buf.String())
	}
}

func TestRenderSingle_WritesValue(t *testing.T) {
	var buf bytes.Buffer
	RenderSingle(&buf, "abc123")
	if strings.TrimSpace(buf.String()) != "abc123" {
		t.Errorf("RenderSingle: got %q, want 'abc123'", buf.String())
	}
}

func TestBoot_NoColor(t *testing.T) {
	var buf bytes.Buffer
	Boot(&buf, Options{Theme: Cyber, NoColor: true})
	if !strings.Contains(buf.String(), "[ UNUM ]") {
		t.Errorf("Boot NoColor: %q, want '[ UNUM ]'", buf.String())
	}
}

func TestBoot_Quiet(t *testing.T) {
	var buf bytes.Buffer
	Boot(&buf, Options{Theme: Cyber, Quiet: true})
	if buf.Len() != 0 {
		t.Error("Boot Quiet should produce no output")
	}
}
