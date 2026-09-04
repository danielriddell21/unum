package optimize

import (
	"image"
	"image/color"
	"testing"
)

func TestColorsForQuality(t *testing.T) {
	tests := []struct {
		name    string
		quality int
		want    int
	}{
		{"full quality keeps the whole palette", 100, 256},
		{"high", 90, 230},
		{"default", 80, 204},
		{"half", 50, 128},
		{"low clamps to two", 0, 2},
		{"negative clamps to two", -20, 2},
		{"above full clamps to 256", 200, 256},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ColorsForQuality(tt.quality); got != tt.want {
				t.Errorf("ColorsForQuality(%d) = %d, want %d", tt.quality, got, tt.want)
			}
		})
	}
}

func TestQuantizeRespectsPaletteSize(t *testing.T) {
	img := gradientImage(64, 64)

	for _, n := range []int{2, 8, 64, 256} {
		pal := Quantize(img, n)
		if len(pal) > n {
			t.Errorf("Quantize(n=%d) returned %d colors, want at most %d", n, len(pal), n)
		}
		if len(pal) == 0 {
			t.Errorf("Quantize(n=%d) returned an empty palette", n)
		}
	}
}

func TestQuantizeClampsRequestedSize(t *testing.T) {
	img := gradientImage(32, 32)

	if got := len(Quantize(img, 1000)); got > 256 {
		t.Errorf("palette size = %d, want at most 256", got)
	}
	if got := len(Quantize(img, 0)); got > 2 {
		t.Errorf("palette size = %d, want at most 2", got)
	}
}

func TestQuantizeSolidImageCollapsesToOneColor(t *testing.T) {
	want := color.RGBA{40, 80, 160, 255}
	img := solidImage(32, 32, want)

	// A single-colour image has nothing to split, so the palette stays at one
	// bucket however many colours are asked for.
	pal := Quantize(img, 64)
	if len(pal) != 1 {
		t.Fatalf("palette size = %d, want 1", len(pal))
	}
	if got := pal[0].(color.RGBA); got != want {
		t.Errorf("palette color = %v, want %v", got, want)
	}
}

func TestQuantizeEmptyImage(t *testing.T) {
	pal := Quantize(image.NewRGBA(image.Rect(0, 0, 0, 0)), 16)

	if len(pal) != 1 {
		t.Errorf("palette size = %d, want a single fallback color", len(pal))
	}
}

func TestQuantizeSeparatesDistinctColors(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img.Set(1, 0, color.RGBA{0, 0, 255, 255})

	pal := Quantize(img, 2)

	if len(pal) != 2 {
		t.Fatalf("palette size = %d, want 2", len(pal))
	}
	// The two buckets should land near the two source colours, not on a
	// muddy average of them.
	for _, want := range []color.RGBA{{255, 0, 0, 255}, {0, 0, 255, 255}} {
		if pal.Convert(want) != color.Color(want) {
			t.Errorf("palette lost color %v; palette = %v", want, pal)
		}
	}
}

func TestToPalettedProducesPalettedImage(t *testing.T) {
	img := gradientImage(16, 16)

	got := toPaletted(img, Quantize(img, 16))

	if b := got.Bounds(); b.Dx() != 16 || b.Dy() != 16 {
		t.Errorf("bounds = %dx%d, want 16x16", b.Dx(), b.Dy())
	}
	if len(got.Palette) > 16 {
		t.Errorf("palette size = %d, want at most 16", len(got.Palette))
	}
}

func TestSamplePixelsCapsSampleCount(t *testing.T) {
	got := samplePixels(gradientImage(512, 512))

	if len(got) > maxSamples+1 {
		t.Errorf("sampled %d pixels, want at most %d", len(got), maxSamples+1)
	}
	if len(got) == 0 {
		t.Error("sampled no pixels")
	}
}
