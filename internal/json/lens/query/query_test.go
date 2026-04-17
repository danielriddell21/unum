package query_test

import (
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/json/lens/query"
	"github.com/danielriddell21/unum/internal/json/parse"
)

func TestQueryFieldAccess(t *testing.T) {
	root, err := parse.Parse([]byte(`{"name": "alice", "age": 30}`))
	if err != nil {
		t.Fatal(err)
	}

	result, err := query.Execute(root, ".name")
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

	result, err := query.Execute(root, ".items | length")
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

	result, err := query.Execute(root, ".user.email")
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

	_, err = query.Execute(root, "!!invalid!!")
	if err == nil {
		t.Error("expected error for invalid jq expression")
	}
}

func TestQueryEmptyExpression(t *testing.T) {
	root, err := parse.Parse([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}

	_, err = query.Execute(root, "")
	if err == nil {
		t.Error("expected error for empty expression")
	}
}

func TestQueryIdentity(t *testing.T) {
	root, err := parse.Parse([]byte(`[1, 2, 3]`))
	if err != nil {
		t.Fatal(err)
	}

	result, err := query.Execute(root, ".")
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
	result, err := query.Execute(root, ".[]")
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
			_, err2 = query.Execute(root, tc.expr)
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
	result, err := query.Execute(root, "empty")
	if err != nil {
		t.Fatalf("empty query: %v", err)
	}
	if result != "" {
		t.Errorf("empty query result=%q, want empty string", result)
	}
}
