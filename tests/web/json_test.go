package web_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestJSONWebContainerMode checks the json --web server starts and returns 204
// on /api/tree
func TestJSONWebContainerMode(t *testing.T) {
	const port = "19883"
	cmd := exec.Command(unumBin, "json", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	resp := waitForServer(t, fmt.Sprintf("http://localhost:%s/api/tree", port))
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("GET /api/tree: status %d, want 204", resp.StatusCode)
	}
}

// TestJSONWebAPI_TreeBodyHasExpectedFields uploads a file via /api/upload and
// asserts the payload returned by /api/tree?key=... contains expected fields.
func TestJSONWebAPI_TreeBodyHasExpectedFields(t *testing.T) {
	const port = "19870"
	cmd := exec.Command(unumBin, "json", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	_ = waitForServer(t, fmt.Sprintf("http://localhost:%s/api/tree", port))

	content, err := os.ReadFile("testdata/sample.json")
	if err != nil {
		t.Fatal(err)
	}
	uploadPayload, _ := json.Marshal(map[string]string{
		"filename": "sample.json",
		"content":  string(content),
	})
	upResp, err := http.Post( //nolint:noctx // integration test hitting local server; context not needed in tests
		fmt.Sprintf("http://localhost:%s/api/upload", port),
		"application/json",
		bytes.NewReader(uploadPayload),
	)
	if err != nil {
		t.Fatalf("POST /api/upload: %v", err)
	}
	upBody, _ := io.ReadAll(upResp.Body)
	_ = upResp.Body.Close()
	if upResp.StatusCode != http.StatusOK {
		t.Fatalf("POST /api/upload: status %d\nbody: %s", upResp.StatusCode, upBody)
	}
	var upResult map[string]string
	if err := json.Unmarshal(upBody, &upResult); err != nil {
		t.Fatalf("upload response not JSON: %v\nbody: %s", err, upBody)
	}
	key := upResult["key"]
	if key == "" {
		t.Fatalf("upload returned no key: %s", upBody)
	}

	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/api/tree?key=%s", port, key)) //nolint:noctx // integration test hitting local server; context not needed in tests
	if err != nil {
		t.Fatalf("GET /api/tree?key=: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/tree: status %d, want 200", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("body is not JSON: %v\nbody: %s", err, body)
	}

	for _, field := range []string{"tree", "yaml", "typegen", "merkleRoot", "nodeCount"} {
		if _, ok := payload[field]; !ok {
			t.Errorf("payload missing field %q; keys: %v", field, keys(payload))
		}
	}

	if tree, ok := payload["tree"].(map[string]any); ok {
		if kind, _ := tree["kind"].(string); kind != "object" {
			t.Errorf("tree.kind = %q, want object", kind)
		}
	} else {
		t.Error("payload.tree is not an object")
	}
}

// TestJSONWebAPI_Upload posts JSON content to /api/upload, gets back a key,
// then fetches /api/tree?key=... and verifies the payload.
func TestJSONWebAPI_Upload(t *testing.T) {
	const port = "19869"
	cmd := exec.Command(unumBin, "json", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	_ = waitForServer(t, fmt.Sprintf("http://localhost:%s/api/tree", port))

	uploadBody := `{"filename":"test.json","content":"{\"hello\":\"world\"}"}`
	resp, err := http.Post( //nolint:noctx // integration test hitting local server; context not needed in tests
		fmt.Sprintf("http://localhost:%s/api/upload", port),
		"application/json",
		strings.NewReader(uploadBody),
	)
	if err != nil {
		t.Fatalf("POST /api/upload: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /api/upload: status %d, want 200", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read upload response: %v", err)
	}

	var uploadResp map[string]string
	if err := json.Unmarshal(body, &uploadResp); err != nil {
		t.Fatalf("upload response not JSON: %v\nbody: %s", err, body)
	}
	key := uploadResp["key"]
	if key == "" {
		t.Fatalf("upload response missing key: %s", body)
	}

	treeResp, err := http.Get(fmt.Sprintf("http://localhost:%s/api/tree?key=%s", port, key)) //nolint:noctx // integration test hitting local server; context not needed in tests
	if err != nil {
		t.Fatalf("GET /api/tree?key=: %v", err)
	}
	defer func() { _ = treeResp.Body.Close() }()

	if treeResp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/tree?key=: status %d, want 200", treeResp.StatusCode)
	}
	treeBody, _ := io.ReadAll(treeResp.Body)

	var treePayload map[string]any
	if err := json.Unmarshal(treeBody, &treePayload); err != nil {
		t.Fatalf("tree payload not JSON: %v\nbody: %s", err, treeBody)
	}
	if _, ok := treePayload["tree"]; !ok {
		t.Errorf("tree payload missing 'tree' field; keys: %v", keys(treePayload))
	}
}

// TestJSONWebAPI_IndexHTML checks GET / returns an HTML page.
func TestJSONWebAPI_IndexHTML(t *testing.T) {
	const port = "19868"
	cmd := exec.Command(unumBin, "json", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	resp := waitForServer(t, fmt.Sprintf("http://localhost:%s/", port))
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("<html")) {
		t.Errorf("GET /: expected HTML body, got: %s", body[:min(len(body), 200)])
	}
}

// TestWebFrontend_JSONUIExploreLoads uploads sample.json via the HTTP API,
// navigates the browser to the explore page, and verifies the JSON tree
// renders. Requires Chromium to be installed.
func TestWebFrontend_JSONUIExploreLoads(t *testing.T) {
	const port = "19858"
	cmd := exec.Command(unumBin, "json", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	_ = waitForServer(t, fmt.Sprintf("http://localhost:%s/api/tree", port))

	// Upload sample.json to get a session key.
	content, err := os.ReadFile("testdata/sample.json")
	if err != nil {
		t.Fatal(err)
	}
	uploadPayload, _ := json.Marshal(map[string]string{
		"filename": "sample.json",
		"content":  string(content),
	})
	upResp, err := http.Post( //nolint:noctx // integration test hitting local server; context not needed in tests
		fmt.Sprintf("http://localhost:%s/api/upload", port),
		"application/json",
		bytes.NewReader(uploadPayload),
	)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	upBody, _ := io.ReadAll(upResp.Body)
	_ = upResp.Body.Close()
	var upResult map[string]string
	if err := json.Unmarshal(upBody, &upResult); err != nil {
		t.Fatalf("upload response not JSON: %v\nbody: %s", err, upBody)
	}
	key := upResult["key"]
	if key == "" {
		t.Fatalf("upload returned no key: %s", upBody)
	}

	// Navigate to the explore page and verify the tree renders.
	browser := newBrowser(t)
	page := browser.MustPage(fmt.Sprintf("http://localhost:%s/explore?key=%s", port, key))
	page.MustWaitLoad()

	waitForElement(t, page, ".tree-node")

	nodes := page.MustElements(".tree-node")
	if len(nodes) == 0 {
		t.Error("expected tree nodes to render in explorer")
	}
}
