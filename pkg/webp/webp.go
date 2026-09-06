// Package webp encodes images in the lossless WebP (VP8L) format.
//
// WebP's lossless mode routinely produces smaller files than PNG for the same
// pixels. This encoder implements the subtract-green transform, LZ77 backward
// references, a colour cache and canonical Huffman coding. It does not
// implement the predictor or cross-colour transforms, so on photographic
// content it will not match libwebp; on flat, palettised or synthetic images
// it is competitive.
//
// Lossy WebP (VP8) encoding is out of scope: that is a separate intra-frame
// video codec rather than a variation on this one.
package webp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	"io"
)

// maxDimension is the largest width or height VP8L can describe, as the
// bit-stream stores each dimension minus one in fourteen bits.
const maxDimension = 1 << 14

// Encode writes m to w in the lossless WebP format.
//
// The image is encoded exactly: decoding the result reproduces the original
// pixels. An image wider or taller than 16384 pixels cannot be represented and
// returns an error.
func Encode(w io.Writer, m image.Image) error {
	data, err := EncodeToBytes(m)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("webp: write: %w", err)
	}
	return nil
}

// EncodeToBytes returns the lossless WebP encoding of m.
//
// It is the allocation-friendly form of Encode for callers that need the bytes
// themselves, such as size comparisons against another encoder.
func EncodeToBytes(m image.Image) ([]byte, error) {
	bounds := m.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("webp: empty image")
	}
	if width > maxDimension || height > maxDimension {
		return nil, fmt.Errorf("webp: image is %d × %d, over the %d pixel limit in either direction",
			width, height, maxDimension)
	}

	return riffContainer(encodeVP8L(toARGB(m), width, height)), nil
}

// toARGB flattens an image into packed ARGB pixels.
func toARGB(m image.Image) []uint32 {
	bounds := m.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	nrgba, ok := m.(*image.NRGBA)
	if !ok || nrgba.Rect != bounds {
		converted := image.NewNRGBA(image.Rect(0, 0, width, height))
		draw.Draw(converted, converted.Bounds(), m, bounds.Min, draw.Src)
		nrgba = converted
	}

	out := make([]uint32, width*height)
	for y := range height {
		row := nrgba.Pix[y*nrgba.Stride:]
		for x := range width {
			p := row[4*x:]
			out[y*width+x] = uint32(p[3])<<24 | uint32(p[0])<<16 | uint32(p[1])<<8 | uint32(p[2])
		}
	}
	return out
}

// riffContainer wraps a VP8L bit-stream in the RIFF envelope WebP requires.
func riffContainer(vp8l []byte) []byte {
	chunkSize := len(vp8l)
	padded := chunkSize + chunkSize&1 // RIFF chunks are even-aligned

	var buf bytes.Buffer
	buf.Grow(20 + padded)
	buf.WriteString("RIFF")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(12+padded))
	buf.WriteString("WEBP")
	buf.WriteString("VP8L")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(chunkSize))
	buf.Write(vp8l)
	if padded != chunkSize {
		buf.WriteByte(0)
	}
	return buf.Bytes()
}

// encodeVP8L produces the raw VP8L bit-stream for the given pixels.
func encodeVP8L(argb []uint32, width, height int) []byte {
	var w bitWriter
	w.writeBits(0x2f, 8) // VP8L signature
	w.writeBits(uint32(width-1), 14)
	w.writeBits(uint32(height-1), 14)
	w.writeBits(alphaHint(argb), 1)
	w.writeBits(0, 3) // version

	if wantSubtractGreen(argb) {
		w.writeBits(1, 1) // a transform follows
		w.writeBits(transformTypeSubtractGreen, 2)
		applySubtractGreen(argb)
	}

	// The predictor transform stores each pixel as its difference from a
	// prediction made out of its neighbours, which is what makes smooth
	// content cheap. Without it a gradient costs as much as noise.
	if len(argb) >= minPredictorPixels && height > 1 {
		w.writeBits(1, 1)
		w.writeBits(transformTypePredictor, 2)
		w.writeBits(predictorBits-2, 3)

		residual, modes := applyPredictor(argbToBytes(argb), width, height)
		writeEntropyImage(&w, modesToARGB(modes), nTiles(width, predictorBits), false)
		argb = bytesToARGB(residual)
	}

	w.writeBits(0, 1) // no further transforms

	writeEntropyImage(&w, argb, width, true)
	return w.bytes()
}

func alphaHint(argb []uint32) uint32 {
	for _, p := range argb {
		if p>>24 != 0xff {
			return 1
		}
	}
	return 0
}

// argbToBytes lays pixels out as the RGBA bytes the predictor works on, in the
// same order the decoder uses.
func argbToBytes(argb []uint32) []byte {
	out := make([]byte, 4*len(argb))
	for i, p := range argb {
		out[4*i+0] = byte(p >> 16)
		out[4*i+1] = byte(p >> 8)
		out[4*i+2] = byte(p)
		out[4*i+3] = byte(p >> 24)
	}
	return out
}

func bytesToARGB(pix []byte) []uint32 {
	out := make([]uint32, len(pix)/4)
	for i := range out {
		out[i] = uint32(pix[4*i+3])<<24 | uint32(pix[4*i+0])<<16 |
			uint32(pix[4*i+1])<<8 | uint32(pix[4*i+2])
	}
	return out
}

// modesToARGB packs predictor modes into the sub-image the decoder reads them
// from. Only the green channel carries the mode; the rest stay zero so they
// cost nothing to code.
func modesToARGB(modes []byte) []uint32 {
	out := make([]uint32, len(modes))
	for i, m := range modes {
		out[i] = uint32(m) << 8
	}
	return out
}

// writeEntropyImage writes a Huffman-coded image. Only the top-level image
// carries the meta-Huffman flag; transform sub-images do not.
func writeEntropyImage(w *bitWriter, argb []uint32, width int, topLevel bool) {
	ccBits := chooseCacheBits(argb)
	if ccBits > 0 {
		w.writeBits(1, 1)
		w.writeBits(uint32(ccBits), 4)
	} else {
		w.writeBits(0, 1)
	}
	if topLevel {
		w.writeBits(0, 1) // a single Huffman group covers the whole image
	}

	dc := newDistanceCodes(width)
	tokens := tokenize(argb, ccBits)
	hist := newHistograms(tokens, ccBits, dc)

	green := buildHuffman(hist.green, maxCodeLength)
	red := buildHuffman(hist.red, maxCodeLength)
	blue := buildHuffman(hist.blue, maxCodeLength)
	alpha := buildHuffman(hist.alpha, maxCodeLength)
	dist := buildHuffman(hist.dist, maxCodeLength)

	for _, h := range []huffman{green, red, blue, alpha, dist} {
		writeHuffmanTree(w, h)
	}

	for _, t := range tokens {
		switch t.kind {
		case tokenLiteral:
			writeSymbol(w, green, int(t.argb>>8&0xff))
			writeSymbol(w, red, int(t.argb>>16&0xff))
			writeSymbol(w, blue, int(t.argb&0xff))
			writeSymbol(w, alpha, int(t.argb>>24&0xff))

		case tokenCopy:
			lsym, leb, lextra := prefixEncode(t.length)
			writeSymbol(w, green, nLiteralCodes+lsym)
			w.writeBits(lextra, leb)

			dsym, deb, dextra := prefixEncode(dc.code(t.dist))
			writeSymbol(w, dist, dsym)
			w.writeBits(dextra, deb)

		case tokenCache:
			writeSymbol(w, green, nLiteralCodes+nLengthCodes+int(t.cache))
		}
	}
}
