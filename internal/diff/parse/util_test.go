package parse

import "testing"

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
