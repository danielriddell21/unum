package typeinfo_test

import (
	"testing"

	"github.com/danielriddell21/unum/internal/json/parse"
	"github.com/danielriddell21/unum/internal/json/typeinfo"
)

func infer(t *testing.T, src string) *typeinfo.TypeInfo {
	t.Helper()
	root, err := parse.Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	return typeinfo.Infer(root)
}

func TestInferScalars(t *testing.T) {
	cases := []struct {
		src  string
		kind typeinfo.TypeKind
	}{
		{`null`, typeinfo.TypeNull},
		{`true`, typeinfo.TypeBool},
		{`42`, typeinfo.TypeNumber},
		{`"hello"`, typeinfo.TypeString},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			ti := infer(t, tc.src)
			if ti.Kind != tc.kind {
				t.Errorf("got %v, want %v", ti.Kind, tc.kind)
			}
		})
	}
}

func TestInferObjectFields(t *testing.T) {
	ti := infer(t, `{"name": "alice", "age": 30}`)
	if ti.Kind != typeinfo.TypeObject {
		t.Fatalf("expected TypeObject, got %v", ti.Kind)
	}
	if len(ti.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(ti.Fields))
	}
	if ti.Fields[0].Name != "name" || ti.Fields[0].GoName != "Name" {
		t.Errorf("field[0]: name=%q goName=%q", ti.Fields[0].Name, ti.Fields[0].GoName)
	}
	if ti.Fields[0].Type.Kind != typeinfo.TypeString {
		t.Errorf("field[0] type: got %v, want TypeString", ti.Fields[0].Type.Kind)
	}
	if ti.Fields[1].Name != "age" || ti.Fields[1].GoName != "Age" {
		t.Errorf("field[1]: name=%q goName=%q", ti.Fields[1].Name, ti.Fields[1].GoName)
	}
}

func TestInferPascalCaseConversions(t *testing.T) {
	cases := []struct{ key, want string }{
		{"foo_bar", "FooBar"},
		{"camelCase", "CamelCase"},
		{"kebab-case", "KebabCase"},
		{"simple", "Simple"},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			src := `{"` + tc.key + `": 1}`
			ti := infer(t, src)
			if len(ti.Fields) == 0 {
				t.Fatal("no fields")
			}
			if ti.Fields[0].GoName != tc.want {
				t.Errorf("GoName: got %q, want %q", ti.Fields[0].GoName, tc.want)
			}
		})
	}
}

func TestInferNullField(t *testing.T) {
	ti := infer(t, `{"val": null}`)
	if !ti.Fields[0].Nullable {
		t.Error("field with null value should be nullable")
	}
}

func TestInferArrayOfNumbers(t *testing.T) {
	ti := infer(t, `[1, 2, 3]`)
	if ti.Kind != typeinfo.TypeArray {
		t.Fatalf("expected TypeArray, got %v", ti.Kind)
	}
	if ti.Elem.Kind != typeinfo.TypeNumber {
		t.Errorf("elem: got %v, want TypeNumber", ti.Elem.Kind)
	}
}

func TestInferArrayMergesObjectFields(t *testing.T) {
	// Second element has extra field "b" → "b" must be nullable (absent in first)
	// First element has "a" absent in second → "a" must be nullable
	ti := infer(t, `[{"a": 1}, {"a": 2, "b": "x"}]`)
	if ti.Kind != typeinfo.TypeArray {
		t.Fatalf("expected TypeArray, got %v", ti.Kind)
	}
	elem := ti.Elem
	if elem.Kind != typeinfo.TypeObject {
		t.Fatalf("elem: expected TypeObject, got %v", elem.Kind)
	}

	fields := make(map[string]*typeinfo.FieldInfo)
	for _, f := range elem.Fields {
		fields[f.Name] = f
	}

	if _, ok := fields["a"]; !ok {
		t.Error("expected field 'a'")
	}
	if _, ok := fields["b"]; !ok {
		t.Error("expected field 'b'")
	}
	if !fields["b"].Nullable {
		t.Error("field 'b' should be nullable (absent in first element)")
	}
}

func TestInferEmptyArray(t *testing.T) {
	ti := infer(t, `[]`)
	if ti.Kind != typeinfo.TypeArray {
		t.Fatalf("expected TypeArray, got %v", ti.Kind)
	}
	if ti.Elem == nil {
		t.Fatal("elem should not be nil")
	}
	if ti.Elem.Kind != typeinfo.TypeMixed {
		t.Errorf("empty array elem: got %v, want TypeMixed", ti.Elem.Kind)
	}
}

func TestInferNullMakesFieldNullable(t *testing.T) {
	// Array where one element is null → merged elem should be nullable
	root, err := parse.Parse([]byte(`[1, null, 3]`))
	if err != nil {
		t.Fatal(err)
	}
	ti := typeinfo.Infer(root)
	if !ti.Elem.Nullable {
		t.Error("array containing null should produce nullable elem")
	}
}
