package telemetry

import (
	ometric "go.opentelemetry.io/otel/metric"
)

// Metrics holds pre-created OTel instruments for recording usage data.
// All instruments are initialised from a real or noop meter — never nil.
type Metrics struct {
	// Invocations counts tool executions. Attributes: tool, mode, os, arch, version.
	Invocations ometric.Int64Counter
	// InputBytes records input file sizes in bytes. Attributes: tool.
	InputBytes ometric.Int64Histogram
	// Duration records processing time in seconds. Attributes: tool, mode.
	Duration ometric.Float64Histogram
	// Errors counts failed executions. Attributes: tool, error_type.
	Errors ometric.Int64Counter
	// JSONNodes records JSON node count per parse.
	JSONNodes ometric.Int64Histogram
	// DiffChanges counts added/removed/modified lines. Attributes: kind.
	DiffChanges ometric.Int64Counter
	// WebUploads counts file uploads to the hosted web UI. Attributes: tool.
	WebUploads ometric.Int64Counter
}

func newMetrics(meter ometric.Meter) Metrics {
	m := Metrics{}
	var err error

	m.Invocations, err = meter.Int64Counter("unum.invocations",
		ometric.WithDescription("Number of tool executions"))
	if err != nil {
		debugf("metrics: invocations counter: %v", err)
	}

	m.InputBytes, err = meter.Int64Histogram("unum.input.bytes",
		ometric.WithDescription("Input file size in bytes"),
		ometric.WithUnit("By"))
	if err != nil {
		debugf("metrics: input.bytes histogram: %v", err)
	}

	m.Duration, err = meter.Float64Histogram("unum.duration.seconds",
		ometric.WithDescription("Tool processing time"),
		ometric.WithUnit("s"))
	if err != nil {
		debugf("metrics: duration.seconds histogram: %v", err)
	}

	m.Errors, err = meter.Int64Counter("unum.errors",
		ometric.WithDescription("Number of failed executions"))
	if err != nil {
		debugf("metrics: errors counter: %v", err)
	}

	m.JSONNodes, err = meter.Int64Histogram("unum.json.nodes",
		ometric.WithDescription("Number of nodes in parsed JSON document"),
		ometric.WithExplicitBucketBoundaries(10, 100, 1_000, 10_000, 100_000, 1_000_000))
	if err != nil {
		debugf("metrics: json.nodes histogram: %v", err)
	}

	m.DiffChanges, err = meter.Int64Counter("unum.diff.changes",
		ometric.WithDescription("Total lines added, removed, or modified across all diffs"))
	if err != nil {
		debugf("metrics: diff.changes counter: %v", err)
	}

	m.WebUploads, err = meter.Int64Counter("unum.web.uploads",
		ometric.WithDescription("Number of file uploads to the hosted web UI"))
	if err != nil {
		debugf("metrics: web.uploads counter: %v", err)
	}

	return m
}
