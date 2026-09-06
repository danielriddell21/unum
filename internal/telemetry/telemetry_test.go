package telemetry

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/danielriddell21/unum/internal/config"
)

const (
	testVersion = "v0.0.1"
	closeErrFmt = "Close: %v"
)

func boolPtr(v bool) *bool { return &v }

func TestInit_Disabled(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("UMAMI_URL", "")
	t.Setenv("UMAMI_WEBSITE_ID", "")
	t.Setenv("DO_NOT_TRACK", "")
	t.Setenv("UNUM_NO_TELEMETRY", "")

	tel := Init(config.Config{}, "test", testVersion)
	if tel == nil {
		t.Fatal("Init should return non-nil even when disabled")
	}

	// Methods must not panic.
	tel.Tracer()
	tel.Meter()
	tel.TrackEvent("test", "/test", nil)
	tel.Disable()
	if err := tel.Close(context.Background()); err != nil {
		t.Errorf(closeErrFmt, err)
	}
}

func TestInit_DONotTrack(t *testing.T) {
	t.Setenv("DO_NOT_TRACK", "1")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")
	t.Setenv("UMAMI_URL", "")
	t.Setenv("UMAMI_WEBSITE_ID", "")

	tel := Init(config.Config{}, "test", testVersion)
	// Should not panic and should be effectively a no-op.
	tel.Tracer()
	tel.Meter()
	if err := tel.Close(context.Background()); err != nil {
		t.Errorf(closeErrFmt, err)
	}
}

func TestInit_ConfigDisabled(t *testing.T) {
	t.Setenv("DO_NOT_TRACK", "")
	t.Setenv("UNUM_NO_TELEMETRY", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")
	t.Setenv("UMAMI_URL", "")
	t.Setenv("UMAMI_WEBSITE_ID", "")

	cfg := config.Config{Telemetry: boolPtr(false)}
	tel := Init(cfg, "test", testVersion)
	// Should produce no-op providers since config says disabled.
	tel.Tracer()
	tel.Meter()
	if err := tel.Close(context.Background()); err != nil {
		t.Errorf(closeErrFmt, err)
	}
}

func TestDisable_SwitchesToNoop(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("UMAMI_URL", "")
	t.Setenv("UMAMI_WEBSITE_ID", "")
	t.Setenv("DO_NOT_TRACK", "")
	t.Setenv("UNUM_NO_TELEMETRY", "")

	tel := Init(config.Config{}, "test", testVersion)
	tel.Disable()

	// All methods should still be safe.
	tel.Tracer()
	tel.Meter()
	tel.TrackEvent("test", "/test", nil)
}

func TestNilTelemetry_Safe(t *testing.T) {
	var tel *Telemetry

	// None of these should panic.
	tel.Tracer()
	tel.Meter()
	tel.TrackEvent("test", "/test", nil)
	tel.Disable()
	if err := tel.Close(context.Background()); err != nil {
		t.Errorf(closeErrFmt, err)
	}
}

func TestUmami_Track(t *testing.T) {
	var calls atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			return
		}

		var payload umamiPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("unmarshal: %v", err)
			return
		}

		if payload.Payload.Name != "test_event" {
			t.Errorf("event name=%q, want \"test_event\"", payload.Payload.Name)
		}
		if payload.Payload.Website != "test-site-id" {
			t.Errorf("website=%q, want \"test-site-id\"", payload.Payload.Website)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	u := NewUmami(srv.URL, "test-site-id", "test.example.com")
	u.Track("test_event", "/api/tree", map[string]string{"tool": "json"})

	// Wait for the fire-and-forget goroutine.
	time.Sleep(200 * time.Millisecond)

	if c := calls.Load(); c != 1 {
		t.Errorf("expected 1 call, got %d", c)
	}
}

func TestUmami_Disabled(t *testing.T) {
	var calls atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	u := NewUmami("", "", "")
	u.Track("test", "/test", nil)

	time.Sleep(100 * time.Millisecond)

	if c := calls.Load(); c != 0 {
		t.Errorf("disabled client should not call server, got %d calls", c)
	}
}

func TestUmami_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	u := NewUmami(srv.URL, "test-id", "test.example.com")
	// Should not panic on server error.
	u.Track("test", "/test", nil)

	time.Sleep(100 * time.Millisecond)
}

func TestConcurrentTrack(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("UMAMI_URL", "")
	t.Setenv("UMAMI_WEBSITE_ID", "")
	t.Setenv("DO_NOT_TRACK", "")
	t.Setenv("UNUM_NO_TELEMETRY", "")

	tel := Init(config.Config{}, "test", testVersion)
	done := make(chan struct{})

	for i := range 10 {
		go func(n int) {
			tel.Tracer()
			tel.Meter()
			tel.TrackEvent("event", "/test", nil)
			if n%2 == 0 {
				tel.Disable()
			}
			done <- struct{}{}
		}(i)
	}

	for range 10 {
		<-done
	}

	if err := tel.Close(context.Background()); err != nil {
		t.Errorf(closeErrFmt, err)
	}
}

func TestNew_NoopBeforeStart(t *testing.T) {
	tel := New()
	if tel.M.Invocations == nil {
		t.Fatal("metrics must be usable before Start")
	}

	// Must not panic: commands hold this before PersistentPreRun runs Start.
	tel.M.Invocations.Add(context.Background(), 1)
	tel.Tracer()
	tel.Meter()
	tel.TrackEvent("test", "/test", nil)
	if err := tel.Close(context.Background()); err != nil {
		t.Errorf(closeErrFmt, err)
	}
}

func TestStart_DisabledLeavesNoopUsable(t *testing.T) {
	t.Setenv("DO_NOT_TRACK", "1")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")

	tel := New()
	tel.Start(config.Config{}, "test", testVersion)
	tel.M.Invocations.Add(context.Background(), 1)
	if err := tel.Close(context.Background()); err != nil {
		t.Errorf(closeErrFmt, err)
	}
}

func TestEndpoint_PrefersEnv(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://example.invalid:4318")
	if got := Endpoint(); got != "http://example.invalid:4318" {
		t.Errorf("Endpoint()=%q", got)
	}

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	if got := Endpoint(); got != otelEndpoint {
		t.Errorf("Endpoint()=%q, want build-time default %q", got, otelEndpoint)
	}
}
