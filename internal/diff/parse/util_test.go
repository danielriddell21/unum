package parse

import (
	"fmt"
	"strings"
	"testing"
)

func TestObjPath(t *testing.T) {
	tests := []struct {
		parent, key, want string
	}{
		{".", "version", ".version"},          // root: no double-dot
		{".", "name", ".name"},                // root: no double-dot
		{".parent", "child", ".parent.child"}, // nested
		{".a.b", "c", ".a.b.c"},               // deeply nested
	}
	for _, tt := range tests {
		got := objPath(tt.parent, tt.key)
		if got != tt.want {
			t.Errorf("objPath(%q, %q) = %q, want %q", tt.parent, tt.key, got, tt.want)
		}
	}
}

func formatOps(ops []alignOp) string {
	parts := make([]string, 0, len(ops))
	for _, op := range ops {
		parts = append(parts, fmt.Sprintf("(%d,%d)", op.a, op.b))
	}
	return strings.Join(parts, " ")
}

func TestAlignByKey(t *testing.T) {
	tests := []struct {
		name       string
		aKeys      []string
		bKeys      []string
		want       string
		wantIgnore bool
	}{
		{
			name:  "identical",
			aKeys: []string{"a", "b"},
			bKeys: []string{"a", "b"},
			want:  "(0,0) (1,1)",
		},
		{
			name:  "insert at front keeps the rest aligned",
			aKeys: []string{"b", "c"},
			bKeys: []string{"a", "b", "c"},
			want:  "(-1,0) (0,1) (1,2)",
		},
		{
			name:  "insert in the middle keeps the rest aligned",
			aKeys: []string{"a", "c"},
			bKeys: []string{"a", "b", "c"},
			want:  "(0,0) (-1,1) (1,2)",
		},
		{
			name:  "removal from the middle",
			aKeys: []string{"a", "b", "c"},
			bKeys: []string{"a", "c"},
			want:  "(0,0) (1,-1) (2,1)",
		},
		{
			name:  "in-place replacement pairs as modified",
			aKeys: []string{"a", "b", "c"},
			bKeys: []string{"a", "x", "c"},
			want:  "(0,0) (1,1) (2,2)",
		},
		{
			name:  "reorder matches every element",
			aKeys: []string{"c", "a", "b"},
			bKeys: []string{"a", "b", "c"},
			want:  "(1,0) (2,1) (0,2)",
		},
		{
			name:  "empty before",
			aKeys: nil,
			bKeys: []string{"a"},
			want:  "(-1,0)",
		},
		{
			name:  "empty after",
			aKeys: []string{"a"},
			bKeys: nil,
			want:  "(0,-1)",
		},
		{
			name:  "both empty",
			aKeys: nil,
			bKeys: nil,
			want:  "",
		},
		{
			name:  "uneven replacement run",
			aKeys: []string{"a", "b", "c"},
			bKeys: []string{"x"},
			want:  "(0,0) (1,-1) (2,-1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatOps(alignByKey(tt.aKeys, tt.bKeys)); got != tt.want {
				t.Errorf("alignByKey(%v, %v) = %s, want %s", tt.aKeys, tt.bKeys, got, tt.want)
			}
		})
	}
}

func TestAlignByKeyFallsBackWhenHuge(t *testing.T) {
	aKeys := make([]string, 600)
	bKeys := make([]string, 600)
	for i := range aKeys {
		aKeys[i] = fmt.Sprintf("k%d", i)
		bKeys[i] = fmt.Sprintf("k%d", i)
	}
	ops := alignByKey(aKeys, append(bKeys, "new"))
	if len(ops) != 601 {
		t.Fatalf("ops = %d, want 601", len(ops))
	}
	for i := range 600 {
		if ops[i] != (alignOp{a: i, b: i}) {
			t.Fatalf("ops[%d] = %v, want positional pairing", i, ops[i])
		}
	}
	if last := ops[600]; last != (alignOp{a: -1, b: 600}) {
		t.Errorf("trailing op = %v, want an addition", last)
	}
}
