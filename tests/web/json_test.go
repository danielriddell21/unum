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

const (
	jsonTreeURLFmt     = "http://localhost:%s/api/tree"
	jsonUploadURLFmt   = "http://localhost:%s/api/upload"
	jsonContentType    = "application/json"
	jsonStartServerErr = "start server: %v"
	jsonUploadNotJSON  = "upload response not JSON: %v\nbody: %s"
)

// TestJSONWebContainerMode checks the json --web server starts and returns 204
// on /api/tree
func TestJSONWebContainerMode(t *testing.T) {
	const port = "19883"
	cmd := exec.Command(unumBin, "json", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port, "UNUM_ENV=test")
	if err := cmd.Start(); err != nil {
		t.Fatalf(jsonStartServerErr, err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	resp := waitForServer(t, fmt.Sprintf(jsonTreeURLFmt, port))
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
	cmd.Env = append(os.Environ(), "PORT="+port, "UNUM_ENV=test")
	if err := cmd.Start(); err != nil {
		t.Fatalf(jsonStartServerErr, err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	_ = waitForServer(t, fmt.Sprintf(jsonTreeURLFmt, port))

	content, err := os.ReadFile("testdata/sample.json")
	if err != nil {
		t.Fatal(err)
	}
	uploadPayload, _ := json.Marshal(map[string]string{
		"filename": "sample.json",
		"content":  string(content),
	})
	upResp, err := http.Post( //nolint:noctx // integration test hitting local server; context not needed in tests
		fmt.Sprintf(jsonUploadURLFmt, port),
		jsonContentType,
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
		t.Fatalf(jsonUploadNotJSON, err, upBody)
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

	tree, ok := payload["tree"].(map[string]any)
	if !ok {
		t.Error("payload.tree is not an object")
		return
	}
	if kind, _ := tree["kind"].(string); kind != "object" {
		t.Errorf("tree.kind = %q, want object", kind)
	}
}

// TestJSONWebAPI_Upload posts JSON content to /api/upload, gets back a key,
// then fetches /api/tree?key=... and verifies the payload.
func TestJSONWebAPI_Upload(t *testing.T) {
	const port = "19869"
	cmd := exec.Command(unumBin, "json", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port, "UNUM_ENV=test")
	if err := cmd.Start(); err != nil {
		t.Fatalf(jsonStartServerErr, err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	_ = waitForServer(t, fmt.Sprintf(jsonTreeURLFmt, port))

	uploadBody := `{"filename":"test.json","content":"{\"hello\":\"world\"}"}`
	resp, err := http.Post( //nolint:noctx // integration test hitting local server; context not needed in tests
		fmt.Sprintf(jsonUploadURLFmt, port),
		jsonContentType,
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
		t.Fatalf(jsonUploadNotJSON, err, body)
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
	cmd.Env = append(os.Environ(), "PORT="+port, "UNUM_ENV=test")
	if err := cmd.Start(); err != nil {
		t.Fatalf(jsonStartServerErr, err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	resp := waitForServer(t, fmt.Sprintf("http://localhost:%s/", port))
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("<html")) {
		t.Errorf("GET /: expected HTML body, got: %s", body[:min(len(body), 200)])
	}
}

// TestWebFrontend_JSONUIExploreLoads uses the browser upload UI to load
// sample.json via the file input and verifies the JSON tree renders.
// Requires Chromium to be installed.
func TestWebFrontend_JSONUIExploreLoads(t *testing.T) {
	const port = "19858"
	cmd := exec.Command(unumBin, "json", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port, "UNUM_ENV=test")
	if err := cmd.Start(); err != nil {
		t.Fatalf(jsonStartServerErr, err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	_ = waitForServer(t, fmt.Sprintf(jsonTreeURLFmt, port))

	absPath, err := filepath.Abs("testdata/sample.json")
	if err != nil {
		t.Fatal(err)
	}

	browser := newBrowser(t)
	page := browser.MustPage(fmt.Sprintf("http://localhost:%s/", port))
	page.MustWaitLoad()

	// Set the hidden file input — JS fires FileReader then auto-uploads.
	page.MustElement("#json-file-input").MustSetFiles(absPath)

	// Wait for auto-upload to complete and tree to render.
	waitForElement(t, page, ".tree-node")

	nodes := page.MustElements(".tree-node")
	if len(nodes) == 0 {
		t.Error("expected tree nodes to render after upload")
	}
}

func postJSONQuery(t *testing.T, port, key, expr string) (resp struct{ Result, Error string }, elapsed time.Duration) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"expr": expr})
	start := time.Now()
	r, err := http.Post( //nolint:noctx // integration test hitting local server; context not needed in tests
		fmt.Sprintf("http://localhost:%s/api/query?key=%s", port, key),
		jsonContentType,
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("POST /api/query: %v", err)
	}
	defer func() { _ = r.Body.Close() }()
	raw, _ := io.ReadAll(r.Body)
	elapsed = time.Since(start)
	var decoded struct {
		Result string `json:"result"`
		Error  string `json:"error"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("query response not JSON: %v\nbody: %s", err, raw)
	}
	resp.Result, resp.Error = decoded.Result, decoded.Error
	return resp, elapsed
}

// TestJSONWebAPI_QueryBoundedAndReusable checks runaway queries are cut off
// with a distinct error and the server keeps answering afterwards.
func TestJSONWebAPI_QueryBoundedAndReusable(t *testing.T) {
	const port = "19871"
	cmd := exec.Command(unumBin, "json", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port, "UNUM_ENV=test")
	if err := cmd.Start(); err != nil {
		t.Fatalf(jsonStartServerErr, err)
	}
	defer func() { _ = cmd.Process.Kill() }()
	_ = waitForServer(t, fmt.Sprintf(jsonTreeURLFmt, port))

	uploadPayload, _ := json.Marshal(map[string]string{"filename": "a.json", "content": `{"a": [1, 2, 3]}`})
	upResp, err := http.Post( //nolint:noctx // integration test hitting local server; context not needed in tests
		fmt.Sprintf(jsonUploadURLFmt, port), jsonContentType, bytes.NewReader(uploadPayload))
	if err != nil {
		t.Fatalf("POST /api/upload: %v", err)
	}
	upBody, _ := io.ReadAll(upResp.Body)
	_ = upResp.Body.Close()
	var up map[string]string
	if err := json.Unmarshal(upBody, &up); err != nil || up["key"] == "" {
		t.Fatalf(jsonUploadNotJSON, err, upBody)
	}
	key := up["key"]

	resp, elapsed := postJSONQuery(t, port, key, "[repeat(1)]")
	if resp.Error != "query timed out" {
		t.Errorf("runaway query error=%q, want %q", resp.Error, "query timed out")
	}
	if elapsed > 15*time.Second {
		t.Errorf("runaway query took %s, want it cut off near 5s", elapsed)
	}

	resp, _ = postJSONQuery(t, port, key, `"x" * 20000000`)
	if !strings.Contains(resp.Error, "output truncated") {
		t.Errorf("oversized query error=%q, want output truncated", resp.Error)
	}

	resp, _ = postJSONQuery(t, port, key, ".a | length")
	if resp.Error != "" || strings.TrimSpace(resp.Result) != "3" {
		t.Errorf("follow-up query result=%q error=%q, want 3", resp.Result, resp.Error)
	}
}
