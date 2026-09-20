// letsgo-env.mod
//
// Replaces the two -X assignments the GoReleaser ldflags read from the
// environment. Neither is a secret: a value passed to -X is compiled into the
// binary and recoverable with `strings`, and letsgo records both in
// letsgo.json so the build can be reproduced.
inject internal/telemetry.otelEndpoint  OTEL_ENDPOINT
inject internal/telemetry.otelAuthToken OTEL_AUTH_TOKEN
