package web_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const publicWarning = "reachable from the network"

func startJSONWeb(t *testing.T, port string, env ...string) *bytes.Buffer {
	t.Helper()
	cmd := exec.Command(unumBin, "json", "--web") //nolint:noctx // test runner invoking compiled binary
	cmd.Env = append(os.Environ(), append([]string{"PORT=" + port}, env...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill() })

	resp := waitForServer(t, fmt.Sprintf("http://localhost:%s/", port))
	_ = resp.Body.Close()
	return &stderr
}

func TestWebBind_PortAloneStaysOnLoopback(t *testing.T) {
	stderr := startJSONWeb(t, "19901", "UNUM_ENV=", "UNUM_BIND=")

	if strings.Contains(stderr.String(), publicWarning) {
		t.Errorf("PORT alone must not bind beyond loopback, stderr:\n%s", stderr.String())
	}
}

func TestWebBind_HostedModeWarnsAndBindsPublicly(t *testing.T) {
	stderr := startJSONWeb(t, "19902", "UNUM_ENV=test")

	if !strings.Contains(stderr.String(), publicWarning) {
		t.Errorf("hosted bind should warn on stderr, got:\n%s", stderr.String())
	}
}

func TestWebBind_ExplicitBindWarns(t *testing.T) {
	stderr := startJSONWeb(t, "19903", "UNUM_ENV=", "UNUM_BIND=0.0.0.0")

	if !strings.Contains(stderr.String(), publicWarning) {
		t.Errorf("UNUM_BIND=0.0.0.0 should warn on stderr, got:\n%s", stderr.String())
	}
}

func TestWebBind_MetricsOnlyWhenPublic(t *testing.T) {
	// The "/" catch-all answers any unregistered path with the index page, so
	// the Prometheus exposition format is what distinguishes a real handler.
	startJSONWeb(t, "19904", "UNUM_ENV=", "UNUM_BIND=")
	if body := getBody(t, "19904", "/metrics"); strings.Contains(body, "# HELP") {
		t.Error("/metrics should not be registered on a loopback server")
	}

	startJSONWeb(t, "19905", "UNUM_ENV=test")
	if body := getBody(t, "19905", "/metrics"); !strings.Contains(body, "# HELP") {
		t.Error("/metrics should be registered in hosted mode")
	}
}

func TestWebBind_MetricsOptIn(t *testing.T) {
	startJSONWeb(t, "19906", "UNUM_ENV=", "UNUM_BIND=", "UNUM_METRICS=1")
	if body := getBody(t, "19906", "/metrics"); !strings.Contains(body, "# HELP") {
		t.Error("UNUM_METRICS should register /metrics on a loopback server")
	}
}

func getBody(t *testing.T, port, path string) string {
	t.Helper()
	resp, err := httpGet(fmt.Sprintf("http://localhost:%s%s", port, path))
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
}
