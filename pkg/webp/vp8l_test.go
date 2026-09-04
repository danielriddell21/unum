package webp

import (
	"math"
	"testing"
)

// decodeLZ77Param mirrors the decoder's lz77Param so prefixEncode can be
// checked against the thing it must invert.
func decodeLZ77Param(sym int, extra uint32) int {
	if sym < 4 {
		return sym + 1
	}
	extraBits := uint(sym-2) >> 1
	offset := (2 + sym&1) << extraBits
	return offset + int(extra) + 1
}

func TestPrefixEncodeInvertsTheDecoder(t *testing.T) {
	for v := 1; v <= 5000; v++ {
		sym, extraBits, extra := prefixEncode(v)

		if sym < 0 || sym >= nDistanceCodes {
			t.Fatalf("value %d produced symbol %d, outside the alphabet", v, sym)
		}
		if extra >= 1<<extraBits && extraBits > 0 {
			t.Fatalf("value %d produced %d extra in %d bits", v, extra, extraBits)
		}
		if got := decodeLZ77Param(sym, extra); got != v {
			t.Fatalf("value %d round-tripped to %d (sym %d, extra %d)", v, got, sym, extra)
		}
	}
}

func TestPrefixEncodeSmallValuesUseNoExtraBits(t *testing.T) {
	for v := 1; v <= 4; v++ {
		sym, extraBits, extra := prefixEncode(v)
		if sym != v-1 || extraBits != 0 || extra != 0 {
			t.Errorf("prefixEncode(%d) = (%d, %d, %d), want (%d, 0, 0)", v, sym, extraBits, extra, v-1)
		}
	}
}

func TestPrefixEncodeStaysWithinTheLengthAlphabet(t *testing.T) {
	sym, _, _ := prefixEncode(maxMatch)
	if sym >= nLengthCodes {
		t.Errorf("the longest match encodes to symbol %d, outside the %d length codes", sym, nLengthCodes)
	}
}

func TestDistanceCodesPreferTheCompactTable(t *testing.T) {
	dc := newDistanceCodes(64)

	// A distance of one pixel is the commonest case and must map into the
	// cheap two-dimensional table rather than the 120+d fallback.
	if got := dc.code(1); got > len(distanceMapTable) {
		t.Errorf("distance 1 coded as %d, want a small table code", got)
	}
}

func TestDistanceCodesFallBackForLargeDistances(t *testing.T) {
	dc := newDistanceCodes(64)
	const far = 100000

	if got := dc.code(far); got != far+len(distanceMapTable) {
		t.Errorf("distance %d coded as %d, want %d", far, got, far+len(distanceMapTable))
	}
}

// The decoder turns a code back into a distance; every code we emit must
// survive that trip.
func TestDistanceCodesRoundTrip(t *testing.T) {
	const width = 64
	dc := newDistanceCodes(width)

	decode := func(code int) int {
		if code > len(distanceMapTable) {
			return code - len(distanceMapTable)
		}
		distCode := int(distanceMapTable[code-1])
		d := (distCode>>4)*width + (8 - distCode&0xf)
		return max(d, 1)
	}

	for dist := 1; dist <= 2000; dist++ {
		if got := decode(dc.code(dist)); got != dist {
			t.Fatalf("distance %d coded as %d, which decodes to %d", dist, dc.code(dist), got)
		}
	}
}

func TestEntropy(t *testing.T) {
	tests := []struct {
		name string
		hist []int
		want float64
	}{
		{"empty", []int{}, 0},
		{"single symbol", []int{100}, 0},
		{"two equal symbols", []int{50, 50}, 1},
		{"four equal symbols", []int{1, 1, 1, 1}, 2},
		{"all zero", []int{0, 0, 0}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := entropy(tt.hist); math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("entropy(%v) = %v, want %v", tt.hist, got, tt.want)
			}
		})
	}
}

func TestApplySubtractGreenIsReversible(t *testing.T) {
	original := []uint32{0xFF102030, 0xFFAABBCC, 0x80000000, 0xFFFFFFFF}
	work := append([]uint32(nil), original...)

	applySubtractGreen(work)

	// Undo it the way the decoder does: add green back into red and blue.
	for i, p := range work {
		g := uint8(p >> 8)
		r := uint8(p>>16) + g
		b := uint8(p) + g
		work[i] = p&0xff00ff00 | uint32(r)<<16 | uint32(b)
	}

	for i := range original {
		if work[i] != original[i] {
			t.Errorf("pixel %d = %08x after inverting, want %08x", i, work[i], original[i])
		}
	}
}

func TestWantSubtractGreen(t *testing.T) {
	// Grey pixels have r == g == b, so subtracting green collapses red and
	// blue to a single value — a clear win.
	grey := make([]uint32, 256)
	for i := range grey {
		v := uint32(i)
		grey[i] = 0xFF000000 | v<<16 | v<<8 | v
	}
	if !wantSubtractGreen(grey) {
		t.Error("grey ramp should favour subtract-green")
	}

	// Red and blue already constant, green varying: subtracting would spread
	// them out instead.
	flat := make([]uint32, 256)
	for i := range flat {
		flat[i] = 0xFF000000 | 0x10<<16 | uint32(i)<<8 | 0x20
	}
	if wantSubtractGreen(flat) {
		t.Error("constant red and blue should not favour subtract-green")
	}
}

func TestChooseCacheBits(t *testing.T) {
	small := make([]uint32, 100)
	if got := chooseCacheBits(small); got != 0 {
		t.Errorf("cache bits = %d for a tiny image, want 0", got)
	}

	varied := make([]uint32, 4096)
	for i := range varied {
		varied[i] = uint32(i)
	}
	got := chooseCacheBits(varied)
	if got < 1 || got > 10 {
		t.Errorf("cache bits = %d, want 1-10", got)
	}
}

func TestTokenizeCoversEveryPixel(t *testing.T) {
	argb := make([]uint32, 500)
	for i := range argb {
		argb[i] = uint32(i % 7)
	}

	tokens := tokenize(argb, 4)

	covered := 0
	for _, tok := range tokens {
		switch tok.kind {
		case tokenCopy:
			covered += tok.length
		default:
			covered++
		}
	}
	if covered != len(argb) {
		t.Errorf("tokens cover %d pixels, want %d", covered, len(argb))
	}
}

func TestTokenizeFindsRunsInFlatData(t *testing.T) {
	argb := make([]uint32, 1000) // all identical

	tokens := tokenize(argb, 0)

	if len(tokens) > 10 {
		t.Errorf("a flat run produced %d tokens, expected a handful of long copies", len(tokens))
	}
	var sawCopy bool
	for _, tok := range tokens {
		if tok.kind == tokenCopy {
			sawCopy = true
			if tok.dist < 1 {
				t.Errorf("copy token has distance %d", tok.dist)
			}
		}
	}
	if !sawCopy {
		t.Error("expected at least one backward reference in a flat image")
	}
}
