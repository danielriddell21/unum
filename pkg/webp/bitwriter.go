package webp

// bitWriter accumulates a VP8L bit-stream.
//
// VP8L packs bits least-significant-first within each byte: the first bit
// written is bit 0 of the first byte. Huffman codes are the exception — the
// decoder walks them starting from the most significant bit — so they go out
// through writeCode rather than writeBits.
type bitWriter struct {
	buf   []byte
	acc   uint64
	nBits uint
}

// writeBits appends the low n bits of v, least significant bit first.
func (w *bitWriter) writeBits(v uint32, n uint) {
	if n == 0 {
		return
	}
	w.acc |= uint64(v&(1<<n-1)) << w.nBits
	w.nBits += n
	for w.nBits >= 8 {
		w.buf = append(w.buf, byte(w.acc))
		w.acc >>= 8
		w.nBits -= 8
	}
}

// writeCode appends an n-bit Huffman code, most significant bit first.
func (w *bitWriter) writeCode(code uint32, n uint) {
	for i := int(n) - 1; i >= 0; i-- {
		w.writeBits(code>>uint(i)&1, 1)
	}
}

// bytes returns the stream, padding the final byte with zero bits.
func (w *bitWriter) bytes() []byte {
	out := w.buf
	if w.nBits > 0 {
		out = append(out, byte(w.acc))
	}
	return out
}
