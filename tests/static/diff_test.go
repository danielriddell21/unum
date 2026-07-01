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

const (
	flagNoColor = "--no-color"
	exitFmt     = "exit %d"
)

func TestDiffTextStatic(t *testing.T) {
	stdout, _, code := run("diff", diffA, diffB, flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
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
	stdout, _, code := run("diff", diffA, diffB, "--stat", flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
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
	stdout, _, code := run("diff", diffAJSON, diffBJSON, flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if !strings.ContainsAny(stdout, "~+-") {
		t.Errorf("expected semantic change markers in output:\n%s", stdout)
	}
}

func TestDiffYAMLSemantic(t *testing.T) {
	stdout, _, code := run("diff", diffAYAML, diffBYAML, flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if !strings.ContainsAny(stdout, "~+-") {
		t.Errorf("expected semantic change markers in output:\n%s", stdout)
	}
}

func TestDiffTerraform(t *testing.T) {
	stdout, _, code := run("diff", diffATF, diffATF, "--format", "terraform", flagNoColor)
	if code != 0 {
		t.Fatalf("terraform diff: exit %d, want 0", code)
	}
	// The CLI prints just the diff body — no preamble, no footer.
	wants := []string{
		"# aws_instance.web will be updated in-place",
		`resource "aws_instance" "web" {`,
		`instance_type = "t2.micro" -> "t3.small"`,
		"# aws_s3_bucket.logs will be created",
		"# aws_security_group.legacy will be destroyed",
		"unchanged attribute hidden)",
	}
	for _, w := range wants {
		if !strings.Contains(stdout, w) {
			t.Errorf("terraform diff output missing %q\n%s", w, stdout)
		}
	}
	if strings.Contains(stdout, "Terraform will perform") || strings.Contains(stdout, "Plan:") {
		t.Errorf("CLI diff should be body-only (no preamble/footer)\n%s", stdout)
	}
}

func TestDiffTerraformStat(t *testing.T) {
	stdout, _, code := run("diff", diffATF, diffATF, "--format", "terraform", "--stat", flagNoColor)
	if code != 0 {
		t.Fatalf("terraform diff --stat: exit %d, want 0", code)
	}
	if !strings.Contains(stdout, "Plan: 1 to add, 1 to change, 1 to destroy, 0 to replace.") {
		t.Errorf("--stat should print the Plan summary line\n%s", stdout)
	}
	if strings.Contains(stdout, "resource \"aws_instance\"") {
		t.Errorf("--stat should not print the full diff body\n%s", stdout)
	}
}

func TestDiffFormatOverride(t *testing.T) {
	stdout1, _, code1 := run("diff", diffAJSON, diffBJSON, flagNoColor)
	stdout2, _, code2 := run("diff", diffAJSON, diffBJSON, "--format", "json", flagNoColor)
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
	stdout, _, code := run("diff", diffA, diffB, flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if strings.Contains(stdout, "\033[") {
		t.Errorf("--no-color output contains ANSI escapes:\n%s", stdout)
	}
}

func TestDiffContextFlag(t *testing.T) {
	stdout, _, code := run("diff", diffA, diffB, "--context", "0", flagNoColor)
	if code != 0 {
		t.Fatalf(exitFmt, code)
	}
	if !strings.Contains(stdout, "+") || !strings.Contains(stdout, "-") {
		t.Errorf("--context 0: expected + and - lines:\n%s", stdout)
	}
}
