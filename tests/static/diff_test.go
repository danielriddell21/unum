package static_test

import (
	"path/filepath"
	"strings"
	"testing"
)

var (
	diffA     = filepath.Join("testdata", "diff-a.txt")
	diffB     = filepath.Join("testdata", "diff-b.txt")
	diffAJSON = filepath.Join("testdata", "diff-a.json")
	diffBJSON = filepath.Join("testdata", "diff-b.json")
	diffAYAML = filepath.Join("testdata", "diff-a.yaml")
	diffBYAML = filepath.Join("testdata", "diff-b.yaml")
	diffATF   = filepath.Join("testdata", "diff-a.tfplan.json")
)

func TestDiffTextStatic(t *testing.T) {
	stdout, _, code := run("diff", diffA, diffB, "--no-color")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "+") || !strings.Contains(stdout, "-") {
		t.Errorf("expected + and - lines in output:\n%s", stdout)
	}
}

func TestDiffIdentical(t *testing.T) {
	_, _, code := run("diff", diffA, diffA)
	if code != 0 {
		t.Fatalf("identical diff: exit %d, want 0", code)
	}
}

func TestDiffStat(t *testing.T) {
	stdout, _, code := run("diff", diffA, diffB, "--stat", "--no-color")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "+") || !strings.Contains(stdout, "-") {
		t.Errorf("expected count summary in stat output:\n%s", stdout)
	}
	if strings.Contains(stdout, "@@") {
		t.Errorf("--stat output should not contain hunk headers:\n%s", stdout)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) > 2 {
		t.Errorf("--stat: expected ≤2 lines, got %d:\n%s", len(lines), stdout)
	}
}

func TestDiffJSONSemantic(t *testing.T) {
	stdout, _, code := run("diff", diffAJSON, diffBJSON, "--no-color")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.ContainsAny(stdout, "~+-") {
		t.Errorf("expected semantic change markers in output:\n%s", stdout)
	}
}

func TestDiffYAMLSemantic(t *testing.T) {
	stdout, _, code := run("diff", diffAYAML, diffBYAML, "--no-color")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.ContainsAny(stdout, "~+-") {
		t.Errorf("expected semantic change markers in output:\n%s", stdout)
	}
}

func TestDiffTerraform(t *testing.T) {
	_, _, code := run("diff", diffATF, diffATF, "--format", "terraform", "--no-color")
	if code != 0 {
		t.Fatalf("terraform diff: exit %d, want 0", code)
	}
}

func TestDiffFormatOverride(t *testing.T) {
	stdout1, _, code1 := run("diff", diffAJSON, diffBJSON, "--no-color")
	stdout2, _, code2 := run("diff", diffAJSON, diffBJSON, "--format", "json", "--no-color")
	if code1 != 0 || code2 != 0 {
		t.Fatalf("exit codes: %d %d", code1, code2)
	}
	if stdout1 != stdout2 {
		t.Errorf("explicit --format json differs from auto-detect")
	}
}

func TestDiffNoArgsNoWebFlag(t *testing.T) {
	_, stderr, code := run("diff")
	if code == 0 {
		t.Fatal("expected non-zero exit when no args and no --web")
	}
	if stderr == "" {
		t.Error("expected error message in stderr")
	}
}

func TestDiffMissingFile(t *testing.T) {
	_, stderr, code := run("diff", "nonexistent.txt", diffB)
	if code == 0 {
		t.Fatal("expected non-zero exit for missing file")
	}
	if stderr == "" {
		t.Error("expected error message in stderr")
	}
}

func TestDiffNoColorFlag(t *testing.T) {
	stdout, _, code := run("diff", diffA, diffB, "--no-color")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.Contains(stdout, "\033[") {
		t.Errorf("--no-color output contains ANSI escapes:\n%s", stdout)
	}
}

func TestDiffContextFlag(t *testing.T) {
	stdout, _, code := run("diff", diffA, diffB, "--context", "0", "--no-color")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "+") || !strings.Contains(stdout, "-") {
		t.Errorf("--context 0: expected + and - lines:\n%s", stdout)
	}
}
