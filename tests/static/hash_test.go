package static_test

import (
	"regexp"
	"strings"
	"sync"
	"testing"
)

var (
	hashUUIDRe  = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	hashColorRe = regexp.MustCompile(`^#[0-9a-f]{6}$`)
)

func TestHash_TableExitZero(t *testing.T) {
	stdout, _, code := run("hash", "my-api-service")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if !strings.Contains(stdout, "my-api-service") {
		t.Error("output should contain the input string")
	}
	if !strings.Contains(stdout, "port") {
		t.Error("output should contain 'port'")
	}
	if !strings.Contains(stdout, "uuid") {
		t.Error("output should contain 'uuid'")
	}
}

func TestHash_NoColor(t *testing.T) {
	stdout, _, code := run("hash", "--no-color", "my-api-service")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if !strings.Contains(stdout, "my-api-service") {
		t.Error("output should contain the input string")
	}
	if strings.Contains(stdout, "\x1b[") {
		t.Error("output should not contain ANSI escape codes with --no-color")
	}
}

func TestHash_Deterministic(t *testing.T) {
	stdout1, _, _ := run("hash", "--no-color", "stable-service")
	stdout2, _, _ := run("hash", "--no-color", "stable-service")
	if stdout1 != stdout2 {
		t.Errorf("hash is not deterministic:\n  run1=%q\n  run2=%q", stdout1, stdout2)
	}
}

func TestHash_PortFlag(t *testing.T) {
	stdout, _, code := run("hash", "--port", "my-api-service")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	trimmed := strings.TrimSpace(stdout)
	if trimmed == "" {
		t.Fatal("--port produced empty output")
	}
	for _, ch := range trimmed {
		if ch < '0' || ch > '9' {
			t.Errorf("--port output %q contains non-digit %q", trimmed, ch)
			break
		}
	}
}

func TestHash_UUIDFlag(t *testing.T) {
	stdout, _, code := run("hash", "--uuid", "my-api-service")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	trimmed := strings.TrimSpace(stdout)
	if !hashUUIDRe.MatchString(trimmed) {
		t.Errorf("--uuid output %q does not match UUID v5 pattern", trimmed)
	}
}

func TestHash_ColorFlag(t *testing.T) {
	stdout, _, code := run("hash", "--color", "my-api-service")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	trimmed := strings.TrimSpace(stdout)
	if !hashColorRe.MatchString(trimmed) {
		t.Errorf("--color output %q does not match #rrggbb pattern", trimmed)
	}
}

func TestHash_ShortFlag(t *testing.T) {
	stdout, _, code := run("hash", "--short", "my-api-service")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	trimmed := strings.TrimSpace(stdout)
	if len(trimmed) != 8 {
		t.Errorf("--short output %q: want 8 chars, got %d", trimmed, len(trimmed))
	}
}

func TestHash_EmojiFlag(t *testing.T) {
	stdout, _, code := run("hash", "--emoji", "my-api-service")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Error("--emoji produced empty output")
	}
}

func TestHash_PhraseFlag(t *testing.T) {
	stdout, _, code := run("hash", "--phrase", "my-api-service")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	trimmed := strings.TrimSpace(stdout)
	parts := strings.Split(trimmed, "-")
	if len(parts) != 3 {
		t.Errorf("--phrase output %q: want 3 words, got %d", trimmed, len(parts))
	}
}

func TestHash_NoArgs_Error(t *testing.T) {
	_, _, code := run("hash")
	if code == 0 {
		t.Error("expected non-zero exit when no text argument provided")
	}
}

func TestHashConcurrent(t *testing.T) {
	inputs := []string{"service-alpha", "service-beta"}
	results := make([]string, 2)
	errs := make([]int, 2)

	var wg sync.WaitGroup
	for i, input := range inputs {
		i, input := i, input
		wg.Add(1)
		go func() {
			defer wg.Done()
			stdout, _, code := run("hash", input, "--no-color")
			results[i] = stdout
			errs[i] = code
		}()
	}
	wg.Wait()

	for i, code := range errs {
		if code != 0 {
			t.Errorf("goroutine %d: exit %d", i, code)
		}
	}
	for i, out := range results {
		if !strings.Contains(out, inputs[i]) {
			t.Errorf("goroutine %d: output missing input %q:\n%s", i, inputs[i], out)
		}
	}
}
