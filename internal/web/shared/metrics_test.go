package shared

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func metricsRegistered(t *testing.T, b Bind) bool {
	t.Helper()
	mux := http.NewServeMux()
	RegisterMetrics(mux, b)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	_, pattern := mux.Handler(req)
	return pattern != ""
}

func TestRegisterMetrics(t *testing.T) {
	tests := []struct {
		name     string
		bind     Bind
		envOptIn string
		want     bool
	}{
		{name: "loopback does not expose metrics", bind: Bind{Host: "localhost"}, want: false},
		{name: "public bind exposes metrics", bind: Bind{Host: "0.0.0.0", Public: true}, want: true},
		{name: "explicit opt-in on loopback", bind: Bind{Host: "localhost"}, envOptIn: "1", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("UNUM_METRICS", tt.envOptIn)
			if got := metricsRegistered(t, tt.bind); got != tt.want {
				t.Errorf("metrics registered=%v, want %v", got, tt.want)
			}
		})
	}
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w
	fn()
	_ = w.Close()
	os.Stderr = orig

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return string(out)
}

func TestPrintBindWarning(t *testing.T) {
	tests := []struct {
		name     string
		bind     Bind
		wantWarn bool
	}{
		{name: "loopback is silent", bind: Bind{Host: "localhost"}, wantWarn: false},
		{name: "public warns", bind: Bind{Host: "0.0.0.0", Public: true}, wantWarn: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStderr(t, func() { PrintBindWarning(tt.bind) })
			got := strings.Contains(out, "reachable from the network")
			if got != tt.wantWarn {
				t.Errorf("warned=%v, want %v (output %q)", got, tt.wantWarn, out)
			}
			if tt.wantWarn && !strings.Contains(out, tt.bind.Host) {
				t.Errorf("warning should name the host, got %q", out)
			}
		})
	}
}
