package static

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/danielriddell21/unum/internal/diff/parse"
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
		if got.Added.Render("x") != tc.theme.Added.Render("x") {
			t.Errorf("ResolveTheme(%q) Added style mismatch", tc.name)
		}
	}
}

func TestResolveTheme_DefaultIsCyber(t *testing.T) {
	for _, name := range []string{"", "unknown", "cyber"} {
		got := ResolveTheme(name)
		if got.Added.Render("x") != Cyber.Added.Render("x") {
			t.Errorf("ResolveTheme(%q) should default to Cyber", name)
		}
	}
}

func TestRender_TextDiff(t *testing.T) {
	d, err := parse.Text([]byte("line one\nline two\n"), []byte("line one\nline three\n"), 3)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := Render(&buf, d, Options{Theme: Cyber, NoColor: true}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(buf.String(), "line three") {
		t.Errorf("diff output doesn't contain 'line three':\n%s", buf.String())
	}
}

func TestRender_Stat_OnlySummary(t *testing.T) {
	d, err := parse.Text([]byte("a\nb\n"), []byte("a\nc\n"), 3)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = Render(&buf, d, Options{Theme: Cyber, NoColor: true, Stat: true})
	out := buf.String()
	if strings.Contains(out, "@@") {
		t.Errorf("stat output should not contain hunk headers:\n%s", out)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) > 2 {
		t.Errorf("stat render: expected ≤2 lines, got %d:\n%s", len(lines), out)
	}
}

func TestRender_Identical(t *testing.T) {
	d, err := parse.Text([]byte("same\n"), []byte("same\n"), 3)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_ = Render(&buf, d, Options{Theme: Cyber, NoColor: true})
	out := buf.String()
	// Summary line should show +0 -0
	if !strings.Contains(out, "+0") || !strings.Contains(out, "-0") {
		t.Errorf("identical diff: expected +0 -0 in output:\n%s", out)
	}
}

func TestBoot_WritesOnFinish(t *testing.T) {
	var buf bytes.Buffer
	finish := Boot(&buf, "a.txt", "b.txt", Options{Theme: Cyber, NoColor: true})
	finish(2, 1, 0, 5*time.Millisecond)
	if buf.Len() == 0 {
		t.Error("Boot finish wrote nothing")
	}
}

func TestBoot_QuietSuppresses(t *testing.T) {
	var buf bytes.Buffer
	finish := Boot(&buf, "a.txt", "b.txt", Options{Theme: Cyber, Quiet: true})
	finish(0, 0, 0, 0)
	if buf.Len() != 0 {
		t.Error("Boot quiet should produce no output")
	}
}
