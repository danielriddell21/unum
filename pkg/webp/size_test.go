package webp

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func pngSize(t *testing.T, m image.Image) int {
	t.Helper()
	var buf bytes.Buffer
	if err := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(&buf, m); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Len()
}

// The whole point of this encoder is to beat PNG. Noise is incompressible by
// either format, so it is allowed to come out marginally larger.
func TestSmallerThanPNG(t *testing.T) {
	tests := []struct {
		name      string
		img       image.Image
		tolerance float64 // permitted fraction of the PNG size
	}{
		{"flat", solid(256, 256, color.NRGBA{90, 120, 200, 255}), 0.5},
		{"gradient", gradient(320, 240), 0.8},
		{"checker", checker(128, 128), 0.5},
		{"grey ramp", greyRamp(256, 64), 0.5},
		{"photographic", photo(320, 240), 1.0},
		{"transparent", transparent(128, 128), 0.8},
		{"noise", noise(200, 150), 1.02},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := EncodeToBytes(tt.img)
			if err != nil {
				t.Fatalf("EncodeToBytes: %v", err)
			}

			want := float64(pngSize(t, tt.img)) * tt.tolerance
			if float64(len(data)) > want {
				t.Errorf("webp is %d bytes against a png of %d — over the %.0f%% budget",
					len(data), pngSize(t, tt.img), tt.tolerance*100)
			}
		})
	}
}

// A smooth gradient is the case the predictor transform exists for: without it
// every pixel codes as a raw literal and the file explodes.
func TestPredictorMakesGradientsCheap(t *testing.T) {
	img := gradient(320, 240)

	data, err := EncodeToBytes(img)
	if err != nil {
		t.Fatalf("EncodeToBytes: %v", err)
	}

	raw := 4 * 320 * 240
	if len(data) > raw/100 {
		t.Errorf("gradient encoded to %d bytes, over 1%% of the %d byte raw image — "+
			"prediction is probably not being applied", len(data), raw)
	}
}

func TestFlatImageIsTiny(t *testing.T) {
	data, err := EncodeToBytes(solid(512, 512, color.NRGBA{1, 2, 3, 255}))
	if err != nil {
		t.Fatalf("EncodeToBytes: %v", err)
	}

	if len(data) > 200 {
		t.Errorf("a single-colour 512x512 image encoded to %d bytes, want under 200", len(data))
	}
}

func photo(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	seed := uint32(99)
	grain := func() int {
		seed = seed*1664525 + 1013904223
		return int(seed>>22)%7 - 3
	}
	clamp := func(v int) uint8 { return uint8(min(max(v, 0), 255)) }

	for y := range h {
		for x := range w {
			fy := float64(y) / float64(h)
			r := int(40 + 150*fy)
			g := int(90 + 110*fy)
			b := int(190 - 50*fy)
			if y > h*2/3 {
				r, g, b = 60, 100, 50
			}
			img.SetNRGBA(x, y, color.NRGBA{clamp(r + grain()), clamp(g + grain()), clamp(b + grain()), 255})
		}
	}
	return img
}
