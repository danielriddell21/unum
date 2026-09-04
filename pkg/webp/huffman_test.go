package webp

import (
	"math"
	"testing"
)

// kraftSum reports whether the code lengths form a complete prefix code:
// sum of 2^-length over all used symbols must equal exactly 1.
func kraftSum(lengths []uint8) float64 {
	var sum float64
	for _, l := range lengths {
		if l > 0 {
			sum += math.Ldexp(1, -int(l))
		}
	}
	return sum
}

func TestBuildHuffmanEmptyAlphabet(t *testing.T) {
	h := buildHuffman(make([]int, 20), maxCodeLength)

	if len(h.present) != 1 || h.present[0] != 0 {
		t.Errorf("present = %v, want a single placeholder symbol", h.present)
	}
	if h.lengths[0] != 0 {
		t.Errorf("placeholder length = %d, want 0", h.lengths[0])
	}
}

func TestBuildHuffmanSingleSymbolCostsNoBits(t *testing.T) {
	freq := make([]int, 20)
	freq[7] = 100

	h := buildHuffman(freq, maxCodeLength)

	if len(h.present) != 1 || h.present[0] != 7 {
		t.Fatalf("present = %v, want just symbol 7", h.present)
	}
	if h.lengths[7] != 0 {
		t.Errorf("length = %d, want 0 — a lone symbol needs no bits", h.lengths[7])
	}
}

func TestBuildHuffmanIsComplete(t *testing.T) {
	tests := []struct {
		name string
		freq []int
	}{
		{"two symbols", []int{5, 5}},
		{"uniform", []int{1, 1, 1, 1, 1, 1, 1, 1}},
		{"skewed", []int{1000, 100, 10, 1}},
		{"sparse", []int{0, 0, 42, 0, 7, 0, 0, 1}},
		{"powers of two", []int{1, 2, 4, 8, 16, 32, 64, 128}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := buildHuffman(tt.freq, maxCodeLength)

			if got := kraftSum(h.lengths); math.Abs(got-1) > 1e-9 {
				t.Errorf("Kraft sum = %v, want exactly 1 (lengths %v)", got, h.lengths)
			}
			for sym, f := range tt.freq {
				if f > 0 && h.lengths[sym] == 0 {
					t.Errorf("symbol %d has frequency %d but no code", sym, f)
				}
				if f == 0 && h.lengths[sym] != 0 {
					t.Errorf("symbol %d is unused but got length %d", sym, h.lengths[sym])
				}
			}
		})
	}
}

// A Fibonacci distribution is the classic case that drives a plain Huffman
// tree past 15 bits — exactly what VP8L rejects.
func TestPackageMergeRespectsTheLengthLimit(t *testing.T) {
	freq := make([]int, 40)
	a, b := 1, 1
	for i := range freq {
		freq[i] = a
		a, b = b, a+b
	}

	h := buildHuffman(freq, maxCodeLength)

	for sym, l := range h.lengths {
		if l > maxCodeLength {
			t.Errorf("symbol %d got length %d, over the %d bit limit", sym, l, maxCodeLength)
		}
	}
	if got := kraftSum(h.lengths); math.Abs(got-1) > 1e-9 {
		t.Errorf("Kraft sum = %v, want exactly 1", got)
	}
}

func TestPackageMergeTightLimit(t *testing.T) {
	freq := make([]int, 16)
	for i := range freq {
		freq[i] = 1 << i
	}

	// Sixteen symbols cannot all fit in fewer than 4 bits, so 4 is the
	// tightest legal limit and must still produce a complete code.
	h := buildHuffman(freq, 4)

	for sym, l := range h.lengths {
		if l > 4 {
			t.Errorf("symbol %d got length %d, over the 4 bit limit", sym, l)
		}
	}
	if got := kraftSum(h.lengths); math.Abs(got-1) > 1e-9 {
		t.Errorf("Kraft sum = %v, want exactly 1", got)
	}
}

// More frequent symbols must never get longer codes than rarer ones.
func TestBuildHuffmanOrdersLengthsByFrequency(t *testing.T) {
	freq := []int{100, 50, 25, 12, 6, 3, 2, 1}

	h := buildHuffman(freq, maxCodeLength)

	for i := 1; i < len(freq); i++ {
		if h.lengths[i] < h.lengths[i-1] {
			t.Errorf("symbol %d (freq %d) got length %d, shorter than symbol %d (freq %d) at length %d",
				i, freq[i], h.lengths[i], i-1, freq[i-1], h.lengths[i-1])
		}
	}
}

func TestAssignCodesIsPrefixFree(t *testing.T) {
	h := buildHuffman([]int{8, 4, 2, 1, 1}, maxCodeLength)

	type code struct {
		bits uint32
		n    uint8
	}
	var codes []code
	for sym, l := range h.lengths {
		if l > 0 {
			codes = append(codes, code{h.codes[sym], l})
		}
	}

	for i, a := range codes {
		for j, b := range codes {
			if i == j || a.n > b.n {
				continue
			}
			// a is no longer than b: b must not start with a.
			if b.bits>>(b.n-a.n) == a.bits {
				t.Errorf("code %0*b is a prefix of %0*b", a.n, a.bits, b.n, b.bits)
			}
		}
	}
}

func TestBuildHuffmanShortestCodeGoesToTheCommonestSymbol(t *testing.T) {
	freq := make([]int, 280)
	freq[42] = 10000
	for i := range 50 {
		freq[i+100] = 1
	}

	h := buildHuffman(freq, maxCodeLength)

	for sym, l := range h.lengths {
		if l > 0 && sym != 42 && l < h.lengths[42] {
			t.Errorf("symbol %d got a shorter code (%d) than the dominant symbol 42 (%d)",
				sym, l, h.lengths[42])
		}
	}
}
