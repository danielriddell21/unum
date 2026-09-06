package webp

import "testing"

// expandRLE replays a code-length token stream the way the decoder does, so
// the encoding can be checked against what it will actually mean.
func expandRLE(tokens []clToken, n int) []uint8 {
	out := make([]uint8, 0, n)
	prev := uint8(8)
	for _, tok := range tokens {
		switch {
		case tok.sym < 16:
			out = append(out, uint8(tok.sym))
			if tok.sym != 0 {
				prev = uint8(tok.sym)
			}
		case tok.sym == 16:
			for range int(tok.extra) + 3 {
				out = append(out, prev)
			}
		case tok.sym == 17:
			for range int(tok.extra) + 3 {
				out = append(out, 0)
			}
		case tok.sym == 18:
			for range int(tok.extra) + 11 {
				out = append(out, 0)
			}
		}
	}
	return out
}

func TestCodeLengthRLERoundTrips(t *testing.T) {
	tests := []struct {
		name    string
		lengths []uint8
	}{
		{"empty run", []uint8{}},
		{"single", []uint8{4}},
		{"no repeats", []uint8{1, 2, 3, 4, 5}},
		{"short zero run", []uint8{3, 0, 0, 3}},
		{"medium zero run", []uint8{3, 0, 0, 0, 0, 0, 3}},
		{"long zero run", zeros(200, 5)},
		{"repeat run", []uint8{7, 7, 7, 7, 7, 7, 7, 7, 7}},
		{"long repeat run", repeated(9, 100)},
		{"mixed", []uint8{0, 0, 0, 5, 5, 5, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}},
		{"all zero", make([]uint8, 50)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandRLE(codeLengthRLE(tt.lengths), len(tt.lengths))

			if len(got) != len(tt.lengths) {
				t.Fatalf("expanded to %d lengths, want %d", len(got), len(tt.lengths))
			}
			for i := range tt.lengths {
				if got[i] != tt.lengths[i] {
					t.Fatalf("length %d = %d, want %d", i, got[i], tt.lengths[i])
				}
			}
		})
	}
}

func TestCodeLengthRLEUsesLongRunCodes(t *testing.T) {
	tokens := codeLengthRLE(make([]uint8, 138))

	if len(tokens) != 1 || tokens[0].sym != 18 {
		t.Errorf("138 zeroes encoded as %d tokens (first sym %d), want a single code 18",
			len(tokens), tokens[0].sym)
	}
}

func TestCodeLengthRLESplitsOverlongRuns(t *testing.T) {
	// 300 zeroes exceeds what one code 18 can carry.
	got := expandRLE(codeLengthRLE(make([]uint8, 300)), 300)

	if len(got) != 300 {
		t.Errorf("expanded to %d lengths, want 300", len(got))
	}
}

func TestCanWriteSimple(t *testing.T) {
	tests := []struct {
		name    string
		present []int
		want    bool
	}{
		{"one small symbol", []int{5}, true},
		{"two small symbols", []int{3, 200}, true},
		{"three symbols", []int{1, 2, 3}, false},
		{"symbol out of byte range", []int{300}, false},
		{"one small one large", []int{5, 280}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canWriteSimple(huffman{present: tt.present}); got != tt.want {
				t.Errorf("canWriteSimple(%v) = %v, want %v", tt.present, got, tt.want)
			}
		})
	}
}

// A single-symbol code spends no bits, but the table it declares must still
// name one symbol or the decoder rejects the tree as empty.
func TestDeclaredLengthsMarksALoneSymbol(t *testing.T) {
	h := buildHuffman(append([]int{0, 0, 42}, make([]int, 10)...), maxCodeLength)

	if h.lengths[2] != 0 {
		t.Fatalf("actual length = %d, want 0 bits emitted", h.lengths[2])
	}
	if got := declaredLengths(h)[2]; got != 1 {
		t.Errorf("declared length = %d, want 1", got)
	}
}

func TestDeclaredLengthsLeavesRealCodesAlone(t *testing.T) {
	h := buildHuffman([]int{5, 3, 2, 1}, maxCodeLength)

	declared := declaredLengths(h)
	for i := range h.lengths {
		if declared[i] != h.lengths[i] {
			t.Errorf("symbol %d declared as %d but coded as %d", i, declared[i], h.lengths[i])
		}
	}
}

func zeros(n, marker int) []uint8 {
	out := make([]uint8, n)
	out[0] = uint8(marker)
	return out
}

func repeated(v uint8, n int) []uint8 {
	out := make([]uint8, n)
	for i := range out {
		out[i] = v
	}
	return out
}
