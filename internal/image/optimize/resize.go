package optimize

import (
	"image"
	"math"

	xdraw "golang.org/x/image/draw"
)

func TargetDimensions(w, h int, scale float64, maxW, maxH int) (int, int) {
	if scale <= 0 {
		scale = 1
	}
	tw := int(math.Round(float64(w) * scale))
	th := int(math.Round(float64(h) * scale))

	if maxW > 0 && tw > maxW {
		th = int(math.Round(float64(th) * float64(maxW) / float64(tw)))
		tw = maxW
	}
	if maxH > 0 && th > maxH {
		tw = int(math.Round(float64(tw) * float64(maxH) / float64(th)))
		th = maxH
	}

	return max(tw, 1), max(th, 1)
}

func Resample(src image.Image, w, h int) image.Image {
	b := src.Bounds()
	if w == b.Dx() && h == b.Dy() {
		return src
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, b, xdraw.Src, nil)
	return dst
}
