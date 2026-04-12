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
)

var (
	diffA     = filepath.Join("testdata", "diff-a.txt")
	diffB     = filepath.Join("testdata", "diff-b.txt")
	diffAJSON = filepath.Join("testdata", "diff-a.json")
	diffBJSON = filepath.Join("testdata", "diff-b.json")
	diffATF   = filepath.Join("testdata", "diff-a.tfplan.json")
)

// TestDiffWebContainerMode checks the diff --web server starts and returns 204
// on GET /api/diff.
func TestDiffWebContainerMode(t *testing.T) {
	const port = "19877"
	cmd := exec.Command(unumBin, "diff", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	resp := waitForServer(t, fmt.Sprintf("http://localhost:%s/api/diff", port))
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("GET /api/diff: status %d, want 204", resp.StatusCode)
	}
}

// TestDiffWebAPI_PostTextDiffReturnsPayload sends two plain-text files via
// POST /api/diff and asserts the response contains hunks.
func TestDiffWebAPI_PostTextDiffReturnsPayload(t *testing.T) {
	const port = "19867"
	cmd := exec.Command(unumBin, "diff", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	resp := waitForServer(t, fmt.Sprintf("http://localhost:%s/api/diff", port))
	_ = resp.Body.Close()

	contentA, err := os.ReadFile(diffA)
	if err != nil {
		t.Fatal(err)
	}
	contentB, err := os.ReadFile(diffB)
	if err != nil {
		t.Fatal(err)
	}

	reqBody, _ := json.Marshal(map[string]string{
		"nameA":    "a.txt",
		"contentA": string(contentA),
		"nameB":    "b.txt",
		"contentB": string(contentB),
		"format":   "text",
	})

	postResp, err := http.Post( //nolint:noctx
		fmt.Sprintf("http://localhost:%s/api/diff", port),
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		t.Fatalf("POST /api/diff: %v", err)
	}
	defer func() { _ = postResp.Body.Close() }()

	if postResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(postResp.Body)
		t.Fatalf("POST /api/diff: status %d, want 200\nbody: %s", postResp.StatusCode, body)
	}

	body, err := io.ReadAll(postResp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("response not JSON: %v\nbody: %s", err, body)
	}

	for _, field := range []string{"fileA", "fileB", "format", "hunks"} {
		if _, ok := payload[field]; !ok {
			t.Errorf("payload missing %q; keys: %v", field, keys(payload))
		}
	}

	if format, _ := payload["format"].(string); format != "text" {
		t.Errorf("format = %q, want text", format)
	}

	if hunks, ok := payload["hunks"].([]any); !ok || len(hunks) == 0 {
		t.Errorf("expected non-empty hunks array; payload: %s", body)
	}
}

// TestDiffWebAPI_PostJSONDiffReturnsSemanticTree sends two JSON files and
// asserts the response includes a semantic tree and the json format label.
func TestDiffWebAPI_PostJSONDiffReturnsSemanticTree(t *testing.T) {
	const port = "19866"
	cmd := exec.Command(unumBin, "diff", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	resp := waitForServer(t, fmt.Sprintf("http://localhost:%s/api/diff", port))
	_ = resp.Body.Close()

	contentA, err := os.ReadFile(diffAJSON)
	if err != nil {
		t.Fatal(err)
	}
	contentB, err := os.ReadFile(diffBJSON)
	if err != nil {
		t.Fatal(err)
	}

	reqBody, _ := json.Marshal(map[string]string{
		"nameA":    "a.json",
		"contentA": string(contentA),
		"nameB":    "b.json",
		"contentB": string(contentB),
		"format":   "json",
	})

	postResp, err := http.Post( //nolint:noctx
		fmt.Sprintf("http://localhost:%s/api/diff", port),
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		t.Fatalf("POST /api/diff: %v", err)
	}
	defer func() { _ = postResp.Body.Close() }()

	if postResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(postResp.Body)
		t.Fatalf("POST /api/diff: status %d, want 200\nbody: %s", postResp.StatusCode, body)
	}

	body, err := io.ReadAll(postResp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("response not JSON: %v\nbody: %s", err, body)
	}

	if format, _ := payload["format"].(string); format != "json" {
		t.Errorf("format = %q, want json", format)
	}
	if _, ok := payload["tree"]; !ok {
		t.Errorf("JSON diff payload missing 'tree'; keys: %v", keys(payload))
	}
}

// TestDiffWebAPI_PostMissingContentReturns400 checks that omitting contentB
// results in a 400 response.
func TestDiffWebAPI_PostMissingContentReturns400(t *testing.T) {
	const port = "19865"
	cmd := exec.Command(unumBin, "diff", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	resp := waitForServer(t, fmt.Sprintf("http://localhost:%s/api/diff", port))
	_ = resp.Body.Close()

	badBody := strings.NewReader(`{"nameA":"a.txt","contentA":"hello"}`)
	postResp, err := http.Post( //nolint:noctx
		fmt.Sprintf("http://localhost:%s/api/diff", port),
		"application/json",
		badBody,
	)
	if err != nil {
		t.Fatalf("POST /api/diff: %v", err)
	}
	defer func() { _ = postResp.Body.Close() }()

	if postResp.StatusCode != http.StatusBadRequest {
		t.Errorf("missing contentB: status %d, want 400", postResp.StatusCode)
	}
}

// TestDiffWebAPI_PostTerraformDiffReturnsSemanticTree sends a Terraform plan
// file via POST /api/diff and asserts the response includes a semantic tree
// and the terraform format label.
func TestDiffWebAPI_PostTerraformDiffReturnsSemanticTree(t *testing.T) {
	const port = "19856"
	cmd := exec.Command(unumBin, "diff", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	resp := waitForServer(t, fmt.Sprintf("http://localhost:%s/api/diff", port))
	_ = resp.Body.Close()

	content, err := os.ReadFile(diffATF)
	if err != nil {
		t.Fatal(err)
	}

	reqBody, _ := json.Marshal(map[string]string{
		"nameA":    "plan.tfplan.json",
		"contentA": string(content),
		"nameB":    "plan.tfplan.json",
		"contentB": string(content),
		"format":   "terraform",
	})

	postResp, err := http.Post( //nolint:noctx
		fmt.Sprintf("http://localhost:%s/api/diff", port),
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		t.Fatalf("POST /api/diff: %v", err)
	}
	defer func() { _ = postResp.Body.Close() }()

	if postResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(postResp.Body)
		t.Fatalf("POST /api/diff: status %d, want 200\nbody: %s", postResp.StatusCode, body)
	}

	body, err := io.ReadAll(postResp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("response not JSON: %v\nbody: %s", err, body)
	}

	if format, _ := payload["format"].(string); format != "terraform" {
		t.Errorf("format = %q, want terraform", format)
	}
	if _, ok := payload["tree"]; !ok {
		t.Errorf("terraform diff payload missing 'tree'; keys: %v", keys(payload))
	}
}

// TestWebFrontend_DiffUIRendersAndToggles loads two text files via the browser
// upload form, verifies the diff renders with added/removed lines, then
// switches to split view and asserts the panel class changes.
// Requires Chromium to be installed.
func TestWebFrontend_DiffUIRendersAndToggles(t *testing.T) {
	const port = "19859"
	cmd := exec.Command(unumBin, "diff", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	_ = waitForServer(t, fmt.Sprintf("http://localhost:%s/api/diff", port))

	browser := newBrowser(t)
	page := browser.MustPage(fmt.Sprintf("http://localhost:%s/", port))
	page.MustWaitLoad()

	// Set the hidden file inputs directly — the JS fires FileReader on change.
	absA, err := filepath.Abs(diffA)
	if err != nil {
		t.Fatal(err)
	}
	absB, err := filepath.Abs(diffB)
	if err != nil {
		t.Fatal(err)
	}
	page.MustElement("#file-input-a").MustSetFiles(absA)
	page.MustElement("#file-input-b").MustSetFiles(absB)

	// Wait for FileReader.onload — JS adds class "loaded" to each drop zone.
	waitForElement(t, page, "#drop-a.loaded")
	waitForElement(t, page, "#drop-b.loaded")

	page.MustElement("#up-submit").MustClick()

	// Wait for diff to render.
	waitForElement(t, page, ".diff-line")

	added := page.MustElements(".diff-added")
	removed := page.MustElements(".diff-removed")
	if len(added) == 0 && len(removed) == 0 {
		t.Error("expected added or removed lines in rendered diff")
	}

	// Switch to split view — JS replaces #toggle-view placeholder with
	// #view-modes buttons and adds is-split class to #diff-panel.
	waitForElement(t, page, ".view-mode-btn[data-mode=\"split\"]").MustClick()
	waitForElement(t, page, "#diff-panel.is-split")
}
