package tui

import (
	"fmt"
	"image"
	"math"
	"strings"

	"golang.org/x/image/draw"
)

func imageArt(img image.Image, cols, rows int, zoom float64, panX, panY int) string {
	if img == nil || cols < 1 || rows < 1 {
		return ""
	}
	src := img.Bounds()
	iw, ih := src.Dx(), src.Dy()
	if iw == 0 || ih == 0 {
		return ""
	}
	if zoom <= 0 {
		zoom = 1
	}

	// Each character cell shows two vertical pixels via the upper half block,
	// so the viewport is cols wide and rows*2 tall. Fit while preserving aspect
	// ratio, then apply zoom so the scaled image can exceed the viewport.
	fit := math.Min(float64(cols)/float64(iw), float64(rows*2)/float64(ih))
	scale := fit * zoom
	ow := max(1, int(float64(iw)*scale))
	oh := max(1, int(float64(ih)*scale))

	dst := image.NewRGBA(image.Rect(0, 0, ow, oh))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, src, draw.Over, nil)

	// Window the scaled image into the viewport: center when it is smaller,
	// pan (clamped) when larger. Vertical units are terminal rows (2 px each).
	rowsImg := (oh + 1) / 2
	startCol, padLeft := windowRange(ow, cols, panX)
	startRow, padTop := windowRange(rowsImg, rows, panY)

	pad := strings.Repeat(" ", padLeft)
	var b strings.Builder
	for i := 0; i < padTop; i++ {
		b.WriteByte('\n')
	}
	for r := startRow; r < rowsImg && r < startRow+rows-padTop; r++ {
		y := r * 2
		b.WriteString(pad)
		for x := startCol; x < ow && x < startCol+cols-padLeft; x++ {
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
