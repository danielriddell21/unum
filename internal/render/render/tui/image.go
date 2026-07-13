package tui

import (
	"fmt"
	"image"
	"math"
	"strings"

	"golang.org/x/image/draw"
)

func imageArt(img image.Image, cols, rows int) string {
	if img == nil || cols < 1 || rows < 1 {
		return ""
	}
	src := img.Bounds()
	iw, ih := src.Dx(), src.Dy()
	if iw == 0 || ih == 0 {
		return ""
	}

	// Each character cell shows two vertical pixels via the upper half block,
	// so the target raster is cols wide and rows*2 tall. Fit while preserving
	// aspect ratio, then centre the result within the pane.
	scale := math.Min(float64(cols)/float64(iw), float64(rows*2)/float64(ih))
	ow := max(1, int(float64(iw)*scale))
	oh := max(1, int(float64(ih)*scale))

	dst := image.NewRGBA(image.Rect(0, 0, ow, oh))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, src, draw.Over, nil)

	_, padTop := windowRange((oh+1)/2, rows)
	pad := strings.Repeat(" ", max(0, (cols-ow)/2))
	var b strings.Builder
	for i := 0; i < padTop; i++ {
		b.WriteByte('\n')
	}
	for y := 0; y < oh; y += 2 {
		b.WriteString(pad)
		for x := 0; x < ow; x++ {
			tr, tg, tb, _ := dst.At(x, y).RGBA()
			br, bg, bb := tr, tg, tb
			if y+1 < oh {
				br, bg, bb, _ = dst.At(x, y+1).RGBA()
			}
			fmt.Fprintf(&b, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀",
				tr>>8, tg>>8, tb>>8, br>>8, bg>>8, bb>>8)
		}
		b.WriteString("\x1b[0m\n")
	}
	return b.String()
}
