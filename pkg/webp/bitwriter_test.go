package webp

import (
	"bytes"
	"testing"

	"golang.org/x/image/vp8l"
)

func TestBitWriterPacksLSBFirst(t *testing.T) {
	var w bitWriter
	w.writeBits(1, 1) // bit 0
	w.writeBits(0, 1) // bit 1
	w.writeBits(1, 1) // bit 2

	got := w.bytes()
	if len(got) != 1 || got[0] != 0b101 {
		t.Errorf("bytes = %08b, want 00000101", got)
	}
}

func TestBitWriterSpansByteBoundaries(t *testing.T) {
	var w bitWriter
	w.writeBits(0xFFFF, 16)

	if got := w.bytes(); !bytes.Equal(got, []byte{0xFF, 0xFF}) {
		t.Errorf("bytes = % x, want ff ff", got)
	}
}

func TestBitWriterWideValues(t *testing.T) {
	var w bitWriter
	w.writeBits(0x12345678, 32)

	// Little-endian bit order means the low byte comes out first.
	if got := w.bytes(); !bytes.Equal(got, []byte{0x78, 0x56, 0x34, 0x12}) {
		t.Errorf("bytes = % x, want 78 56 34 12", got)
	}
}

func TestBitWriterZeroBitsIsANoOp(t *testing.T) {
	var w bitWriter
	w.writeBits(0xFF, 0)

	if got := w.bytes(); len(got) != 0 {
		t.Errorf("bytes = % x, want empty", got)
	}
}

func TestBitWriterCodeIsMSBFirst(t *testing.T) {
	var w bitWriter
	w.writeCode(0b110, 3)

	// MSB (1) written first lands in bit 0, so the byte reads 011.
	if got := w.bytes(); got[0] != 0b011 {
		t.Errorf("byte = %08b, want 00000011", got[0])
	}
}

// The decoder is the authority on bit order, so read the header back with it.
func TestBitWriterRoundTripsAHeaderThroughTheDecoder(t *testing.T) {
	var w bitWriter
	w.writeBits(0x2f, 8)   // VP8L signature
	w.writeBits(640-1, 14) // width - 1
	w.writeBits(480-1, 14) // height - 1
	w.writeBits(0, 1)      // alpha hint
	w.writeBits(0, 3)      // version

	cfg, err := vp8l.DecodeConfig(bytes.NewReader(w.bytes()))
	if err != nil {
		t.Fatalf("DecodeConfig: %v", err)
	}
	if cfg.Width != 640 || cfg.Height != 480 {
		t.Errorf("decoded %dx%d, want 640x480", cfg.Width, cfg.Height)
	}
}
