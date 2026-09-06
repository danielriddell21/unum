package webp

const (
	transformTypePredictor = 0

	// predictorBits sets the tile size (2^bits) over which one predictor mode
	// is chosen. Four is libwebp's usual choice: small enough to follow local
	// structure, large enough that the mode sub-image stays cheap.
	predictorBits = 4
	nPredictors   = 14

	// minPredictorPixels is the point below which the mode sub-image costs
	// more than the prediction saves.
	minPredictorPixels = 256
)

func nTiles(size, bits int) int {
	return (size + 1<<bits - 1) >> bits
}

func avg2(a, b uint8) uint8 {
	return uint8((int32(a) + int32(b)) / 2)
}

func clampAddSubtractFull(a, b, c uint8) uint8 {
	x := int32(a) + int32(b) - int32(c)
	return uint8(min(max(x, 0), 255))
}

func clampAddSubtractHalf(a, b uint8) uint8 {
	x := int32(a) + (int32(a)-int32(b))/2
	return uint8(min(max(x, 0), 255))
}

func absInt32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

// predictors holds one function per predictor mode, in the order the format
// numbers them. Each returns the predicted RGBA bytes for the pixel at byte
// offset p, whose upward neighbour sits at offset top.
//
// The index arithmetic mirrors the decoder exactly, including its quirk at the
// right-hand edge: TR reads top+4, which for the last column is the first pixel
// of the current row rather than anything above it. Deviating there would
// desynchronise the two sides.
var predictors = [nPredictors]func(pix []byte, p, top int) [4]uint8{
	0:  func([]byte, int, int) [4]uint8 { return [4]uint8{0, 0, 0, 0xff} },
	1:  func(pix []byte, p, _ int) [4]uint8 { return quad(pix, p-4) },
	2:  func(pix []byte, _, top int) [4]uint8 { return quad(pix, top) },
	3:  func(pix []byte, _, top int) [4]uint8 { return quad(pix, top+4) },
	4:  func(pix []byte, _, top int) [4]uint8 { return quad(pix, top-4) },
	5:  perChannel(func(l, tl, t, tr uint8) uint8 { return avg2(avg2(l, tr), t) }),
	6:  perChannel(func(l, tl, t, tr uint8) uint8 { return avg2(l, tl) }),
	7:  perChannel(func(l, tl, t, tr uint8) uint8 { return avg2(l, t) }),
	8:  perChannel(func(l, tl, t, tr uint8) uint8 { return avg2(tl, t) }),
	9:  perChannel(func(l, tl, t, tr uint8) uint8 { return avg2(t, tr) }),
	10: perChannel(func(l, tl, t, tr uint8) uint8 { return avg2(avg2(l, tl), avg2(t, tr)) }),
	11: predictSelect,
	12: perChannel(func(l, tl, t, tr uint8) uint8 { return clampAddSubtractFull(l, t, tl) }),
	13: perChannel(func(l, tl, t, tr uint8) uint8 { return clampAddSubtractHalf(avg2(l, t), tl) }),
}

// quad reads the four bytes of one pixel.
func quad(pix []byte, at int) [4]uint8 {
	return [4]uint8(pix[at : at+4])
}

// perChannel lifts a scalar predictor over the four channels, handing it the
// left, top-left, top and top-right neighbours.
func perChannel(f func(l, tl, t, tr uint8) uint8) func(pix []byte, p, top int) [4]uint8 {
	return func(pix []byte, p, top int) [4]uint8 {
		var out [4]uint8
		for i := range 4 {
			out[i] = f(pix[p-4+i], pix[top-4+i], pix[top+i], pix[top+4+i])
		}
		return out
	}
}

// predictSelect picks whichever of L or T the gradient through TL points at.
func predictSelect(pix []byte, p, top int) [4]uint8 {
	var dl, dt int32
	for i := range 4 {
		l, tl, t := int32(pix[p-4+i]), int32(pix[top-4+i]), int32(pix[top+i])
		dl += absInt32(tl - t)
		dt += absInt32(tl - l)
	}
	if dl < dt {
		return quad(pix, p-4)
	}
	return quad(pix, top)
}

func predict(pix []byte, p, top, mode int) [4]uint8 {
	return predictors[mode](pix, p, top)
}

// residualCost approximates how expensive a residual byte is to code: values
// near zero (in either direction on the byte ring) compress well.
func residualCost(v uint8) int {
	return int(min(uint16(v), 256-uint16(v)))
}

// bestModeForTile tries every predictor over one tile and returns whichever
// leaves the residuals closest to zero.
func bestModeForTile(pix []byte, width, x0, x1, y0, y1 int) byte {
	var costs [nPredictors]int

	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			p := 4 * (y*width + x)
			top := p - 4*width
			for mode := range nPredictors {
				pred := predictors[mode](pix, p, top)
				for i := range 4 {
					costs[mode] += residualCost(pix[p+i] - pred[i])
				}
			}
		}
	}

	best := 0
	for mode, c := range costs {
		if c < costs[best] {
			best = mode
		}
	}
	return byte(best)
}

// chooseModes picks a predictor mode for every tile in the image.
func chooseModes(pix []byte, width, height int) []byte {
	tilesX, tilesY := nTiles(width, predictorBits), nTiles(height, predictorBits)
	modes := make([]byte, tilesX*tilesY)

	for ty := range tilesY {
		for tx := range tilesX {
			// Row 0 and column 0 use forced predictors, so they never
			// influence the choice.
			modes[ty*tilesX+tx] = bestModeForTile(pix, width,
				max(tx<<predictorBits, 1), min((tx+1)<<predictorBits, width),
				max(ty<<predictorBits, 1), min((ty+1)<<predictorBits, height))
		}
	}
	return modes
}

// applyPredictor replaces each pixel with its residual against the predicted
// value, returning the residual buffer and the per-tile modes.
//
// Predictions read the original pixels rather than the residuals: the
// transform is lossless, so what the decoder has reconstructed at each point is
// exactly what we are reading here.
func applyPredictor(pix []byte, width, height int) (residual, modes []byte) {
	modes = chooseModes(pix, width, height)
	tilesX := nTiles(width, predictorBits)

	residual = make([]byte, len(pix))
	copy(residual, pix)

	// The very first pixel is predicted as opaque black.
	residual[3] = pix[3] - 0xff

	// The rest of the top row is predicted from the pixel to its left.
	for x := 1; x < width; x++ {
		p := 4 * x
		for i := range 4 {
			residual[p+i] = pix[p+i] - pix[p-4+i]
		}
	}

	for y := 1; y < height; y++ {
		// The first column is predicted from the pixel above.
		p := 4 * y * width
		for i := range 4 {
			residual[p+i] = pix[p+i] - pix[p-4*width+i]
		}

		for x := 1; x < width; x++ {
			p := 4 * (y*width + x)
			top := p - 4*width
			mode := int(modes[(y>>predictorBits)*tilesX+(x>>predictorBits)])
			pred := predict(pix, p, top, mode)
			for i := range 4 {
				residual[p+i] = pix[p+i] - pred[i]
			}
		}
	}
	return residual, modes
}
