package tests_test

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

var hashUUIDRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
var hashColorRe = regexp.MustCompile(`^#[0-9a-f]{6}$`)

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
	// No ANSI escape codes
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
	// Should be a single number, nothing else
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

// ─── Web container mode ───────────────────────────────────────────────────────

func TestHashWebContainerMode(t *testing.T) {
	port := "19882"
	cmd := exec.Command(unumBin, "hash", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)

	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	url := fmt.Sprintf("http://localhost:%s/api/derive?input=test", port)
	deadline := time.Now().Add(5 * time.Second)
	var resp *http.Response
	var err error
	for time.Now().Before(deadline) {
		resp, err = http.Get(url) //nolint:noctx
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("server did not start within 5s: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /api/derive?input=test: status %d, want 200", resp.StatusCode)
	}
}
