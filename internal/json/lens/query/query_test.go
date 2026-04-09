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
