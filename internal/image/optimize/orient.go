package optimize

import (
	"encoding/binary"
	"image"
)

const orientationTag = 0x0112

func jpegOrientation(data []byte) int {
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return 1
	}

	i := 2
	for i+4 <= len(data) {
		if data[i] != 0xFF {
			return 1
		}
		marker := data[i+1]
		// Padding and standalone markers carry no length field.
		if marker == 0xFF || marker == 0x01 || (marker >= 0xD0 && marker <= 0xD9) {
			i += 2
			continue
		}
		// Start of scan: entropy-coded data follows, no more metadata segments.
		if marker == 0xDA {
			return 1
		}
		segLen := int(binary.BigEndian.Uint16(data[i+2:]))
		if segLen < 2 || i+2+segLen > len(data) {
			return 1
		}
		if marker == 0xE1 {
			if n, ok := exifOrientation(data[i+4 : i+2+segLen]); ok {
				return n
			}
		}
		i += 2 + segLen
	}
	return 1
}

func exifOrientation(seg []byte) (int, bool) {
	const header = "Exif\x00\x00"
	if len(seg) < len(header)+8 || string(seg[:len(header)]) != header {
		return 0, false
	}
	tiff := seg[len(header):]

	var bo binary.ByteOrder
	switch string(tiff[:2]) {
	case "II":
		bo = binary.LittleEndian
	case "MM":
		bo = binary.BigEndian
	default:
		return 0, false
	}
	if bo.Uint16(tiff[2:]) != 42 {
		return 0, false
	}

	off := int(bo.Uint32(tiff[4:]))
	if off < 8 || off+2 > len(tiff) {
		return 0, false
	}
	entry := off + 2
	for range int(bo.Uint16(tiff[off:])) {
		if entry+12 > len(tiff) {
			return 0, false
		}
		if bo.Uint16(tiff[entry:]) == orientationTag {
			v := int(bo.Uint16(tiff[entry+8:]))
			return v, v >= 1 && v <= 8
		}
		entry += 12
	}
	return 0, false
}

func applyOrientation(img image.Image, n int) image.Image {
	if n <= 1 || n > 8 {
		return img
	}

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dw, dh := w, h
	// Orientations 5-8 transpose the image, so width and height swap.
	if n >= 5 {
		dw, dh = h, w
	}

	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	for y := range h {
		for x := range w {
			dx, dy := orientedPoint(n, x, y, w, h)
			dst.Set(dx, dy, img.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}

func orientedPoint(n, x, y, w, h int) (int, int) {
	switch n {
	case 2:
		return w - 1 - x, y
	case 3:
		return w - 1 - x, h - 1 - y
	case 4:
		return x, h - 1 - y
	case 5:
		return y, x
	case 6:
		return h - 1 - y, x
	case 7:
		return h - 1 - y, w - 1 - x
	case 8:
		return y, w - 1 - x
	default:
		return x, y
	}
}
