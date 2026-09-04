package optimize

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func TestClampQuality(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{"in range", 80, 80},
		{"floor", 0, 1},
		{"negative", -50, 1},
		{"ceiling", 100, 100},
		{"above ceiling", 300, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClampQuality(tt.in); got != tt.want {
				t.Errorf("ClampQuality(%d) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestEncodeProducesDecodableImages(t *testing.T) {
	img := gradientImage(64, 48)

	tests := []struct {
		name   string
		format Format
		magic  []byte
	}{
		{"jpeg", FormatJPEG, []byte{0xFF, 0xD8, 0xFF}},
		{"png", FormatPNG, []byte{0x89, 'P', 'N', 'G'}},
		{"gif", FormatGIF, []byte("GIF8")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Encode(img, tt.format, 80)
			if err != nil {
				t.Fatalf("Encode: %v", err)
			}
			if !bytes.HasPrefix(data, tt.magic) {
				t.Errorf("output does not start with the %s magic bytes: % x", tt.name, data[:min(8, len(data))])
			}

			decoded, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("re-decode %s: %v", tt.name, err)
			}
			if b := decoded.Bounds(); b.Dx() != 64 || b.Dy() != 48 {
				t.Errorf("decoded bounds = %dx%d, want 64x48", b.Dx(), b.Dy())
			}
		})
	}
}

func TestEncodeRejectsUnwritableFormats(t *testing.T) {
	img := gradientImage(8, 8)

	for _, f := range []Format{FormatWebP, FormatTIFF, FormatBMP, FormatUnknown} {
		if _, err := Encode(img, f, 80); err == nil {
			t.Errorf("Encode to %v should fail while only jpeg, png and gif can be written", f)
		}
	}
}

func TestEncodeJPEGQualityReducesSize(t *testing.T) {
	img := gradientImage(128, 128)

	high, err := Encode(img, FormatJPEG, 95)
	if err != nil {
		t.Fatalf("Encode high: %v", err)
	}
	low, err := Encode(img, FormatJPEG, 30)
	if err != nil {
		t.Fatalf("Encode low: %v", err)
	}

	if len(low) >= len(high) {
		t.Errorf("quality 30 produced %d bytes, not smaller than quality 95 at %d bytes", len(low), len(high))
	}
}

// Dithered quantization can inflate a smooth image well past its lossless
// size, so lowering quality must never grow the file.
func TestEncodePNGLowerQualityNeverGrowsTheFile(t *testing.T) {
	images := map[string]image.Image{
		"smooth gradient": gradientImage(128, 128),
		"noisy":           noiseImage(128, 128),
		"flat":            solidImage(128, 128, color.RGBA{90, 120, 200, 255}),
	}

	for name, img := range images {
		t.Run(name, func(t *testing.T) {
			lossless, err := Encode(img, FormatPNG, 100)
			if err != nil {
				t.Fatalf("Encode lossless: %v", err)
			}
			for _, q := range []int{90, 70, 40, 10} {
				got, err := Encode(img, FormatPNG, q)
				if err != nil {
					t.Fatalf("Encode q%d: %v", q, err)
				}
				if len(got) > len(lossless) {
					t.Errorf("quality %d produced %d bytes, larger than the lossless %d bytes",
						q, len(got), len(lossless))
				}
			}
		})
	}
}

// On a noisy, photo-like image the palette reduction is a real win, which is
// the case the quality knob exists for.
func TestEncodePNGQuantizationShrinksNoisyImages(t *testing.T) {
	img := noiseImage(128, 128)

	lossless, err := Encode(img, FormatPNG, 100)
	if err != nil {
		t.Fatalf("Encode lossless: %v", err)
	}
	quantized, err := Encode(img, FormatPNG, 40)
	if err != nil {
		t.Fatalf("Encode quantized: %v", err)
	}

	if len(quantized) >= len(lossless) {
		t.Errorf("quantized png is %d bytes, expected a real saving against the lossless %d bytes",
			len(quantized), len(lossless))
	}
}

func TestEncodePNGAtFullQualityStaysTruecolor(t *testing.T) {
	img := gradientImage(32, 32)

	data, err := Encode(img, FormatPNG, 100)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("re-decode: %v", err)
	}
	if _, paletted := decoded.(*image.Paletted); paletted {
		t.Error("quality 100 should keep the png truecolor, not palettize it")
	}
}

func TestQuantizedForPassesThroughAtFullQuality(t *testing.T) {
	img := gradientImage(8, 8)

	if got := quantizedFor(img, 100); got != image.Image(img) {
		t.Error("quality 100 should hand back the original image untouched")
	}
	if got := quantizedFor(img, 99); got == image.Image(img) {
		t.Error("quality below 100 should palettize the image")
	}
}
