package webp

import (
	"math"
	"math/bits"
)

const (
	nLiteralCodes  = 256
	nLengthCodes   = 24
	nDistanceCodes = 40

	// colorCacheMultiplier is the hash multiplier defined by the VP8L
	// specification, section 4.2.3.
	colorCacheMultiplier = 0x1e35a7bd

	transformTypeSubtractGreen = 2
)

// codeLengthCodeOrder is the order in which code-length code lengths appear in
// the bit-stream, from the specification's section 5.2.2.
var codeLengthCodeOrder = [19]uint8{
	17, 18, 0, 1, 2, 3, 4, 5, 16, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
}

// distanceMapTable maps the first 120 distance codes to a two-dimensional
// pixel offset, letting nearby pixels above the current row be referenced far
// more cheaply than a raw distance would allow.
var distanceMapTable = [120]uint8{
	0x18, 0x07, 0x17, 0x19, 0x28, 0x06, 0x27, 0x29, 0x16, 0x1a,
	0x26, 0x2a, 0x38, 0x05, 0x37, 0x39, 0x15, 0x1b, 0x36, 0x3a,
	0x25, 0x2b, 0x48, 0x04, 0x47, 0x49, 0x14, 0x1c, 0x35, 0x3b,
	0x46, 0x4a, 0x24, 0x2c, 0x58, 0x45, 0x4b, 0x34, 0x3c, 0x03,
	0x57, 0x59, 0x13, 0x1d, 0x56, 0x5a, 0x23, 0x2d, 0x44, 0x4c,
	0x55, 0x5b, 0x33, 0x3d, 0x68, 0x02, 0x67, 0x69, 0x12, 0x1e,
	0x66, 0x6a, 0x22, 0x2e, 0x54, 0x5c, 0x43, 0x4d, 0x65, 0x6b,
	0x32, 0x3e, 0x78, 0x01, 0x77, 0x79, 0x53, 0x5d, 0x11, 0x1f,
	0x64, 0x6c, 0x42, 0x4e, 0x76, 0x7a, 0x21, 0x2f, 0x75, 0x7b,
	0x31, 0x3f, 0x63, 0x6d, 0x52, 0x5e, 0x00, 0x74, 0x7c, 0x41,
	0x4f, 0x10, 0x20, 0x62, 0x6e, 0x30, 0x73, 0x7d, 0x51, 0x5f,
	0x40, 0x72, 0x7e, 0x61, 0x6f, 0x50, 0x71, 0x7f, 0x60, 0x70,
}

// prefixEncode splits an LZ77 length or distance into the symbol that carries
// it plus any trailing extra bits. It is the inverse of the decoder's
// lz77Param.
func prefixEncode(v int) (sym int, extraBits uint, extra uint32) {
	if v <= 4 {
		return v - 1, 0, 0
	}
	u := v - 1
	msb := bits.Len32(uint32(u)) - 1
	eb := msb - 1
	top := u >> uint(eb) // always 2 or 3
	sym = 2*eb + top
	offset := (2 + sym&1) << uint(eb)
	return sym, uint(eb), uint32(u - offset)
}

// distanceCodes maps a pixel distance to the cheapest code that expresses it.
type distanceCodes struct {
	byDistance map[int]int
}

func newDistanceCodes(width int) distanceCodes {
	m := make(map[int]int, len(distanceMapTable))
	for code := 1; code <= len(distanceMapTable); code++ {
		distCode := int(distanceMapTable[code-1])
		yOffset := distCode >> 4
		xOffset := 8 - distCode&0xf
		d := yOffset*width + xOffset
		if d < 1 {
			d = 1
		}
		// Codes ascend in cost, so the first one to reach a distance wins.
		if _, seen := m[d]; !seen {
			m[d] = code
		}
	}
	return distanceCodes{byDistance: m}
}

// code returns the distance code for a pixel distance. Distances beyond the
// two-dimensional table fall back to the plain offset encoding.
func (d distanceCodes) code(dist int) int {
	if c, ok := d.byDistance[dist]; ok {
		return c
	}
	return dist + len(distanceMapTable)
}

type tokenKind uint8

const (
	tokenLiteral tokenKind = iota
	tokenCopy
	tokenCache
)

type token struct {
	kind   tokenKind
	argb   uint32
	length int
	dist   int
	cache  uint32
}

// entropy returns the zero-order entropy of a histogram, in bits per symbol.
func entropy(hist []int) float64 {
	total := 0
	for _, c := range hist {
		total += c
	}
	if total == 0 {
		return 0
	}
	var h float64
	for _, c := range hist {
		if c > 0 {
			p := float64(c) / float64(total)
			h -= p * math.Log2(p)
		}
	}
	return h
}

// wantSubtractGreen reports whether storing red and blue relative to green
// would make them cheaper to code. It is a win on photographic content, where
// the channels move together, and a loss on palettes that were already flat.
func wantSubtractGreen(argb []uint32) bool {
	var plainR, plainB, subR, subB [256]int
	for _, p := range argb {
		g := uint8(p >> 8)
		r := uint8(p >> 16)
		b := uint8(p)
		plainR[r]++
		plainB[b]++
		subR[r-g]++
		subB[b-g]++
	}
	return entropy(subR[:])+entropy(subB[:]) < entropy(plainR[:])+entropy(plainB[:])
}

// applySubtractGreen stores red and blue as their difference from green.
func applySubtractGreen(argb []uint32) {
	for i, p := range argb {
		g := uint8(p >> 8)
		r := uint8(p>>16) - g
		b := uint8(p) - g
		argb[i] = p&0xff00ff00 | uint32(r)<<16 | uint32(b)
	}
}

// chooseCacheBits sizes the colour cache from how many distinct colours the
// image actually contains. Too large a cache pays for itself in a bloated
// Huffman table for the green channel.
func chooseCacheBits(argb []uint32) uint {
	if len(argb) < 512 {
		return 0
	}
	const cap = 1 << 12
	seen := make(map[uint32]struct{}, 1024)
	for _, p := range argb {
		seen[p] = struct{}{}
		if len(seen) >= cap {
			break
		}
	}
	n := uint(bits.Len(uint(len(seen))))
	return min(max(n, 1), 10)
}
