package web_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"testing"
)

var (
	webUUIDRe  = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	webColorRe = regexp.MustCompile(`^#[0-9a-f]{6}$`)
)

// TestHashWebContainerMode checks the hash --web server starts and returns 200
// on /api/derive?input=test.
func TestHashWebContainerMode(t *testing.T) {
	const port = "19882"
	cmd := exec.Command(unumBin, "hash", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	resp := waitForServer(t, fmt.Sprintf("http://localhost:%s/api/derive?input=test", port))
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /api/derive?input=test: status %d, want 200", resp.StatusCode)
	}
}

// TestHashWebAPI_DeriveReturnsAllFields checks that /api/derive?input=X
// returns a JSON object with all derived fields populated.
func TestHashWebAPI_DeriveReturnsAllFields(t *testing.T) {
	const port = "19864"
	cmd := exec.Command(unumBin, "hash", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	url := fmt.Sprintf("http://localhost:%s/api/derive?input=hello-world", port)
	resp := waitForServer(t, url)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/derive: status %d, want 200", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response not JSON: %v\nbody: %s", err, body)
	}

	for _, field := range []string{"Input", "Port", "UUID", "Color", "Short", "Emoji", "Phrase"} {
		if _, ok := result[field]; !ok {
			t.Errorf("response missing field %q; keys: %v", field, keys(result))
		}
	}
}

// TestHashWebAPI_DeriveFieldFormats validates UUID, Color, Port, Short, and
// Phrase formats.
func TestHashWebAPI_DeriveFieldFormats(t *testing.T) {
	const port = "19863"
	cmd := exec.Command(unumBin, "hash", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	url := fmt.Sprintf("http://localhost:%s/api/derive?input=my-service", port)
	resp := waitForServer(t, url)
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response not JSON: %v", err)
	}

	uuid, _ := result["UUID"].(string)
	if !webUUIDRe.MatchString(uuid) {
		t.Errorf("UUID = %q, want v5 UUID format", uuid)
	}

	color, _ := result["Color"].(string)
	if !webColorRe.MatchString(color) {
		t.Errorf("Color = %q, want #rrggbb", color)
	}

	port64, _ := result["Port"].(float64)
	p := int(port64)
	if p < 1024 || p > 65535 {
		t.Errorf("Port = %d, want 1024–65535", p)
	}

	short, _ := result["Short"].(string)
	if len(short) != 8 {
		t.Errorf("Short = %q, want 8 hex chars", short)
	}

	phrase, _ := result["Phrase"].(string)
	parts := splitHyphens(phrase)
	if len(parts) != 3 {
		t.Errorf("Phrase = %q, want 3 hyphen-separated words", phrase)
	}
}

// TestHashWebAPI_DeriveNoInputReturns400 checks that omitting ?input returns
// 400.
func TestHashWebAPI_DeriveNoInputReturns400(t *testing.T) {
	const port = "19862"
	cmd := exec.Command(unumBin, "hash", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	ready := waitForServer(t, fmt.Sprintf("http://localhost:%s/api/derive?input=ping", port))
	_ = ready.Body.Close()

	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/api/derive", port)) //nolint:noctx // integration test hitting local server; context not needed in tests
	if err != nil {
		t.Fatalf("GET /api/derive: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("missing input: status %d, want 400", resp.StatusCode)
	}
}

// TestHashWebAPI_DeriveIsDeterministic checks that identical inputs produce
// identical outputs on repeated calls.
func TestHashWebAPI_DeriveIsDeterministic(t *testing.T) {
	const port = "19861"
	cmd := exec.Command(unumBin, "hash", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	url := fmt.Sprintf("http://localhost:%s/api/derive?input=determinism-check", port)
	_ = waitForServer(t, url)

	fetch := func() []byte {
		r, err := http.Get(url) //nolint:noctx // integration test hitting local server; context not needed in tests
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		defer func() { _ = r.Body.Close() }()
		b, _ := io.ReadAll(r.Body)
		return b
	}

	a, b := fetch(), fetch()
	if string(a) != string(b) {
		t.Errorf("responses differ:\nfirst:  %s\nsecond: %s", a, b)
	}
}

// TestWebFrontend_HashUIDerivesOnSubmit navigates to the hash web UI, types
// into the input, clicks derive, and checks that derived values appear.
// Requires Chromium to be installed.
func TestWebFrontend_HashUIDerivesOnSubmit(t *testing.T) {
	const port = "19860"
	cmd := exec.Command(unumBin, "hash", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	_ = waitForServer(t, fmt.Sprintf("http://localhost:%s/api/derive?input=ping", port))

	browser := newBrowser(t)
	page := browser.MustPage(fmt.Sprintf("http://localhost:%s/", port))
	page.MustWaitLoad()

	waitForElement(t, page, "#hashInput").MustInput("test-service")
	page.MustElement("#deriveBtn").MustClick()

	waitForElement(t, page, "#resultBody")

	rows := page.MustElements("#resultBody tr")
	if len(rows) == 0 {
		t.Error("expected result rows in #resultBody after deriving")
	}
}

func splitHyphens(s string) []string {
	var parts []string
	start := 0
	for i, c := range s {
		if c == '-' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}
