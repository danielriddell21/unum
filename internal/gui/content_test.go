package gui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func textOf(lines []Line) string {
	var b strings.Builder
	for _, ln := range lines {
		b.WriteString(ln.Text)
		b.WriteString("\n")
	}
	return b.String()
}

func TestJSONLines(t *testing.T) {
	cases := []struct {
		name     string
		content  string
		want     string
		wantKind LineKind
	}{
		{"object", `{"name":"unum","port":8080}`, `"name"`, KindPlain},
		{"array", `[1,2,3]`, "1", KindPlain},
		{"invalid", `{"name":`, "error:", KindRemoved},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lines := JSONLines(writeFile(t, "in.json", tc.content))
			if len(lines) == 0 {
				t.Fatal("no lines")
			}
			if got := textOf(lines); !strings.Contains(got, tc.want) {
				t.Errorf("output missing %q:\n%s", tc.want, got)
			}
			if lines[0].Kind != tc.wantKind {
				t.Errorf("first line kind = %d, want %d", lines[0].Kind, tc.wantKind)
			}
		})
	}
}

func TestJSONLinesMissingFile(t *testing.T) {
	lines := JSONLines(filepath.Join(t.TempDir(), "absent.json"))
	if len(lines) != 1 || lines[0].Kind != KindRemoved {
		t.Fatalf("expected single error line, got %+v", lines)
	}
}

func TestDiffLines(t *testing.T) {
	kinds := func(lines []Line) map[LineKind]bool {
		m := map[LineKind]bool{}
		for _, ln := range lines {
			m[ln.Kind] = true
		}
		return m
	}

	t.Run("text change", func(t *testing.T) {
		a := writeFile(t, "a.txt", "one\ntwo\nthree\n")
		b := writeFile(t, "b.txt", "one\n2\nthree\nfour\n")
		lines := DiffLines(a, b)
		k := kinds(lines)
		if !k[KindAdded] || !k[KindRemoved] {
			t.Errorf("expected added and removed lines:\n%s", textOf(lines))
		}
	})

	t.Run("json semantic", func(t *testing.T) {
		a := writeFile(t, "a.json", `{"port":80,"old":true}`)
		b := writeFile(t, "b.json", `{"port":8080,"new":true}`)
		lines := DiffLines(a, b)
		k := kinds(lines)
		if !k[KindAdded] || !k[KindRemoved] || !k[KindModified] {
			t.Errorf("expected added, removed, and modified lines:\n%s", textOf(lines))
		}
	})

	t.Run("identical", func(t *testing.T) {
		a := writeFile(t, "a.json", `{"port":80}`)
		lines := DiffLines(a, a)
		if got := textOf(lines); !strings.Contains(got, "no differences") {
			t.Errorf("expected no-differences marker:\n%s", got)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		a := writeFile(t, "a.txt", "one\n")
		lines := DiffLines(a, filepath.Join(t.TempDir(), "absent.txt"))
		if len(lines) != 1 || lines[0].Kind != KindRemoved {
			t.Fatalf("expected single error line, got %+v", lines)
		}
	})
}

func TestClassifyDiffLine(t *testing.T) {
	cases := []struct {
		line string
		want LineKind
	}{
		{"--- a/old.txt", KindDim},
		{"+++ b/new.txt", KindDim},
		{"@@ -1,3 +1,4 @@", KindDim},
		{"+ .port                                     8080", KindAdded},
		{"- .old                                      true", KindRemoved},
		{"! .port                                     80 → 8080", KindModified},
		{"   1    1  unchanged", KindPlain},
		{"        2 +added", KindAdded},
		{"   2      -removed", KindRemoved},
	}
	for _, tc := range cases {
		if got := classifyDiffLine(tc.line); got != tc.want {
			t.Errorf("classifyDiffLine(%q) = %d, want %d", tc.line, got, tc.want)
		}
	}
}

func TestHashLines(t *testing.T) {
	lines := HashLines("my-api-service")
	got := textOf(lines)
	for _, label := range []string{"port", "uuid", "color", "short", "phrase"} {
		if !strings.Contains(got, label) {
			t.Errorf("derivation table missing %q:\n%s", label, got)
		}
	}
	again := textOf(HashLines("my-api-service"))
	if got != again {
		t.Error("derivation is not deterministic")
	}
}

func TestHashLinesEmpty(t *testing.T) {
	lines := HashLines("")
	if len(lines) != 1 || lines[0].Kind != KindDim {
		t.Fatalf("expected single prompt line, got %+v", lines)
	}
}
