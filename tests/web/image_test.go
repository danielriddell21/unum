package web_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const imageStartServerErr = "start server: %v"

func startImageServer(t *testing.T, port string) {
	t.Helper()

	cmd := exec.Command(unumBin, "image", filepath.Join("testdata", "sample.jpg"), "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf(imageStartServerErr, err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill() })

	resp := waitForServer(t, fmt.Sprintf("http://localhost:%s/api/source", port))
	_ = resp.Body.Close()
}

func uploadSample(t *testing.T, port string) map[string]any {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", "sample.jpg"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	resp, err := http.Post( //nolint:noctx // integration test hitting a local server
		fmt.Sprintf("http://localhost:%s/api/upload?name=sample.jpg", port),
		"application/octet-stream", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST /api/upload: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload: status %d, want 200", resp.StatusCode)
	}

	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	return out
}

func getJSON(t *testing.T, url string) map[string]any {
	t.Helper()

	resp, err := http.Get(url) //nolint:noctx // integration test hitting a local server
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET %s: status %d, body %s", url, resp.StatusCode, body)
	}

	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode %s: %v", url, err)
	}
	return out
}

func TestImageWebContainerMode(t *testing.T) {
	const port = "19891"
	startImageServer(t, port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/", port)) //nolint:noctx // integration test
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /: status %d, want 200", resp.StatusCode)
	}
}

func TestImageWebAPI_ServesTheLaunchImage(t *testing.T) {
	const port = "19892"
	startImageServer(t, port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/api/source", port)) //nolint:noctx // integration test
	if err != nil {
		t.Fatalf("GET /api/source: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get("Content-Type"); got != "image/jpeg" {
		t.Errorf("Content-Type = %q, want image/jpeg", got)
	}
	body, _ := io.ReadAll(resp.Body)
	if !bytes.HasPrefix(body, []byte{0xFF, 0xD8, 0xFF}) {
		t.Error("/api/source did not return a jpeg")
	}
}

func TestImageWebAPI_UploadReturnsMetadata(t *testing.T) {
	const port = "19893"
	startImageServer(t, port)

	got := uploadSample(t, port)

	for _, field := range []string{"id", "name", "format", "width", "height", "bytes"} {
		if _, ok := got[field]; !ok {
			t.Errorf("upload response missing %q; keys: %v", field, keys(got))
		}
	}
	if got["format"] != "jpeg" {
		t.Errorf("format = %v, want jpeg", got["format"])
	}
	if got["width"].(float64) != 1200 || got["height"].(float64) != 800 {
		t.Errorf("dimensions = %vx%v, want 1200x800", got["width"], got["height"])
	}
}

func TestImageWebAPI_OptimizeShrinksTheImage(t *testing.T) {
	const port = "19894"
	startImageServer(t, port)

	uploaded := uploadSample(t, port)
	id := uploaded["id"].(string)
	sourceBytes := uploaded["bytes"].(float64)

	got := getJSON(t, fmt.Sprintf("http://localhost:%s/api/optimize?id=%s&quality=50", port, id))

	outBytes, ok := got["bytes"].(float64)
	if !ok {
		t.Fatalf("response missing bytes; keys: %v", keys(got))
	}
	if outBytes >= sourceBytes {
		t.Errorf("optimized to %v bytes, no smaller than the %v byte source", outBytes, sourceBytes)
	}
	if got["quality"].(float64) != 50 {
		t.Errorf("quality = %v, want 50", got["quality"])
	}
	if url, _ := got["dataUrl"].(string); !strings.HasPrefix(url, "data:image/jpeg;base64,") {
		t.Errorf("dataUrl = %.40q, want a jpeg data url", url)
	}
}

func TestImageWebAPI_OptimizeHonorsScale(t *testing.T) {
	const port = "19895"
	startImageServer(t, port)

	id := uploadSample(t, port)["id"].(string)

	got := getJSON(t, fmt.Sprintf("http://localhost:%s/api/optimize?id=%s&scale=50%%25", port, id))

	if got["width"].(float64) != 600 || got["height"].(float64) != 400 {
		t.Errorf("dimensions = %vx%v, want 600x400 at half scale", got["width"], got["height"])
	}
}

func TestImageWebAPI_AnalyzeReturnsTheLadder(t *testing.T) {
	const port = "19896"
	startImageServer(t, port)

	id := uploadSample(t, port)["id"].(string)

	got := getJSON(t, fmt.Sprintf("http://localhost:%s/api/analyze?id=%s", port, id))

	steps, ok := got["steps"].([]any)
	if !ok {
		t.Fatalf("response missing steps; keys: %v", keys(got))
	}
	if len(steps) == 0 {
		t.Fatal("expected ladder steps")
	}
	for _, raw := range steps {
		step := raw.(map[string]any)
		if step["bytes"].(float64) <= 0 {
			t.Errorf("step %v reported no bytes", step["quality"])
		}
	}
}

func TestImageWebAPI_ErrorCodes(t *testing.T) {
	const port = "19897"
	startImageServer(t, port)

	id := uploadSample(t, port)["id"].(string)

	tests := []struct {
		name string
		path string
		want int
	}{
		{"missing id", "/api/optimize", http.StatusBadRequest},
		{"unknown id", "/api/optimize?id=0123456789abcdef", http.StatusNotFound},
		{"bad format", "/api/optimize?id=" + id + "&format=avif", http.StatusBadRequest},
		{"decode-only format", "/api/optimize?id=" + id + "&format=tiff", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get("http://localhost:" + port + tt.path) //nolint:noctx // integration test
			if err != nil {
				t.Fatalf("GET %s: %v", tt.path, err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.want {
				t.Errorf("status %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

// Ladder rows carry their quality in a closure rather than a data- attribute, so
// clicking one has to be exercised to know the wiring still holds.
func TestWebFrontend_ImageLadderRowSelectsQuality(t *testing.T) {
	const port = "19900"
	startImageServer(t, port)

	browser := newBrowser(t)
	page := browser.MustPage(fmt.Sprintf("http://localhost:%s/", port))
	page.MustWaitLoad()
	waitForElement(t, page, "#ladderBody tr")

	rows := page.MustElements("#ladderBody tr")
	if len(rows) < 2 {
		t.Fatalf("got %d ladder rows, want several to pick from", len(rows))
	}

	// The last rung is the lowest quality, so it differs from the default.
	last := rows[len(rows)-1]
	want := last.MustElements("td")[0].MustText()

	last.MustClick()

	deadline := time.Now().Add(10 * time.Second)
	var got string
	for time.Now().Before(deadline) {
		got = page.MustElement("#quality").MustProperty("value").String()
		if got == want {
			break
		}
		time.Sleep(150 * time.Millisecond)
	}
	if got != want {
		t.Fatalf("quality slider = %q after clicking the %q rung, want %q", got, want, want)
	}

	if shown := page.MustElement("#qualityOut").MustText(); shown != want {
		t.Errorf("quality readout = %q, want %q", shown, want)
	}
	if src := page.MustElement("#afterImg").MustProperty("src").String(); !strings.HasPrefix(src, "data:image/") {
		t.Errorf("optimized preview src = %.30q, want it re-rendered as a data url", src)
	}
}

// Launched with no file, the tool must open on the shared upload screen rather
// than a half-populated optimizer.
func TestWebFrontend_ImageUploadScreen(t *testing.T) {
	const port = "19899"

	cmd := exec.Command(unumBin, "image", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf(imageStartServerErr, err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	resp := waitForServer(t, fmt.Sprintf("http://localhost:%s/", port))
	_ = resp.Body.Close()

	browser := newBrowser(t)
	page := browser.MustPage(fmt.Sprintf("http://localhost:%s/", port))
	page.MustWaitLoad()

	// The structure the json, diff and diagram web UIs all share.
	for _, selector := range []string{
		"#upload-panel", "#drop-area.paste-area", ".browse-btn", ".paste-filename", ".upload-error",
	} {
		waitForElement(t, page, selector)
	}

	if hints := waitForElement(t, page, "#status-hints").MustText(); !strings.Contains(hints, "browse") {
		t.Errorf("status hints = %q, want it to describe drop/paste/browse", hints)
	}

	// The optimizer stays hidden until something is loaded.
	display := page.MustEval(`() => getComputedStyle(document.getElementById('layout')).display`).String()
	if display != "none" {
		t.Errorf("layout display = %q, want none before an image is loaded", display)
	}
}

func TestWebFrontend_ImageUILoadsAndOptimizes(t *testing.T) {
	const port = "19898"
	startImageServer(t, port)

	browser := newBrowser(t)
	page := browser.MustPage(fmt.Sprintf("http://localhost:%s/", port))
	page.MustWaitLoad()

	// The launch image loads itself, so the before/after panes appear without
	// any interaction.
	waitForElement(t, page, "#panes:not(.hidden)")
	waitForElement(t, page, "#ladderBody tr")

	if rows := page.MustElements("#ladderBody tr"); len(rows) == 0 {
		t.Error("expected quality ladder rows")
	}

	after := waitForElement(t, page, "#afterImg")
	src := after.MustProperty("src").String()
	if !strings.HasPrefix(src, "data:image/") {
		t.Errorf("optimized preview src = %.40q, want a data url", src)
	}

	// [ new ] appears once loaded and returns to the upload screen, matching the
	// sibling tools.
	waitForElement(t, page, "#new-btn").MustClick()
	waitForElement(t, page, "#upload-panel")

	display := page.MustEval(`() => getComputedStyle(document.getElementById('layout')).display`).String()
	if display != "none" {
		t.Errorf("layout display = %q after [ new ], want none", display)
	}
}
