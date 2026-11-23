package query_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/danielriddell21/unum/internal/json/lens/query"
	"github.com/danielriddell21/unum/internal/json/parse"
)

func TestQueryFieldAccess(t *testing.T) {
	root, err := parse.Parse([]byte(`{"name": "alice", "age": 30}`))
	if err != nil {
		t.Fatal(err)
	}

	result, err := query.Execute(context.Background(), root, ".name")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(result) != `"alice"` {
		t.Errorf("got %q, want %q", result, `"alice"`)
	}
}

func TestQueryArrayLength(t *testing.T) {
	root, err := parse.Parse([]byte(`{"items": [1, 2, 3]}`))
	if err != nil {
		t.Fatal(err)
	}

	result, err := query.Execute(context.Background(), root, ".items | length")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(result) != "3" {
		t.Errorf("got %q, want \"3\"", result)
	}
}

func TestQueryNestedField(t *testing.T) {
	root, err := parse.Parse([]byte(`{"user": {"email": "alice@example.com"}}`))
	if err != nil {
		t.Fatal(err)
	}

	result, err := query.Execute(context.Background(), root, ".user.email")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(result) != `"alice@example.com"` {
		t.Errorf("got %q", result)
	}
}

func TestQueryInvalidExpression(t *testing.T) {
	root, err := parse.Parse([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}

	_, err = query.Execute(context.Background(), root, "!!invalid!!")
	if err == nil {
		t.Error("expected error for invalid jq expression")
	}
}

func TestQueryEmptyExpression(t *testing.T) {
	root, err := parse.Parse([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}

	_, err = query.Execute(context.Background(), root, "")
	if err == nil {
		t.Error("expected error for empty expression")
	}
}

func TestQueryIdentity(t *testing.T) {
	root, err := parse.Parse([]byte(`[1, 2, 3]`))
	if err != nil {
		t.Fatal(err)
	}

	result, err := query.Execute(context.Background(), root, ".")
	if err != nil {
		t.Fatal(err)
	}
	if result == "" {
		t.Error("identity query returned empty result")
	}
}

func TestQueryMultipleOutputs(t *testing.T) {
	// ".[]" on an array emits one result per element, joined by newlines.
	root, err := parse.Parse([]byte(`[1, 2, 3]`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := query.Execute(context.Background(), root, ".[]")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %q", len(lines), result)
	}
}

func TestQueryAllNodeKinds(t *testing.T) {
	tests := []struct {
		name  string
		input string
		expr  string
	}{
		{"null", `{"x": null}`, ".x"},
		{"bool", `{"x": true}`, ".x"},
		{"number", `{"x": 42}`, ".x"},
		{"string", `{"x": "hi"}`, ".x"},
		{"array", `{"x": [1]}`, ".x"},
		{"object", `{"x": {"k": 1}}`, ".x"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, err2 := parse.Parse([]byte(tc.input))
			if err2 != nil {
				t.Fatal(err2)
			}
			_, err2 = query.Execute(context.Background(), root, tc.expr)
			if err2 != nil {
				t.Errorf("Execute(%q): %v", tc.expr, err2)
			}
		})
	}
}

func TestQueryNoResult(t *testing.T) {
	root, err := parse.Parse([]byte(`{"a": 1}`))
	if err != nil {
		t.Fatal(err)
	}
	// empty expression produces no output — should return empty string, no error.
	result, err := query.Execute(context.Background(), root, "empty")
	if err != nil {
		t.Fatalf("empty query: %v", err)
	}
	if result != "" {
		t.Errorf("empty query result=%q, want empty string", result)
	}
}

func TestQueryTimeout(t *testing.T) {
	root, err := parse.Parse([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}

	for _, expr := range []string{"[repeat(1)]", "repeat(1)", "last(range(1e12))"} {
		t.Run(expr, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			start := time.Now()
			_, err := query.Execute(ctx, root, expr)
			if !errors.Is(err, query.ErrTimeout) {
				t.Fatalf("err=%v, want ErrTimeout", err)
			}
			if elapsed := time.Since(start); elapsed > 5*time.Second {
				t.Errorf("query ran %s past a 100ms deadline", elapsed)
			}
		})
	}
}

func TestQueryCancelled(t *testing.T) {
	root, err := parse.Parse([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = query.Execute(ctx, root, "[repeat(1)]")
	if err == nil {
		t.Fatal("expected an error from a cancelled context")
	}
	if errors.Is(err, query.ErrTimeout) {
		t.Error("cancellation must not be reported as a timeout")
	}
}

func TestQueryOutputTruncated(t *testing.T) {
	root, err := parse.Parse([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}

	// Streams endless results; only the byte cap can stop it.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := query.Execute(ctx, root, `repeat("0123456789")`)
	if !errors.Is(err, query.ErrOutputTruncated) {
		t.Fatalf("err=%v, want ErrOutputTruncated", err)
	}
	if !strings.Contains(err.Error(), strconv.Itoa(query.MaxOutputBytes)) {
		t.Errorf("error %q should state the cap", err)
	}
	if result != "" {
		t.Errorf("a truncated query must not return partial output, got %d bytes", len(result))
	}
}

func TestQueryOutputAtCapBoundary(t *testing.T) {
	root, err := parse.Parse([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}

	// A string of n bytes marshals to n+2 bytes including the quotes.
	exactly := fmt.Sprintf(`"x" * %d`, query.MaxOutputBytes-2)
	result, err := query.Execute(context.Background(), root, exactly)
	if err != nil {
		t.Fatalf("output of exactly the cap should pass: %v", err)
	}
	if len(result) != query.MaxOutputBytes {
		t.Fatalf("len=%d, want %d", len(result), query.MaxOutputBytes)
	}

	over := fmt.Sprintf(`"x" * %d`, query.MaxOutputBytes-1)
	if _, err := query.Execute(context.Background(), root, over); !errors.Is(err, query.ErrOutputTruncated) {
		t.Fatalf("err=%v, want ErrOutputTruncated one byte over the cap", err)
	}
}

func TestQueryExecuteAnyReusesInput(t *testing.T) {
	root, err := parse.Parse([]byte(`{"a": [1, 2]}`))
	if err != nil {
		t.Fatal(err)
	}
	input := root.ToAny()

	for range 3 {
		result, err := query.ExecuteAny(context.Background(), input, ".a | length")
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(result) != "2" {
			t.Errorf("got %q, want 2", result)
		}
	}
}
