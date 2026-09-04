package webp

import (
	"image"
	"testing"
)

func TestNTiles(t *testing.T) {
	tests := []struct {
		name       string
		size, bits int
		want       int
	}{
		{"exact multiple", 32, 4, 2},
		{"rounds up", 33, 4, 3},
		{"smaller than a tile", 5, 4, 1},
		{"single pixel", 1, 4, 1},
		{"large", 1200, 4, 75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nTiles(tt.size, tt.bits); got != tt.want {
				t.Errorf("nTiles(%d, %d) = %d, want %d", tt.size, tt.bits, got, tt.want)
			}
		})
	}
}

func TestArithmeticHelpers(t *testing.T) {
	if got := avg2(10, 20); got != 15 {
		t.Errorf("avg2(10, 20) = %d, want 15", got)
	}
	if got := avg2(255, 0); got != 127 {
		t.Errorf("avg2(255, 0) = %d, want 127 (truncating)", got)
	}
	if got := clampAddSubtractFull(200, 200, 100); got != 255 {
		t.Errorf("clampAddSubtractFull = %d, want it clamped to 255", got)
	}
	if got := clampAddSubtractFull(10, 10, 200); got != 0 {
		t.Errorf("clampAddSubtractFull = %d, want it clamped to 0", got)
	}
	if got := clampAddSubtractHalf(100, 50); got != 125 {
		t.Errorf("clampAddSubtractHalf(100, 50) = %d, want 125", got)
	}
}

func TestResidualCostTreatsTheByteRingSymmetrically(t *testing.T) {
	// 255 is -1, so it should cost the same as +1, not 255.
	if residualCost(255) != residualCost(1) {
		t.Errorf("residualCost(255) = %d, want the same as residualCost(1) = %d",
			residualCost(255), residualCost(1))
	}
	if residualCost(0) != 0 {
		t.Errorf("residualCost(0) = %d, want 0", residualCost(0))
	}
	if residualCost(128) != 128 {
		t.Errorf("residualCost(128) = %d, want 128 — the far side of the ring", residualCost(128))
	}
}

// Residuals must invert exactly, or the image comes back wrong. This walks the
// forward transform and re-applies the predictions by hand.
func TestApplyPredictorIsInvertible(t *testing.T) {
	const w, h = 40, 24
	original := argbToBytes(toARGB(gradient(w, h)))

	residual, modes := applyPredictor(append([]byte(nil), original...), w, h)
	tilesX := nTiles(w, predictorBits)

	// Reconstruct the way the decoder does: in place, in raster order.
	got := append([]byte(nil), residual...)
	got[3] += 0xff
	for x := 1; x < w; x++ {
		p := 4 * x
		for i := range 4 {
			got[p+i] += got[p-4+i]
		}
	}
	for y := 1; y < h; y++ {
		p := 4 * y * w
		for i := range 4 {
			got[p+i] += got[p-4*w+i]
		}
		for x := 1; x < w; x++ {
			p := 4 * (y*w + x)
			top := p - 4*w
			mode := int(modes[(y>>predictorBits)*tilesX+(x>>predictorBits)])
			pred := predict(got, p, top, mode)
			for i := range 4 {
				got[p+i] += pred[i]
			}
		}
	}

	for i := range original {
		if got[i] != original[i] {
			t.Fatalf("byte %d = %d after inverting, want %d", i, got[i], original[i])
		}
	}
}

func TestChooseModesCoversEveryTile(t *testing.T) {
	const w, h = 70, 50
	pix := argbToBytes(toARGB(gradient(w, h)))

	modes := chooseModes(pix, w, h)

	if want := nTiles(w, predictorBits) * nTiles(h, predictorBits); len(modes) != want {
		t.Fatalf("got %d modes, want %d", len(modes), want)
	}
	for i, m := range modes {
		if m >= nPredictors {
			t.Errorf("tile %d has mode %d, outside 0-%d", i, m, nPredictors-1)
		}
	}
}

// A flat image is predicted perfectly by any neighbour, so every residual
// should be zero apart from the forced first pixel.
func TestApplyPredictorZeroesAFlatImage(t *testing.T) {
	const w, h = 32, 32
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = 200
	}

	residual, _ := applyPredictor(argbToBytes(toARGB(img)), w, h)

	for i := 4; i < len(residual); i++ {
		if residual[i] != 0 {
			t.Fatalf("residual byte %d = %d, want 0 on a flat image", i, residual[i])
		}
	}
}

func TestARGBByteRoundTrip(t *testing.T) {
	argb := []uint32{0xFF102030, 0x00AABBCC, 0x12345678}

	if got := bytesToARGB(argbToBytes(argb)); len(got) != len(argb) {
		t.Fatalf("length changed: %d, want %d", len(got), len(argb))
	} else {
		for i := range argb {
			if got[i] != argb[i] {
				t.Errorf("pixel %d = %08x, want %08x", i, got[i], argb[i])
			}
		}
	}
}

func TestModesToARGBPutsModeInGreen(t *testing.T) {
	got := modesToARGB([]byte{0, 5, 13})

	want := []uint32{0, 5 << 8, 13 << 8}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("mode %d packed as %08x, want %08x", i, got[i], want[i])
		}
	}
}
