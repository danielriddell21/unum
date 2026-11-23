package jsontool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/danielriddell21/unum/internal/json/lens/query"
	"github.com/danielriddell21/unum/internal/json/parse"
	"github.com/danielriddell21/unum/internal/json/render/static"
	"github.com/danielriddell21/unum/internal/telemetry"
)

func lensInvocation(t *testing.T, expr string) (*invocation, context.Context) {
	t.Helper()
	root, err := parse.Parse([]byte(`{"a": 1}`))
	if err != nil {
		t.Fatal(err)
	}
	tel := telemetry.New()
	ctx, span := tel.Tracer().Start(context.Background(), "test")
	t.Cleanup(func() { span.End() })

	return &invocation{
		f:          &flags{jqExpr: expr, noColor: true, tel: tel},
		root:       root,
		renderOpts: static.Options{NoColor: true, Quiet: true},
		span:       span,
		start:      time.Now(),
		mode:       "cli",
	}, ctx
}

func TestDispatchLensMode_QueryFailureReturnsExitError(t *testing.T) {
	inv, ctx := lensInvocation(t, "!!invalid!!")

	handled, err := dispatchLensMode(ctx, inv)
	if handled {
		t.Error("a failed query should not report the lens as handled")
	}
	var ee *exitError
	if !errors.As(err, &ee) {
		t.Fatalf("err=%v, want an *exitError", err)
	}
	if ee.ExitCode() != 2 {
		t.Errorf("ExitCode=%d, want 2", ee.ExitCode())
	}
}

func TestDispatchLensMode_QueryTimeoutIsDistinct(t *testing.T) {
	orig := queryTimeout
	queryTimeout = 100 * time.Millisecond
	t.Cleanup(func() { queryTimeout = orig })

	inv, ctx := lensInvocation(t, "[repeat(1)]")

	_, err := dispatchLensMode(ctx, inv)
	if !errors.Is(err, query.ErrTimeout) {
		t.Fatalf("err=%v, want query.ErrTimeout", err)
	}
	var ee *exitError
	if !errors.As(err, &ee) || ee.ExitCode() != 2 {
		t.Errorf("timeout should exit 2, got %v", err)
	}
}

func TestExitError(t *testing.T) {
	base := errors.New("boom")
	err := &exitError{code: 7, err: base}
	if err.Error() != "boom" {
		t.Errorf("Error()=%q", err.Error())
	}
	if !errors.Is(err, base) {
		t.Error("exitError should unwrap to its cause")
	}
	if err.ExitCode() != 7 {
		t.Errorf("ExitCode()=%d", err.ExitCode())
	}
}
