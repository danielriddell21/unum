package optimize

import (
	"image"
	"image/color"
	"testing"
)

func TestTargetDimensions(t *testing.T) {
	tests := []struct {
		name         string
		w, h         int
		scale        float64
		maxW, maxH   int
		wantW, wantH int
	}{
		{"unchanged", 800, 600, 1, 0, 0, 800, 600},
		{"zero scale means unchanged", 800, 600, 0, 0, 0, 800, 600},
		{"half", 800, 600, 0.5, 0, 0, 400, 300},
		{"scale rounds", 801, 601, 0.5, 0, 0, 401, 301},
		{"max width shrinks and keeps aspect", 800, 600, 1, 400, 0, 400, 300},
		{"max height shrinks and keeps aspect", 800, 600, 1, 0, 300, 400, 300},
		{"max width larger than source is a no-op", 800, 600, 1, 2000, 0, 800, 600},
		{"max height larger than source is a no-op", 800, 600, 1, 0, 2000, 800, 600},
		{"scale then max width", 800, 600, 0.5, 200, 0, 200, 150},
		{"tightest of both caps wins", 800, 600, 1, 400, 150, 200, 150},
		{"never smaller than one pixel", 10, 10, 0.01, 0, 0, 1, 1},
		{"portrait max width", 600, 800, 1, 300, 0, 300, 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH := TargetDimensions(tt.w, tt.h, tt.scale, tt.maxW, tt.maxH)
			if gotW != tt.wantW || gotH != tt.wantH {
				t.Errorf("TargetDimensions(%d, %d, %v, %d, %d) = (%d, %d), want (%d, %d)",
					tt.w, tt.h, tt.scale, tt.maxW, tt.maxH, gotW, gotH, tt.wantW, tt.wantH)
			}
		})
	}
}

func TestResampleChangesDimensions(t *testing.T) {
	src := solidImage(80, 60, color.RGBA{200, 100, 50, 255})

	got := Resample(src, 40, 30)

	if b := got.Bounds(); b.Dx() != 40 || b.Dy() != 30 {
		t.Errorf("Resample bounds = %dx%d, want 40x30", b.Dx(), b.Dy())
	}
}

func TestResampleSameSizeReturnsSourceUnchanged(t *testing.T) {
	src := solidImage(20, 10, color.RGBA{1, 2, 3, 255})

	got := Resample(src, 20, 10)

	if got != image.Image(src) {
		t.Error("Resample should return the source untouched when dimensions already match")
	}
}

func TestResamplePreservesSolidColor(t *testing.T) {
	want := color.RGBA{10, 120, 240, 255}
	src := solidImage(64, 64, want)

	got := Resample(src, 16, 16)

	r, g, b, a := got.At(8, 8).RGBA()
	if uint8(r>>8) != want.R || uint8(g>>8) != want.G || uint8(b>>8) != want.B || uint8(a>>8) != want.A {
		t.Errorf("centre pixel = (%d, %d, %d, %d), want (%d, %d, %d, %d)",
			r>>8, g>>8, b>>8, a>>8, want.R, want.G, want.B, want.A)
	}
}

func solidImage(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}
	return img
}

func gradientImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{
				R: uint8(x * 255 / max(w-1, 1)),
				G: uint8(y * 255 / max(h-1, 1)),
				B: uint8((x + y) * 255 / max(w+h-2, 1)),
				A: 255,
			})
		}
	}
	return img
}

// noiseImage builds a deterministic, photo-like image: enough distinct colors
// that a palette reduction is a genuine saving.
func noiseImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	seed := uint32(12345)
	next := func() uint32 {
		seed = seed*1664525 + 1013904223
		return seed >> 16
	}
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{uint8(next()), uint8(next()), uint8(next()), 255})
		}
	}
	return img
}
