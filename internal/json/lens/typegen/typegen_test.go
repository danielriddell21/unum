package typegen_test

import (
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/json/lens/typegen"
	"github.com/danielriddell21/unum/internal/json/parse"
)

func generate(t *testing.T, src string, opts typegen.Options) string {
	t.Helper()
	root, err := parse.Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	out, err := typegen.Generate(root, opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	return out
}

func TestGenerateGoPackageDeclaration(t *testing.T) {
	out := generate(t, `{"name": "alice"}`, typegen.Options{
		Target:      typegen.TargetGo,
		PackageName: "myapp",
	})
	if !strings.HasPrefix(out, "package myapp") {
		t.Errorf("expected 'package myapp' prefix, got:\n%s", out)
	}
}

func TestGenerateGoDefaultPackage(t *testing.T) {
	out := generate(t, `{"x": 1}`, typegen.Options{Target: typegen.TargetGo})
	if !strings.Contains(out, "package main") {
		t.Errorf("expected 'package main', got:\n%s", out)
	}
}

func TestGenerateGoStructFields(t *testing.T) {
	out := generate(t, `{"name": "alice", "age": 30, "active": true}`, typegen.Options{
		Target:   typegen.TargetGo,
		TypeName: "User",
	})
	if !strings.Contains(out, "type User struct") {
		t.Errorf("expected 'type User struct', got:\n%s", out)
	}
	if !strings.Contains(out, `json:"name"`) {
		t.Errorf("expected json tag for 'name', got:\n%s", out)
	}
	if !strings.Contains(out, "float64") {
		t.Errorf("expected 'float64' for number field, got:\n%s", out)
	}
	if !strings.Contains(out, "bool") {
		t.Errorf("expected 'bool' for boolean field, got:\n%s", out)
	}
}

func TestGenerateGoNestedStructNaming(t *testing.T) {
	// Nested struct name must be ParentFieldName, not just FieldName
	out := generate(t, `{"author": {"name": "bob"}}`, typegen.Options{
		Target:   typegen.TargetGo,
		TypeName: "Root",
	})
	if !strings.Contains(out, "type RootAuthor struct") {
		t.Errorf("expected 'type RootAuthor struct', got:\n%s", out)
	}
	if !strings.Contains(out, "Author RootAuthor") {
		t.Errorf("expected 'Author RootAuthor' field, got:\n%s", out)
	}
}

func TestGenerateGoNullableField(t *testing.T) {
	out := generate(t, `{"val": null}`, typegen.Options{
		Target:   typegen.TargetGo,
		TypeName: "Root",
	})
	// null fields become `any`
	if !strings.Contains(out, "any") {
		t.Errorf("expected 'any' for null field, got:\n%s", out)
	}
}

func TestGenerateTSInterface(t *testing.T) {
	out := generate(t, `{"name": "alice", "age": 30}`, typegen.Options{
		Target:   typegen.TargetTypeScript,
		TypeName: "User",
	})
	if !strings.Contains(out, "interface User") {
		t.Errorf("expected 'interface User', got:\n%s", out)
	}
	if !strings.Contains(out, "name: string") {
		t.Errorf("expected 'name: string', got:\n%s", out)
	}
	if !strings.Contains(out, "age: number") {
		t.Errorf("expected 'age: number', got:\n%s", out)
	}
}

func TestGenerateTSOptionalField(t *testing.T) {
	out := generate(t, `{"val": null}`, typegen.Options{
		Target:   typegen.TargetTypeScript,
		TypeName: "Root",
	})
	// null fields should be optional (?) or typed as null
	if !strings.Contains(out, "null") {
		t.Errorf("expected 'null' type for null field, got:\n%s", out)
	}
}

func TestGenerateTSNestedInterface(t *testing.T) {
	out := generate(t, `{"meta": {"count": 1}}`, typegen.Options{
		Target:   typegen.TargetTypeScript,
		TypeName: "Root",
	})
	if !strings.Contains(out, "interface Root") {
		t.Errorf("expected 'interface Root', got:\n%s", out)
	}
	if !strings.Contains(out, "interface RootMeta") {
		t.Errorf("expected 'interface RootMeta', got:\n%s", out)
	}
}

func TestGenerateJSONSchemaHeader(t *testing.T) {
	out := generate(t, `{"name": "alice"}`, typegen.Options{
		Target:   typegen.TargetJSONSchema,
		TypeName: "User",
	})
	if !strings.Contains(out, `"$schema"`) {
		t.Errorf("expected '$schema', got:\n%s", out)
	}
	if !strings.Contains(out, `"title": "User"`) {
		t.Errorf("expected '\"title\": \"User\"', got:\n%s", out)
	}
	if !strings.Contains(out, `"type": "object"`) {
		t.Errorf("expected '\"type\": \"object\"', got:\n%s", out)
	}
}

func TestGenerateJSONSchemaRequired(t *testing.T) {
	// Non-nullable fields should appear in required
	out := generate(t, `{"name": "alice", "age": 30}`, typegen.Options{
		Target:   typegen.TargetJSONSchema,
		TypeName: "Root",
	})
	if !strings.Contains(out, `"required"`) {
		t.Errorf("expected 'required' array, got:\n%s", out)
	}
}

func TestGenerateJSONSchemaTypes(t *testing.T) {
	out := generate(t, `{"flag": true, "count": 1, "label": "x"}`, typegen.Options{
		Target:   typegen.TargetJSONSchema,
		TypeName: "Root",
	})
	if !strings.Contains(out, `"boolean"`) {
		t.Errorf("expected '\"boolean\"', got:\n%s", out)
	}
	if !strings.Contains(out, `"number"`) {
		t.Errorf("expected '\"number\"', got:\n%s", out)
	}
	if !strings.Contains(out, `"string"`) {
		t.Errorf("expected '\"string\"', got:\n%s", out)
	}
}

func TestGenerateUnknownTarget(t *testing.T) {
	root, err := parse.Parse([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	_, err = typegen.Generate(root, typegen.Options{Target: "invalid"})
	if err == nil {
		t.Error("expected error for unknown target")
	}
}

func TestGenerateGo_NullableField(t *testing.T) {
	// A nested array with sometimes-null field → typeinfo marks it nullable → *float64.
	out := generate(t, `{"users": [{"age": 30}, {"age": null}]}`, typegen.Options{
		Target:   typegen.TargetGo,
		TypeName: "Root",
	})
	if !strings.Contains(out, "*float64") {
		t.Errorf("nullable number field should be '*float64', got:\n%s", out)
	}
}

func TestGenerateTS_NullableField(t *testing.T) {
	out := generate(t, `{"users": [{"age": 30}, {"age": null}]}`, typegen.Options{
		Target:   typegen.TargetTypeScript,
		TypeName: "Root",
	})
	// Nullable field should have "?" marker and/or "| null" suffix.
	if !strings.Contains(out, "null") {
		t.Errorf("nullable field should mention 'null' in TS, got:\n%s", out)
	}
}

func TestGenerateJSONSchema_NullableField(t *testing.T) {
	out := generate(t, `{"users": [{"age": 30}, {"age": null}]}`, typegen.Options{
		Target:   typegen.TargetJSONSchema,
		TypeName: "Root",
	})
	// Nullable fields get anyOf in JSON Schema 2020-12.
	if !strings.Contains(out, "anyOf") {
		t.Errorf("nullable field should use anyOf in JSONSchema, got:\n%s", out)
	}
}

func TestGenerateGo_AlwaysNull(t *testing.T) {
	out := generate(t, `{"x": null}`, typegen.Options{
		Target:   typegen.TargetGo,
		TypeName: "Root",
	})
	// TypeNull maps to "any" in Go.
	if !strings.Contains(out, "any") {
		t.Errorf("null field should be 'any' in Go, got:\n%s", out)
	}
}

func TestGenerateTS_AlwaysNull(t *testing.T) {
	out := generate(t, `{"x": null}`, typegen.Options{
		Target:   typegen.TargetTypeScript,
		TypeName: "Root",
	})
	if out == "" {
		t.Error("TS generation returned empty for always-null field")
	}
}

func TestGenerateGo_ArrayField(t *testing.T) {
	out := generate(t, `{"nums": [1, 2, 3]}`, typegen.Options{
		Target:   typegen.TargetGo,
		TypeName: "Root",
	})
	if !strings.Contains(out, "[]float64") {
		t.Errorf("expected '[]float64' for number array, got:\n%s", out)
	}
}

func TestGenerateTS_ArrayField(t *testing.T) {
	out := generate(t, `{"nums": [1, 2, 3]}`, typegen.Options{
		Target:   typegen.TargetTypeScript,
		TypeName: "Root",
	})
	if !strings.Contains(out, "number[]") {
		t.Errorf("expected 'number[]' for number array, got:\n%s", out)
	}
}

func TestGenerateJSONSchema_ArrayField(t *testing.T) {
	out := generate(t, `{"nums": [1, 2]}`, typegen.Options{
		Target:   typegen.TargetJSONSchema,
		TypeName: "Root",
	})
	if !strings.Contains(out, `"array"`) {
		t.Errorf("expected '\"array\"' in JSON Schema, got:\n%s", out)
	}
	if !strings.Contains(out, `"items"`) {
		t.Errorf("expected '\"items\"' in JSON Schema for typed array, got:\n%s", out)
	}
}

func TestGenerateGo_MixedTypeField(t *testing.T) {
	// A field with both number and string values → TypeMixed → "any"
	out := generate(t, `{"items": [{"x": 1}, {"x": "str"}]}`, typegen.Options{
		Target:   typegen.TargetGo,
		TypeName: "Root",
	})
	if !strings.Contains(out, "any") {
		t.Errorf("mixed-type field should be 'any' in Go, got:\n%s", out)
	}
}

func TestGenerateTS_MixedTypeField(t *testing.T) {
	out := generate(t, `{"items": [{"x": 1}, {"x": "str"}]}`, typegen.Options{
		Target:   typegen.TargetTypeScript,
		TypeName: "Root",
	})
	if !strings.Contains(out, "unknown") {
		t.Errorf("mixed-type field should be 'unknown' in TS, got:\n%s", out)
	}
}
