package telemetry

import (
	"context"
	"os"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	ometric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	otrace "go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"

	"github.com/danielriddell21/unum/internal/config"
)

// Telemetry is the unified facade for all telemetry systems.
// All methods are safe to call on a nil or disabled instance.
type Telemetry struct {
	Umami    *UmamiClient
	tracer   otrace.Tracer
	meter    ometric.Meter
	shutdown func(context.Context) error
	disabled atomic.Bool
}

// Init creates a Telemetry instance. If telemetry is disabled (config, env, or
// missing endpoint), returns a fully functional but no-op instance.
func Init(cfg config.Config, serviceName, version string) *Telemetry {
	t := &Telemetry{}

	shutdown, err := initOTel(cfg, serviceName, version)
	if err != nil {
		t.tracer = nooptrace.NewTracerProvider().Tracer(serviceName)
		t.meter = noop.NewMeterProvider().Meter(serviceName)
		t.shutdown = func(context.Context) error { return nil }
		return t
	}

	t.shutdown = shutdown
	t.tracer = otel.GetTracerProvider().Tracer(serviceName)
	t.meter = otel.GetMeterProvider().Meter(serviceName)

	// Umami — only for web mode (env vars set in k8s).
	t.Umami = NewUmami(
		os.Getenv("UMAMI_URL"),
		os.Getenv("UMAMI_WEBSITE_ID"),
		os.Getenv("UMAMI_HOSTNAME"),
	)

	return t
}

// Tracer returns the OTel tracer. Returns a no-op tracer if disabled.
func (t *Telemetry) Tracer() otrace.Tracer {
	if t == nil || t.disabled.Load() {
		return nooptrace.NewTracerProvider().Tracer("noop")
	}
	return t.tracer
}

// Meter returns the OTel meter. Returns a no-op meter if disabled.
func (t *Telemetry) Meter() ometric.Meter {
	if t == nil || t.disabled.Load() {
		return noop.NewMeterProvider().Meter("noop")
	}
	return t.meter
}

// TrackEvent sends a named event to Umami (web mode only). No-op if Umami
// is not configured or telemetry is disabled.
func (t *Telemetry) TrackEvent(event, pageURL string, props map[string]string) {
	if t == nil || t.disabled.Load() || t.Umami == nil {
		return
	}
	t.Umami.Track(event, pageURL, props)
}

// Disable switches telemetry to no-op mode. Called when --no-telemetry is set.
func (t *Telemetry) Disable() {
	if t == nil {
		return
	}
	t.disabled.Store(true)
}

// Close flushes pending telemetry data. Should be called before process exit.
func (t *Telemetry) Close(ctx context.Context) error {
	if t == nil || t.shutdown == nil {
		return nil
	}
	return t.shutdown(ctx)
}
