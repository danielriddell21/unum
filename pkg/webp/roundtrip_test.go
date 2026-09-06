package webp

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	xwebp "golang.org/x/image/webp"
)

// decodeBack runs the encoder's output through the reference decoder in
// golang.org/x/image, which is the authority on the format.
func decodeBack(t *testing.T, m image.Image) image.Image {
	t.Helper()

	data, err := EncodeToBytes(m)
	if err != nil {
		t.Fatalf("EncodeToBytes: %v", err)
	}
	got, err := xwebp.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode our own output: %v", err)
	}
	return got
}

// assertIdentical compares every pixel; lossless means exactly equal.
func assertIdentical(t *testing.T, want, got image.Image) {
	t.Helper()

	wb, gb := want.Bounds(), got.Bounds()
	if wb.Dx() != gb.Dx() || wb.Dy() != gb.Dy() {
		t.Fatalf("decoded %dx%d, want %dx%d", gb.Dx(), gb.Dy(), wb.Dx(), wb.Dy())
	}

	for y := range wb.Dy() {
		for x := range wb.Dx() {
			wr, wg, wbl, wa := want.At(wb.Min.X+x, wb.Min.Y+y).RGBA()
			gr, gg, gbl, ga := got.At(gb.Min.X+x, gb.Min.Y+y).RGBA()
			if wr != gr || wg != gg || wbl != gbl || wa != ga {
				t.Fatalf("pixel (%d,%d) = (%d,%d,%d,%d), want (%d,%d,%d,%d)",
					x, y, gr>>8, gg>>8, gbl>>8, ga>>8, wr>>8, wg>>8, wbl>>8, wa>>8)
			}
		}
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		img  image.Image
	}{
		{"single pixel", solid(1, 1, color.NRGBA{7, 8, 9, 255})},
		{"single row", gradient(64, 1)},
		{"single column", gradient(1, 64)},
		{"flat", solid(32, 32, color.NRGBA{10, 200, 30, 255})},
		{"gradient", gradient(64, 48)},
		{"noise", noise(64, 48)},
		{"two colors", checker(32, 32)},
		{"transparent", transparent(32, 32)},
		{"fully transparent", solid(16, 16, color.NRGBA{0, 0, 0, 0})},
		{"grey ramp", greyRamp(256, 4)},
		{"odd dimensions", gradient(37, 23)},
		{"wide", gradient(509, 3)},
		{"tall", gradient(3, 509)},
		{"large gradient", gradient(320, 240)},
		{"large noise", noise(200, 150)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertIdentical(t, tt.img, decodeBack(t, tt.img))
		})
	}
}

func TestRoundTripPreservesAlpha(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for y := range 16 {
		for x := range 16 {
			img.Set(x, y, color.NRGBA{uint8(x * 16), uint8(y * 16), 128, uint8(x * y)})
		}
	}

	assertIdentical(t, img, decodeBack(t, img))
}

func TestEncodeRejectsEmptyImages(t *testing.T) {
	if _, err := EncodeToBytes(image.NewNRGBA(image.Rect(0, 0, 0, 0))); err == nil {
		t.Error("expected an error for an empty image")
	}
}

func TestEncodeRejectsOversizedImages(t *testing.T) {
	// Deliberately not allocated: bounds alone are enough to reject it.
	huge := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	huge.Rect = image.Rect(0, 0, maxDimension+1, 1)

	if _, err := EncodeToBytes(huge); err == nil {
		t.Error("expected an error for an image over the dimension limit")
	}
}

func TestEncodeWritesToWriter(t *testing.T) {
	var buf bytes.Buffer
	if err := Encode(&buf, gradient(16, 16)); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if got := buf.Bytes(); !bytes.HasPrefix(got, []byte("RIFF")) {
		t.Errorf("output does not start with RIFF: % x", got[:min(8, len(got))])
	}
	if !bytes.Contains(buf.Bytes()[:16], []byte("WEBP")) {
		t.Error("output is missing the WEBP fourcc")
	}
	if !bytes.Contains(buf.Bytes()[:16], []byte("VP8L")) {
		t.Error("output is missing the VP8L chunk")
	}
}

func TestRIFFChunkIsEvenAligned(t *testing.T) {
	// A one-pixel image produces an odd-length chunk, which RIFF pads.
	data, err := EncodeToBytes(solid(1, 1, color.NRGBA{1, 2, 3, 255}))
	if err != nil {
		t.Fatalf("EncodeToBytes: %v", err)
	}

	if len(data)%2 != 0 {
		t.Errorf("container length %d is odd; RIFF chunks must be padded", len(data))
	}
}

func TestDecodeConfigMatches(t *testing.T) {
	data, err := EncodeToBytes(gradient(97, 61))
	if err != nil {
		t.Fatalf("EncodeToBytes: %v", err)
	}

	cfg, err := xwebp.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("DecodeConfig: %v", err)
	}
	if cfg.Width != 97 || cfg.Height != 61 {
		t.Errorf("config = %dx%d, want 97x61", cfg.Width, cfg.Height)
	}
}

// ── fixtures ────────────────────────────────────────────────────────────────

func solid(w, h int, c color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}

func gradient(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(x * 255 / max(w-1, 1)),
				G: uint8(y * 255 / max(h-1, 1)),
				B: uint8((x + y) * 255 / max(w+h-2, 1)),
				A: 255,
			})
		}
	}
	return img
}

func noise(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	seed := uint32(0x12345678)
	next := func() uint8 {
		seed = seed*1664525 + 1013904223
		return uint8(seed >> 16)
	}
	for y := range h {
		for x := range w {
			img.SetNRGBA(x, y, color.NRGBA{next(), next(), next(), 255})
		}
	}
	return img
}

func checker(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			c := color.NRGBA{0, 0, 0, 255}
			if (x/4+y/4)%2 == 0 {
				c = color.NRGBA{255, 255, 255, 255}
			}
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}

func transparent(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetNRGBA(x, y, color.NRGBA{200, 100, 50, uint8(x * 255 / max(w-1, 1))})
		}
	}
	return img
}

func greyRamp(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			v := uint8(x % 256)
			img.SetNRGBA(x, y, color.NRGBA{v, v, v, 255})
		}
	}
	return img
}
