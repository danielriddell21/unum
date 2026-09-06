package panels

import (
	"fmt"
	"image"
	"math"
	"strings"

	xdraw "golang.org/x/image/draw"
)

const halfBlock = "▀"

type PreviewPanel struct {
	width  int
	height int
}

func NewPreviewPanel(w, h int) PreviewPanel {
	return PreviewPanel{width: w, height: h}
}

func (p *PreviewPanel) Resize(w, h int) {
	p.width = w
	p.height = h
}

func (p *PreviewPanel) View(img image.Image) string {
	if img == nil {
		return styleHint.Render("  no image loaded")
	}
	return RenderPreview(img, p.width-2, p.height-2)
}

// RenderPreview draws img into a grid of terminal cells. Each cell carries two
// vertical pixels — the upper half block is the foreground colour and the cell
// background is the pixel below it — so a cell covers roughly a square of the
// image despite terminal cells being twice as tall as they are wide.
func RenderPreview(img image.Image, cols, rows int) string {
	if img == nil || cols < 1 || rows < 1 {
		return ""
	}

	pw, ph := FitWithin(img.Bounds().Dx(), img.Bounds().Dy(), cols, rows*2)
	if pw < 1 || ph < 1 {
		return ""
	}
	if ph%2 == 1 {
		ph++
	}

	scaled := image.NewRGBA(image.Rect(0, 0, pw, ph))
	xdraw.ApproxBiLinear.Scale(scaled, scaled.Bounds(), img, img.Bounds(), xdraw.Src, nil)

	var b strings.Builder
	for y := 0; y < ph; y += 2 {
		for x := range pw {
			top, bottom := scaled.RGBAAt(x, y), scaled.RGBAAt(x, y+1)
			fmt.Fprintf(&b, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm%s",
				top.R, top.G, top.B, bottom.R, bottom.G, bottom.B, halfBlock)
		}
		b.WriteString("\x1b[0m")
		if y+2 < ph {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func FitWithin(w, h, maxW, maxH int) (int, int) {
	if w <= 0 || h <= 0 || maxW <= 0 || maxH <= 0 {
		return 0, 0
	}
	s := math.Min(float64(maxW)/float64(w), float64(maxH)/float64(h))
	if s > 1 {
		s = 1
	}
	return max(int(float64(w)*s), 1), max(int(float64(h)*s), 1)
}
