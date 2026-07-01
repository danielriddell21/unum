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

type Telemetry struct {
	Umami    *UmamiClient
	M        Metrics
	tracer   otrace.Tracer
	meter    ometric.Meter
	shutdown func(context.Context) error
	disabled atomic.Bool
}

func Init(cfg config.Config, serviceName, version string) *Telemetry {
	t := &Telemetry{}

	shutdown, err := initOTel(cfg, serviceName, version)
	if err != nil {
		t.meter = noop.NewMeterProvider().Meter(serviceName)
		t.tracer = nooptrace.NewTracerProvider().Tracer(serviceName)
		t.M = newMetrics(t.meter)
		t.shutdown = func(context.Context) error { return nil }
		return t
	}

	t.shutdown = shutdown
	t.tracer = otel.GetTracerProvider().Tracer(serviceName)
	t.meter = otel.GetMeterProvider().Meter(serviceName)
	t.M = newMetrics(t.meter)

	// Umami — only for web mode (env vars set in k8s).
	t.Umami = NewUmami(
		os.Getenv("UMAMI_URL"),
		os.Getenv("UMAMI_WEBSITE_ID"),
		os.Getenv("UMAMI_HOSTNAME"),
	)

	return t
}

func (t *Telemetry) Tracer() otrace.Tracer {
	if t == nil || t.disabled.Load() {
		return nooptrace.NewTracerProvider().Tracer("noop")
	}
	return t.tracer
}

func (t *Telemetry) Meter() ometric.Meter {
	if t == nil || t.disabled.Load() {
		return noop.NewMeterProvider().Meter("noop")
	}
	return t.meter
}

func (t *Telemetry) TrackEvent(event, pageURL string, props map[string]string) {
	if t == nil || t.disabled.Load() || t.Umami == nil {
		return
	}
	t.Umami.Track(event, pageURL, props)
}

func (t *Telemetry) Disable() {
	if t == nil {
		return
	}
	t.disabled.Store(true)
}

func (t *Telemetry) Close(ctx context.Context) error {
	if t == nil || t.shutdown == nil {
		return nil
	}
	return t.shutdown(ctx)
}
