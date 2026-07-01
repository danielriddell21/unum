package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/danielriddell21/unum/internal/config"
)

var otelAuthToken string //nolint:gochecknoglobals // build-time injection

var otelEndpoint string //nolint:gochecknoglobals // build-time injection

const shutdownTimeout = 2 * time.Second

var debugLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{ //nolint:gochecknoglobals // package-level debug logger
	Level: slog.LevelDebug,
})).With("component", "telemetry")

func debugf(format string, args ...any) {
	if os.Getenv("UNUM_TELEMETRY_DEBUG") == "1" {
		debugLogger.Debug(fmt.Sprintf(format, args...))
	}
}

func initOTel(cfg config.Config, serviceName, version string) (func(context.Context) error, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = otelEndpoint // fall back to build-time default (set by GoReleaser)
	}
	if !cfg.TelemetryEnabled() {
		debugf("disabled (config/env)")
		return func(context.Context) error { return nil }, nil
	}
	if endpoint == "" {
		debugf("disabled (no endpoint)")
		return func(context.Context) error { return nil }, nil
	}
	debugf("init: service=%s version=%s endpoint=%s auth=%v", serviceName, version, endpoint, authToken() != "")

	ctx := context.Background()

	env := os.Getenv("UNUM_ENV")
	if env == "" {
		env = "development"
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
			attribute.String("deployment.environment", env),
			attribute.String("client.id", cfg.ClientID),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel resource: %w", err)
	}

	// Trace exporter — OTLP HTTP with optional bearer token.
	traceOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(endpoint),
	}
	if token := authToken(); token != "" {
		traceOpts = append(traceOpts, otlptracehttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + token,
		}))
	}

	traceExp, err := otlptracehttp.New(ctx, traceOpts...)
	if err != nil {
		return nil, fmt.Errorf("otel trace exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(traceExp,
			trace.WithBatchTimeout(5*time.Second),
			trace.WithMaxExportBatchSize(512),
		),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	// Metric exporter — OTLP HTTP for push-based metrics (CLI/TUI).
	metricOpts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpointURL(endpoint),
	}
	if token := authToken(); token != "" {
		metricOpts = append(metricOpts, otlpmetrichttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + token,
		}))
	}

	metricExp, err := otlpmetrichttp.New(ctx, metricOpts...)
	if err != nil {
		return nil, fmt.Errorf("otel metric exporter: %w", err)
	}

	// Prometheus exporter for /metrics endpoint (web mode scraping).
	promExp, err := promexporter.New()
	if err != nil {
		return nil, fmt.Errorf("prometheus exporter: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExp, metric.WithInterval(30*time.Second))),
		metric.WithReader(promExp),
	)
	otel.SetMeterProvider(mp)

	// W3C Trace Context propagation.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()

		debugf("flushing spans and metrics...")
		tpErr := tp.Shutdown(ctx)
		mpErr := mp.Shutdown(ctx)
		if tpErr != nil {
			debugf("tracer flush error: %v", tpErr)
			return fmt.Errorf("tracer shutdown: %w", tpErr)
		}
		if mpErr != nil {
			debugf("meter flush error: %v", mpErr)
			return fmt.Errorf("meter shutdown: %w", mpErr)
		}
		debugf("flush ok")
		return nil
	}, nil
}

func authToken() string {
	if t := os.Getenv("OTEL_AUTH_TOKEN"); t != "" {
		return t
	}
	return otelAuthToken
}
