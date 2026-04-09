package parse_test

import (
	"testing"

	"github.com/danielriddell21/unum/internal/json/node"
	"github.com/danielriddell21/unum/internal/json/parse"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid object", `{"a": 1}`, false},
		{"valid array", `[1, 2, 3]`, false},
		{"valid string scalar", `"hello"`, false},
		{"valid nested", `{"items": [1, {"x": true}]}`, false},
		{"invalid trailing comma", `{"a": 1,}`, true},
		{"invalid syntax", `{broken`, true},
		{"empty input", ``, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parse.Validate([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseObject(t *testing.T) {
	root, err := parse.Parse([]byte(`{"name": "alice", "age": 30, "active": true}`))
	if err != nil {
		t.Fatal(err)
	}
	if root.Kind != node.KindObject {
		t.Fatalf("expected KindObject, got %v", root.Kind)
	}
	if len(root.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(root.Children))
	}

	name := root.Children[0]
	if name.Key != "name" || name.Kind != node.KindString {
		t.Errorf("first child: key=%q kind=%v", name.Key, name.Kind)
	}

	age := root.Children[1]
	if age.Key != "age" || age.Kind != node.KindNumber || age.Raw != "30" {
		t.Errorf("second child: key=%q kind=%v raw=%q", age.Key, age.Kind, age.Raw)
	}

	active := root.Children[2]
	if active.Kind != node.KindBool {
		t.Errorf("third child: expected KindBool, got %v", active.Kind)
	}
}

func TestParseArray(t *testing.T) {
	root, err := parse.Parse([]byte(`[1, "two", null, false]`))
	if err != nil {
		t.Fatal(err)
	}
	if root.Kind != node.KindArray {
		t.Fatalf("expected KindArray, got %v", root.Kind)
	}
	if len(root.Children) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(root.Children))
	}
	if root.Children[0].Kind != node.KindNumber {
		t.Errorf("index 0: expected KindNumber")
	}
	if root.Children[2].Kind != node.KindNull {
		t.Errorf("index 2: expected KindNull")
	}
	if root.Children[3].Kind != node.KindBool {
		t.Errorf("index 3: expected KindBool")
	}
}

func TestParsePreservesKeyOrder(t *testing.T) {
	root, err := parse.Parse([]byte(`{"z": 1, "a": 2, "m": 3}`))
	if err != nil {
		t.Fatal(err)
	}
	keys := []string{"z", "a", "m"}
	for i, want := range keys {
		if got := root.Children[i].Key; got != want {
			t.Errorf("key[%d]: got %q, want %q", i, got, want)
		}
	}
}

func TestParseNested(t *testing.T) {
	root, err := parse.Parse([]byte(`{"items": [1, 2, 3]}`))
	if err != nil {
		t.Fatal(err)
	}
	items := root.Children[0]
	if items.Kind != node.KindArray {
		t.Fatalf("expected KindArray, got %v", items.Kind)
	}
	if len(items.Children) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items.Children))
	}
}
