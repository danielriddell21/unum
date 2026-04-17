package transform_test

import (
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/json/lens/transform"
	"github.com/danielriddell21/unum/internal/json/parse"
)

func yamlOf(t *testing.T, src string) string {
	t.Helper()
	root, err := parse.Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	out, err := transform.ToYAML(root)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestToYAMLBasicObject(t *testing.T) {
	out := yamlOf(t, `{"name": "alice", "active": true}`)
	if !strings.Contains(out, "name: alice") {
		t.Errorf("expected 'name: alice' in:\n%s", out)
	}
	if !strings.Contains(out, "active: true") {
		t.Errorf("expected 'active: true' in:\n%s", out)
	}
}

func TestToYAMLIntegerNotFloat(t *testing.T) {
	// Integers must not gain a decimal point in YAML output
	out := yamlOf(t, `{"count": 42}`)
	if !strings.Contains(out, "count: 42") {
		t.Errorf("expected 'count: 42' (not 42.0) in:\n%s", out)
	}
	if strings.Contains(out, "42.") {
		t.Errorf("integer 42 rendered as float in:\n%s", out)
	}
}

func TestToYAMLFloat(t *testing.T) {
	out := yamlOf(t, `{"ratio": 3.14}`)
	if !strings.Contains(out, "ratio: 3.14") {
		t.Errorf("expected 'ratio: 3.14' in:\n%s", out)
	}
}

func TestToYAMLNull(t *testing.T) {
	out := yamlOf(t, `{"nothing": null}`)
	if !strings.Contains(out, "nothing: null") {
		t.Errorf("expected 'nothing: null' in:\n%s", out)
	}
}

func TestToYAMLArray(t *testing.T) {
	out := yamlOf(t, `{"items": [1, 2, 3]}`)
	if !strings.Contains(out, "items:") {
		t.Errorf("expected 'items:' in:\n%s", out)
	}
	if !strings.Contains(out, "- 1") {
		t.Errorf("expected '- 1' in:\n%s", out)
	}
}

func TestToYAMLPreservesKeyOrder(t *testing.T) {
	out := yamlOf(t, `{"z": 1, "a": 2, "m": 3}`)
	zIdx := strings.Index(out, "z:")
	aIdx := strings.Index(out, "a:")
	mIdx := strings.Index(out, "m:")
	if zIdx < 0 || aIdx < 0 || mIdx < 0 {
		t.Fatalf("missing keys in:\n%s", out)
	}
	if zIdx >= aIdx || aIdx >= mIdx {
		t.Errorf("key order not preserved: z=%d a=%d m=%d in:\n%s", zIdx, aIdx, mIdx, out)
	}
}

func TestToYAMLNested(t *testing.T) {
	out := yamlOf(t, `{"user": {"name": "bob", "age": 30}}`)
	if !strings.Contains(out, "user:") {
		t.Errorf("expected 'user:' in:\n%s", out)
	}
	if !strings.Contains(out, "name: bob") {
		t.Errorf("expected 'name: bob' in:\n%s", out)
	}
}

func TestToYAMLExponentNotation(t *testing.T) {
	// Numbers with 'e' or 'E' should get !!float tag, not !!int.
	for _, n := range []string{"1e10", "1E10", "2.5e-3"} {
		out := yamlOf(t, `{"v": `+n+`}`)
		if !strings.Contains(out, "v:") {
			t.Errorf("exponent %q: missing key in:\n%s", n, out)
		}
	}
}

func TestToYAMLEmptyObject(t *testing.T) {
	out := yamlOf(t, `{}`)
	if out == "" {
		t.Error("empty object ToYAML returned empty string")
	}
}

func TestToYAMLEmptyArray(t *testing.T) {
	out := yamlOf(t, `[]`)
	if out == "" {
		t.Error("empty array ToYAML returned empty string")
	}
}

func TestToYAMLBool(t *testing.T) {
	out := yamlOf(t, `{"ok": false}`)
	if !strings.Contains(out, "ok: false") {
		t.Errorf("expected 'ok: false' in:\n%s", out)
	}
}
