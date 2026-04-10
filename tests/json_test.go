package tests_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var sampleFile = filepath.Join("testdata", "sample.json")

// ─── Static mode ─────────────────────────────────────────────────────────────

func TestStaticMode(t *testing.T) {
	stdout, _, code := run("json", sampleFile, "--no-color")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "corpus") {
		t.Errorf("expected 'corpus' key in output:\n%s", stdout)
	}
}

func TestValidateOnly(t *testing.T) {
	_, _, code := run("json", sampleFile, "--validate-only")
	if code != 0 {
		t.Fatalf("--validate-only: exit %d, want 0", code)
	}
}

func TestValidateInvalidJSON(t *testing.T) {
	f, err := os.CreateTemp("", "invalid-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(f.Name()) }()
	_, _ = f.WriteString(`{broken`)
	_ = f.Close()

	_, _, code := run("json", f.Name(), "--validate-only")
	if code != 1 {
		t.Fatalf("invalid JSON: exit %d, want 1", code)
	}
}

func TestCompactMode(t *testing.T) {
	stdout, _, code := run("json", sampleFile, "--compact", "--no-color")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 1 {
		t.Errorf("compact: expected 1 line, got %d", len(lines))
	}
}

// ─── Stats lens ───────────────────────────────────────────────────────────────

func TestStatsMode(t *testing.T) {
	stdout, _, code := run("json", sampleFile, "--stats", "--no-color")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "//") {
		t.Errorf("expected stats annotation '//' in output:\n%s", stdout)
	}
}

// ─── Merkle lens ──────────────────────────────────────────────────────────────

func TestHashOnly(t *testing.T) {
	stdout, _, code := run("json", sampleFile, "--merkle", "--hash-only")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	hash := strings.TrimSpace(stdout)
	if len(hash) != 64 {
		t.Errorf("expected 64-char SHA256 hash, got %d chars: %q", len(hash), hash)
	}
}

func TestHashDeterministic(t *testing.T) {
	stdout1, _, _ := run("json", sampleFile, "--merkle", "--hash-only")
	stdout2, _, _ := run("json", sampleFile, "--merkle", "--hash-only")
	if strings.TrimSpace(stdout1) != strings.TrimSpace(stdout2) {
		t.Errorf("non-deterministic hash:\n  run1: %s\n  run2: %s", stdout1, stdout2)
	}
}

// ─── Transform lens ───────────────────────────────────────────────────────────

func TestTransformYAML(t *testing.T) {
	stdout, _, code := run("json", sampleFile, "--transform")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "corpus:") {
		t.Errorf("expected 'corpus:' in YAML output:\n%s", stdout)
	}
	if strings.HasPrefix(strings.TrimSpace(stdout), "{") {
		t.Errorf("output looks like JSON, not YAML:\n%s", stdout)
	}
}

// ─── TypeGen lens ─────────────────────────────────────────────────────────────

func TestTypegenGo(t *testing.T) {
	stdout, _, code := run("json", sampleFile, "--typegen", "go")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "package main") {
		t.Errorf("expected 'package main':\n%s", stdout)
	}
	if !strings.Contains(stdout, "struct") {
		t.Errorf("expected 'struct':\n%s", stdout)
	}
}

func TestTypegenTS(t *testing.T) {
	stdout, _, code := run("json", sampleFile, "--typegen", "ts")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "interface") {
		t.Errorf("expected 'interface':\n%s", stdout)
	}
}

func TestTypegenJSONSchema(t *testing.T) {
	stdout, _, code := run("json", sampleFile, "--typegen", "jsonschema")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, `"$schema"`) {
		t.Errorf("expected '$schema':\n%s", stdout)
	}
}

// ─── Query lens ───────────────────────────────────────────────────────────────

func TestQuery(t *testing.T) {
	stdout, _, code := run("json", sampleFile, "--query", ".corpus")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "Lexica Technica") {
		t.Errorf("expected 'Lexica Technica':\n%s", stdout)
	}
}

func TestQueryArrayLength(t *testing.T) {
	stdout, _, code := run("json", sampleFile, "--query", ".pageScores | length")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.TrimSpace(stdout) != "10" {
		t.Errorf("expected '10', got %q", stdout)
	}
}
