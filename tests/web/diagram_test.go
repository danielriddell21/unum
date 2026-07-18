package web_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const renderStartServerErr = "start server: %v"

func startRenderServer(t *testing.T, port string) {
	t.Helper()
	cmd := exec.Command(unumBin, "diagram", "testdata/sample.d2", "--web")
	cmd.Env = append(os.Environ(), "PORT="+port)
	if err := cmd.Start(); err != nil {
		t.Fatalf(renderStartServerErr, err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill() })
}

func TestRenderWebContainerMode(t *testing.T) {
	const port = "19571"
	startRenderServer(t, port)
	resp := waitForServer(t, fmt.Sprintf("http://localhost:%s/", port))
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /: status %d, want 200", resp.StatusCode)
	}
}

func TestRenderWebAPI_D2SVG(t *testing.T) {
	const port = "19572"
	startRenderServer(t, port)
	waitForServer(t, fmt.Sprintf("http://localhost:%s/", port)).Body.Close()

	resp, err := http.Post( //nolint:noctx // test client call
		fmt.Sprintf("http://localhost:%s/api/render?lang=d2&format=svg", port),
		"text/plain", strings.NewReader("x -> y"))
	if err != nil {
		t.Fatalf("POST /api/render: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "<svg") {
		t.Errorf("expected <svg in response:\n%s", body)
	}
}

func TestRenderWebAPI_Drawio(t *testing.T) {
	const port = "19573"
	startRenderServer(t, port)
	waitForServer(t, fmt.Sprintf("http://localhost:%s/", port)).Body.Close()

	resp, err := http.Post( //nolint:noctx // test client call
		fmt.Sprintf("http://localhost:%s/api/render?lang=d2&format=drawio", port),
		"text/plain", strings.NewReader("a -> b"))
	if err != nil {
		t.Fatalf("POST /api/render: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "mxGraphModel") {
		t.Errorf("expected mxGraphModel in response:\n%s", body)
	}
}
