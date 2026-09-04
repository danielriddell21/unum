package optimize

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
)

func ClampQuality(q int) int {
	return min(max(q, 1), 100)
}

func Encode(img image.Image, format Format, quality int) ([]byte, error) {
	quality = ClampQuality(quality)
	var buf bytes.Buffer

	switch format {
	case FormatJPEG:
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, fmt.Errorf("encode jpeg: %w", err)
		}
	case FormatPNG:
		return encodePNG(img, quality)
	case FormatGIF:
		if err := gif.Encode(&buf, toPaletted(img, Quantize(img, ColorsForQuality(quality))), nil); err != nil {
			return nil, fmt.Errorf("encode gif: %w", err)
		}
	default:
		return nil, fmt.Errorf("cannot write %s: output format must be jpeg, png, or gif", format)
	}

	return buf.Bytes(), nil
}

func encodePNG(img image.Image, quality int) ([]byte, error) {
	lossless, err := encodePNGImage(img)
	if err != nil {
		return nil, err
	}
	// Quality 100 keeps every original pixel; below that the palette shrinks,
	// which is where the size normally comes off a PNG.
	if quality >= 100 {
		return lossless, nil
	}

	quantized, err := encodePNGImage(quantizedFor(img, quality))
	if err != nil {
		return nil, err
	}
	// Dithering scatters per-pixel noise that deflate cannot pack, so on smooth
	// or already-flat images the quantized file comes out larger than the
	// original. An optimizer must never hand back the bigger of the two.
	if len(lossless) <= len(quantized) {
		return lossless, nil
	}
	return quantized, nil
}

func encodePNGImage(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode png: %w", err)
	}
	return buf.Bytes(), nil
}

func quantizedFor(img image.Image, quality int) image.Image {
	if quality >= 100 {
		return img
	}
	return toPaletted(img, Quantize(img, ColorsForQuality(quality)))
}
