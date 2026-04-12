package web_test

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

var unumBin string

func TestMain(m *testing.M) {
	if err := os.Chdir("../.."); err != nil {
		fmt.Fprintf(os.Stderr, "chdir to repo root: %v\n", err)
		os.Exit(1)
	}

	bin, err := buildBinary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	unumBin = bin

	code := m.Run()
	_ = os.Remove(unumBin)
	os.Exit(code)
}

func buildBinary() (string, error) {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	f, err := os.CreateTemp("", "unum-web-test-*"+ext)
	if err != nil {
		return "", err
	}
	_ = f.Close()

	cmd := exec.Command("go", "build", "-o", f.Name(), "./cmd/unum")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%w\n%s", err, out)
	}
	return f.Name(), nil
}

// waitForServer polls url until a response is received or 5 seconds elapse.
func waitForServer(t *testing.T, url string) *http.Response {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var (
		resp *http.Response
		err  error
	)
	for time.Now().Before(deadline) {
		resp, err = http.Get(url) //nolint:noctx
		if err == nil {
			return resp
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("server did not start within 5s: %v", err)
	return nil
}

// keys returns the map keys as a slice for error messages.
func keys(m map[string]any) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

// newBrowser creates a headless rod browser. The browser is closed when the
// test ends. Fails the test if Chromium is not installed.
func newBrowser(t *testing.T) *rod.Browser {
	t.Helper()
	l := launcher.New().Headless(true).Leakless(false)
	if os.Getenv("CI") != "" {
		l = l.Set("no-sandbox")
	}
	u, err := l.Launch()
	if err != nil {
		t.Fatalf("failed to launch browser (is Chromium installed?): %v", err)
	}
	browser := rod.New().ControlURL(u).MustConnect()
	t.Cleanup(func() { _ = browser.Close() })
	return browser
}

// waitForElement waits up to 10 seconds for a CSS selector to appear.
func waitForElement(t *testing.T, page *rod.Page, selector string) *rod.Element {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		el, err := page.Element(selector)
		if err == nil && el != nil {
			return el
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("element %q not found within 10s", selector)
	return nil
}
