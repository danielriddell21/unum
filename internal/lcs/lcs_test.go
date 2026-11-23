package lcs

import (
	"fmt"
	"math/rand/v2"
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
			linear, ok := alignLinear(tt.a, tt.b, workLimit)
			if !ok {
				t.Fatalf("alignLinear(%v, %v) ran out of budget", tt.a, tt.b)
			}
			if got := formatOps(linear); got != tt.want {
				t.Errorf("alignLinear(%v, %v) = %s, want %s", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func checkScript(t *testing.T, a, b []string, ops []Op) int {
	t.Helper()
	nextA, nextB, matched := 0, 0, 0
	for _, op := range ops {
		if op.A >= 0 {
			if op.A != nextA {
				t.Fatalf("ops visit A[%d], want A[%d]: %s", op.A, nextA, formatOps(ops))
			}
			nextA++
		}
		if op.B >= 0 {
			if op.B != nextB {
				t.Fatalf("ops visit B[%d], want B[%d]: %s", op.B, nextB, formatOps(ops))
			}
			nextB++
		}
		switch {
		case op.A >= 0 && op.B >= 0:
			if a[op.A] != b[op.B] {
				t.Fatalf("ops pair %q with %q", a[op.A], b[op.B])
			}
			matched++
		case op.A < 0 && op.B < 0:
			t.Fatalf("op %v touches neither side", op)
		}
	}
	if nextA != len(a) || nextB != len(b) {
		t.Fatalf("ops cover %d/%d of A and %d/%d of B", nextA, len(a), nextB, len(b))
	}
	return matched
}

func randomLines(r *rand.Rand, maxLen int, alphabet int) []string {
	lines := make([]string, r.IntN(maxLen+1))
	for i := range lines {
		lines[i] = string(rune('a' + r.IntN(alphabet)))
	}
	return lines
}

func TestAlignLinearMatchesTableOnRandomInput(t *testing.T) {
	r := rand.New(rand.NewPCG(47, 3776))
	for iter := range 3000 {
		alphabet := 1 + iter%6
		a := randomLines(r, 40, alphabet)
		b := randomLines(r, 40, alphabet)

		table := buildTable(a, b)
		want := table[0][0]
		if got := checkScript(t, a, b, walkTable(a, b, table)); got != want {
			t.Fatalf("table walk matched %d, want %d", got, want)
		}

		linear, ok := alignLinear(a, b, workLimit)
		if !ok {
			t.Fatalf("alignLinear(%v, %v) ran out of budget", a, b)
		}
		if got := checkScript(t, a, b, linear); got != want {
			t.Fatalf("alignLinear(%v, %v) matched %d, want LCS %d: %s", a, b, got, want, formatOps(linear))
		}

		again, _ := alignLinear(a, b, workLimit)
		if formatOps(again) != formatOps(linear) {
			t.Fatalf("alignLinear(%v, %v) is not deterministic", a, b)
		}
	}
}

func numbered(prefix string, n int) []string {
	lines := make([]string, n)
	for i := range lines {
		lines[i] = fmt.Sprintf("%s%d", prefix, i)
	}
	return lines
}

func TestComputeStaysExactBeyondTheTable(t *testing.T) {
	base := numbered("k", 600)
	withHeader := append([]string{"new"}, base...)

	res := Compute(base, withHeader)
	if res.Approximate {
		t.Fatal("Approximate = true, want an exact alignment")
	}
	if got := checkScript(t, base, withHeader, res.Ops); got != 600 {
		t.Fatalf("matched %d lines, want 600", got)
	}
	if first := res.Ops[0]; first != (Op{A: -1, B: 0}) {
		t.Errorf("first op = %v, want the inserted header", first)
	}
	for i, op := range res.Ops[1:] {
		if op != (Op{A: i, B: i + 1}) {
			t.Fatalf("ops[%d] = %v, want (%d,%d)", i+1, op, i, i+1)
		}
	}
}

func TestComputeLargeInputs(t *testing.T) {
	tests := []struct {
		name       string
		lines      int
		dropEvery  int
		addEvery   int
		wantRemove int
		wantAdd    int
	}{
		{name: "identical", lines: 5000},
		{name: "scattered edits", lines: 5000, dropEvery: 50, addEvery: 70, wantRemove: 100, wantAdd: 57},
		{name: "dense edits", lines: 8000, dropEvery: 3, addEvery: 4, wantRemove: 2667, wantAdd: 1333},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := numbered("line ", tt.lines)
			var b []string
			for i, l := range a {
				if tt.dropEvery > 0 && i%tt.dropEvery == 0 {
					continue
				}
				b = append(b, l)
				if tt.addEvery > 0 && i%tt.addEvery == 0 {
					b = append(b, fmt.Sprintf("added %d", i))
				}
			}

			res := Compute(a, b)
			if res.Approximate {
				t.Fatal("Approximate = true, want an exact alignment")
			}
			matched := checkScript(t, a, b, res.Ops)
			if removed := len(a) - matched; removed != tt.wantRemove {
				t.Errorf("removed = %d, want %d", removed, tt.wantRemove)
			}
			if added := len(b) - matched; added != tt.wantAdd {
				t.Errorf("added = %d, want %d", added, tt.wantAdd)
			}
		})
	}
}

func TestComputeDropsLinesUniqueToOneSide(t *testing.T) {
	a := numbered("old ", 3000)
	b := numbered("new ", 3000)
	for i := 0; i < len(a); i += 10 {
		b[i] = a[i]
	}

	res := compute(a, b, 0)
	if res.Approximate {
		t.Fatal("Approximate = true, want the shared lines aligned without any search")
	}
	if got := checkScript(t, a, b, res.Ops); got != 300 {
		t.Errorf("matched %d lines, want 300", got)
	}
}

func TestComputeFallsBackWhenOverBudget(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	a := make([]string, 600)
	b := make([]string, 600)
	for i := range a {
		a[i] = string(rune('a' + r.IntN(4)))
		b[i] = string(rune('a' + r.IntN(4)))
	}

	res := compute(a, b, 1000)
	if !res.Approximate {
		t.Fatal("Approximate = false, want the positional fallback flagged")
	}
	if got, want := formatOps(res.Ops), formatOps(positional(600, 600)); got != want {
		t.Errorf("ops are not positional")
	}

	if _, ok := alignLinear(a, b, workLimit); !ok {
		t.Error("alignLinear ran out of budget on a 600x600 input")
	}
}

func BenchmarkComputeScatteredEdits(b *testing.B) {
	a := numbered("line ", 10000)
	next := make([]string, 0, len(a))
	for i, l := range a {
		if i%40 == 0 {
			next = append(next, fmt.Sprintf("changed %d", i))
			continue
		}
		next = append(next, l)
	}
	for b.Loop() {
		Compute(a, next)
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
