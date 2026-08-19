package parse

import (
	"fmt"
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

func TestAlignByKey(t *testing.T) {
	tests := []struct {
		name  string
		aKeys []string
		bKeys []string
		want  string
	}{
		{
			name:  "identical",
			aKeys: []string{"a", "b"},
			bKeys: []string{"a", "b"},
			want:  "[{0 0} {1 1}]",
		},
		{
			name:  "insert at front keeps the rest aligned",
			aKeys: []string{"b", "c"},
			bKeys: []string{"a", "b", "c"},
			want:  "[{-1 0} {0 1} {1 2}]",
		},
		{
			name:  "insert in the middle keeps the rest aligned",
			aKeys: []string{"a", "c"},
			bKeys: []string{"a", "b", "c"},
			want:  "[{0 0} {-1 1} {1 2}]",
		},
		{
			name:  "removal from the middle",
			aKeys: []string{"a", "b", "c"},
			bKeys: []string{"a", "c"},
			want:  "[{0 0} {1 -1} {2 1}]",
		},
		{
			name:  "in-place replacement pairs as modified",
			aKeys: []string{"a", "b", "c"},
			bKeys: []string{"a", "x", "c"},
			want:  "[{0 0} {1 1} {2 2}]",
		},
		{
			name:  "reorder matches every element",
			aKeys: []string{"c", "a", "b"},
			bKeys: []string{"a", "b", "c"},
			want:  "[{1 0} {2 1} {0 2}]",
		},
		{
			name:  "empty before",
			aKeys: nil,
			bKeys: []string{"a"},
			want:  "[{-1 0}]",
		},
		{
			name:  "empty after",
			aKeys: []string{"a"},
			bKeys: nil,
			want:  "[{0 -1}]",
		},
		{
			name:  "both empty",
			aKeys: nil,
			bKeys: nil,
			want:  "[]",
		},
		{
			name:  "uneven replacement run",
			aKeys: []string{"a", "b", "c"},
			bKeys: []string{"x"},
			want:  "[{0 0} {1 -1} {2 -1}]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmt.Sprint(alignByKey(tt.aKeys, tt.bKeys)); got != tt.want {
				t.Errorf("alignByKey(%v, %v) = %s, want %s", tt.aKeys, tt.bKeys, got, tt.want)
			}
		})
	}
}
