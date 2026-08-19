package lcs

import (
	"fmt"
	"strings"
	"testing"
)

func formatOps(ops []Op) string {
	parts := make([]string, 0, len(ops))
	for _, op := range ops {
		parts = append(parts, fmt.Sprintf("(%d,%d)", op.A, op.B))
	}
	return strings.Join(parts, " ")
}

func TestAlign(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
		want string
	}{
		{
			name: "identical",
			a:    []string{"a", "b"},
			b:    []string{"a", "b"},
			want: "(0,0) (1,1)",
		},
		{
			name: "insert at front",
			a:    []string{"b", "c"},
			b:    []string{"a", "b", "c"},
			want: "(-1,0) (0,1) (1,2)",
		},
		{
			name: "insert in the middle",
			a:    []string{"a", "c"},
			b:    []string{"a", "b", "c"},
			want: "(0,0) (-1,1) (1,2)",
		},
		{
			name: "insert at the end",
			a:    []string{"a"},
			b:    []string{"a", "b"},
			want: "(0,0) (-1,1)",
		},
		{
			name: "removal from the middle",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "c"},
			want: "(0,0) (1,-1) (2,1)",
		},
		{
			name: "replacement removes before it adds",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "x", "c"},
			want: "(0,0) (1,-1) (-1,1) (2,2)",
		},
		{
			name: "reorder keeps the longest run",
			a:    []string{"c", "a", "b"},
			b:    []string{"a", "b", "c"},
			want: "(0,-1) (1,0) (2,1) (-1,2)",
		},
		{
			name: "nothing in common",
			a:    []string{"a"},
			b:    []string{"z"},
			want: "(0,-1) (-1,0)",
		},
		{
			name: "empty before",
			a:    nil,
			b:    []string{"a"},
			want: "(-1,0)",
		},
		{
			name: "empty after",
			a:    []string{"a"},
			b:    nil,
			want: "(0,-1)",
		},
		{
			name: "both empty",
			a:    nil,
			b:    nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatOps(Align(tt.a, tt.b)); got != tt.want {
				t.Errorf("Align(%v, %v) = %s, want %s", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestBuildTableRefusesOversizedInput(t *testing.T) {
	tests := []struct {
		name    string
		aLen    int
		bLen    int
		wantNil bool
	}{
		{name: "within the cell budget", aLen: 100, bLen: 100},
		{name: "too many cells", aLen: 600, bLen: 601, wantNil: true},
		{name: "one side longer than the budget", aLen: cellLimit + 1, bLen: 1, wantNil: true},
		{name: "empty side is always allowed", aLen: 0, bLen: cellLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTable(make([]string, tt.aLen), make([]string, tt.bLen))
			if (got == nil) != tt.wantNil {
				t.Errorf("buildTable(%d, %d) nil = %v, want %v", tt.aLen, tt.bLen, got == nil, tt.wantNil)
			}
		})
	}
}

func TestAlignFallsBackToPositionalWhenHuge(t *testing.T) {
	a := make([]string, 600)
	b := make([]string, 600)
	for i := range a {
		a[i] = fmt.Sprintf("k%d", i)
		b[i] = fmt.Sprintf("k%d", i)
	}

	ops := Align(a, append(b, "new"))
	if len(ops) != 601 {
		t.Fatalf("ops = %d, want 601", len(ops))
	}
	for i := range 600 {
		if ops[i] != (Op{A: i, B: i}) {
			t.Fatalf("ops[%d] = %v, want positional pairing", i, ops[i])
		}
	}
	if last := ops[600]; last != (Op{A: -1, B: 600}) {
		t.Errorf("trailing op = %v, want an addition", last)
	}
}
