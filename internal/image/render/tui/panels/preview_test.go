package panels

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

func TestFitWithin(t *testing.T) {
	tests := []struct {
		name         string
		w, h         int
		maxW, maxH   int
		wantW, wantH int
	}{
		{"already fits", 10, 10, 40, 40, 10, 10},
		{"width bound", 100, 50, 20, 100, 20, 10},
		{"height bound", 50, 100, 100, 20, 10, 20},
		{"square into square", 80, 80, 20, 20, 20, 20},
		{"never upscales", 4, 4, 100, 100, 4, 4},
		{"zero source", 0, 10, 20, 20, 0, 0},
		{"zero bounds", 10, 10, 0, 20, 0, 0},
		{"floors at one pixel", 1000, 10, 5, 5, 5, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH := FitWithin(tt.w, tt.h, tt.maxW, tt.maxH)
			if gotW != tt.wantW || gotH != tt.wantH {
				t.Errorf("FitWithin(%d, %d, %d, %d) = (%d, %d), want (%d, %d)",
					tt.w, tt.h, tt.maxW, tt.maxH, gotW, gotH, tt.wantW, tt.wantH)
			}
		})
	}
}

func TestRenderPreviewEmptyCases(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))

	tests := []struct {
		name       string
		img        image.Image
		cols, rows int
	}{
		{"nil image", nil, 10, 10},
		{"no columns", img, 0, 10},
		{"no rows", img, 10, 0},
		{"negative", img, -5, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RenderPreview(tt.img, tt.cols, tt.rows); got != "" {
				t.Errorf("expected an empty preview, got %q", got)
			}
		})
	}
}

func TestRenderPreviewUsesHalfBlocks(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img.Set(1, 0, color.RGBA{255, 0, 0, 255})
	img.Set(0, 1, color.RGBA{0, 0, 255, 255})
	img.Set(1, 1, color.RGBA{0, 0, 255, 255})

	got := RenderPreview(img, 2, 1)

	if !strings.Contains(got, halfBlock) {
		t.Errorf("preview should use the half block character, got %q", got)
	}
	// Red on top, blue underneath: foreground red, background blue.
	if !strings.Contains(got, "\x1b[38;2;255;0;0m") {
		t.Errorf("expected a red foreground escape in %q", got)
	}
	if !strings.Contains(got, "\x1b[48;2;0;0;255m") {
		t.Errorf("expected a blue background escape in %q", got)
	}
	if !strings.HasSuffix(got, "\x1b[0m") {
		t.Errorf("preview should reset styling at the end, got %q", got)
	}
}

func TestRenderPreviewRowCount(t *testing.T) {
	// A square image in a wide box is bound by height: 4 rows of cells means
	// 8 pixel rows, so the preview should be 4 lines tall.
	got := RenderPreview(image.NewRGBA(image.Rect(0, 0, 64, 64)), 100, 4)

	if lines := strings.Count(got, "\n") + 1; lines != 4 {
		t.Errorf("preview has %d lines, want 4", lines)
	}
}

func TestRenderPreviewOddPixelHeightStillPairsUp(t *testing.T) {
	// 3 source rows scaled into a 2-row box must not read past the last pixel.
	got := RenderPreview(image.NewRGBA(image.Rect(0, 0, 3, 3)), 3, 2)

	if got == "" {
		t.Fatal("expected a preview")
	}
	if strings.Count(got, halfBlock) == 0 {
		t.Error("preview produced no cells")
	}
}

func TestPreviewPanelView(t *testing.T) {
	p := NewPreviewPanel(20, 10)

	if got := p.View(nil); !strings.Contains(got, "no image") {
		t.Errorf("empty panel should explain itself, got %q", got)
	}
	if got := p.View(image.NewRGBA(image.Rect(0, 0, 8, 8))); !strings.Contains(got, halfBlock) {
		t.Error("panel should render the image")
	}
}

func TestPreviewPanelResize(t *testing.T) {
	p := NewPreviewPanel(10, 5)
	p.Resize(40, 20)

	got := p.View(image.NewRGBA(image.Rect(0, 0, 100, 100)))

	// The taller panel should produce more rows than the small one did.
	small := NewPreviewPanel(10, 5)
	if strings.Count(got, "\n") <= strings.Count(small.View(image.NewRGBA(image.Rect(0, 0, 100, 100))), "\n") {
		t.Error("resizing to a larger panel should render more rows")
	}
}
