package optimize

import (
	"image"
	"image/color"
	"image/draw"
	"slices"
)

const maxSamples = 1 << 16

type sample struct{ r, g, b, a uint8 }

func (s sample) channel(i int) uint8 {
	switch i {
	case 0:
		return s.r
	case 1:
		return s.g
	case 2:
		return s.b
	default:
		return s.a
	}
}

func ColorsForQuality(quality int) int {
	n := quality * 256 / 100
	return min(max(n, 2), 256)
}

func Quantize(img image.Image, n int) color.Palette {
	n = min(max(n, 2), 256)

	buckets := [][]sample{samplePixels(img)}
	if len(buckets[0]) == 0 {
		return color.Palette{color.RGBA{}}
	}

	for len(buckets) < n {
		i, ch := widestBucket(buckets)
		if i < 0 {
			break
		}
		lo, hi := splitBucket(buckets[i], ch)
		buckets[i] = lo
		buckets = append(buckets, hi)
	}

	pal := make(color.Palette, 0, len(buckets))
	for _, b := range buckets {
		pal = append(pal, averageColor(b))
	}
	return pal
}

func samplePixels(img image.Image) []sample {
	b := img.Bounds()
	total := b.Dx() * b.Dy()
	if total == 0 {
		return nil
	}

	stride := 1
	if total > maxSamples {
		stride = total/maxSamples + 1
	}

	out := make([]sample, 0, min(total, maxSamples)+1)
	i := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if i%stride == 0 {
				r, g, bl, a := img.At(x, y).RGBA()
				out = append(out, sample{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), uint8(a >> 8)})
			}
			i++
		}
	}
	return out
}

func widestBucket(buckets [][]sample) (int, int) {
	bestBucket, bestChannel, bestRange := -1, 0, 0
	for i, b := range buckets {
		if len(b) < 2 {
			continue
		}
		for ch := range 4 {
			lo, hi := b[0].channel(ch), b[0].channel(ch)
			for _, s := range b[1:] {
				v := s.channel(ch)
				lo, hi = min(lo, v), max(hi, v)
			}
			if r := int(hi) - int(lo); r > bestRange {
				bestBucket, bestChannel, bestRange = i, ch, r
			}
		}
	}
	return bestBucket, bestChannel
}

func splitBucket(b []sample, ch int) (lo, hi []sample) {
	slices.SortFunc(b, func(x, y sample) int {
		return int(x.channel(ch)) - int(y.channel(ch))
	})
	mid := len(b) / 2
	return b[:mid], b[mid:]
}

func averageColor(b []sample) color.RGBA {
	if len(b) == 0 {
		return color.RGBA{}
	}
	var r, g, bl, a int
	for _, s := range b {
		r += int(s.r)
		g += int(s.g)
		bl += int(s.b)
		a += int(s.a)
	}
	n := len(b)
	return color.RGBA{uint8(r / n), uint8(g / n), uint8(bl / n), uint8(a / n)}
}

func toPaletted(img image.Image, pal color.Palette) *image.Paletted {
	b := img.Bounds()
	dst := image.NewPaletted(image.Rect(0, 0, b.Dx(), b.Dy()), pal)
	draw.FloydSteinberg.Draw(dst, dst.Bounds(), img, b.Min)
	return dst
}
