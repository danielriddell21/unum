package optimize

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

var (
	colA = color.RGBA{255, 0, 0, 255}
	colB = color.RGBA{0, 255, 0, 255}
	colC = color.RGBA{0, 0, 255, 255}
	colD = color.RGBA{255, 255, 255, 255}
)

// quad builds a 2x2 image laid out as A B over C D, so every orientation
// produces a distinguishable arrangement.
func quad() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, colA)
	img.Set(1, 0, colB)
	img.Set(0, 1, colC)
	img.Set(1, 1, colD)
	return img
}

func TestApplyOrientation(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want [4]color.RGBA // reading order: (0,0) (1,0) (0,1) (1,1)
	}{
		{"1 normal", 1, [4]color.RGBA{colA, colB, colC, colD}},
		{"2 flip horizontal", 2, [4]color.RGBA{colB, colA, colD, colC}},
		{"3 rotate 180", 3, [4]color.RGBA{colD, colC, colB, colA}},
		{"4 flip vertical", 4, [4]color.RGBA{colC, colD, colA, colB}},
		{"5 transpose", 5, [4]color.RGBA{colA, colC, colB, colD}},
		{"6 rotate 90 clockwise", 6, [4]color.RGBA{colC, colA, colD, colB}},
		{"7 transverse", 7, [4]color.RGBA{colD, colB, colC, colA}},
		{"8 rotate 90 anticlockwise", 8, [4]color.RGBA{colB, colD, colA, colC}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyOrientation(quad(), tt.n)
			at := [4]color.RGBA{rgbaAt(got, 0, 0), rgbaAt(got, 1, 0), rgbaAt(got, 0, 1), rgbaAt(got, 1, 1)}
			if at != tt.want {
				t.Errorf("applyOrientation(n=%d) = %v, want %v", tt.n, at, tt.want)
			}
		})
	}
}

func TestApplyOrientationIgnoresOutOfRange(t *testing.T) {
	src := quad()
	for _, n := range []int{0, 1, 9, -3} {
		if got := applyOrientation(src, n); got != image.Image(src) {
			t.Errorf("applyOrientation(n=%d) should return the source untouched", n)
		}
	}
}

func TestApplyOrientationSwapsDimensions(t *testing.T) {
	src := gradientImage(4, 2)

	for _, n := range []int{5, 6, 7, 8} {
		got := applyOrientation(src, n).Bounds()
		if got.Dx() != 2 || got.Dy() != 4 {
			t.Errorf("applyOrientation(n=%d) bounds = %dx%d, want 2x4", n, got.Dx(), got.Dy())
		}
	}
	for _, n := range []int{2, 3, 4} {
		got := applyOrientation(src, n).Bounds()
		if got.Dx() != 4 || got.Dy() != 2 {
			t.Errorf("applyOrientation(n=%d) bounds = %dx%d, want 4x2", n, got.Dx(), got.Dy())
		}
	}
}

func TestJPEGOrientation(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{"big endian exif", jpegWithOrientation(t, 6, binary.BigEndian), 6},
		{"little endian exif", jpegWithOrientation(t, 8, binary.LittleEndian), 8},
		{"orientation 1", jpegWithOrientation(t, 1, binary.BigEndian), 1},
		{"no exif segment", encodeJPEG(t, quad()), 1},
		{"not a jpeg", []byte("plain text, not an image"), 1},
		{"truncated", []byte{0xFF, 0xD8}, 1},
		{"empty", nil, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jpegOrientation(tt.data); got != tt.want {
				t.Errorf("jpegOrientation() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDecodeAppliesJPEGOrientation(t *testing.T) {
	// A 4x2 image tagged "rotate 90" must come back out as 2x4.
	data := jpegWithOrientation(t, 6, binary.BigEndian)

	src, err := Decode("rotated.jpg", data)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if src.Width != 2 || src.Height != 4 {
		t.Errorf("decoded dimensions = %dx%d, want 2x4 (orientation applied)", src.Width, src.Height)
	}
}

func rgbaAt(img image.Image, x, y int) color.RGBA {
	r, g, b, a := img.At(x, y).RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func encodeJPEG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return buf.Bytes()
}

// jpegWithOrientation splices a minimal EXIF APP1 segment carrying just the
// orientation tag into a real 4x2 JPEG, right after the SOI marker.
func jpegWithOrientation(t *testing.T, orientation int, bo binary.ByteOrder) []byte {
	t.Helper()

	tiff := new(bytes.Buffer)
	if bo == binary.BigEndian {
		tiff.WriteString("MM")
	} else {
		tiff.WriteString("II")
	}
	write := func(v any) {
		if err := binary.Write(tiff, bo, v); err != nil {
			t.Fatalf("build exif: %v", err)
		}
	}
	write(uint16(42))
	write(uint32(8))              // IFD0 begins immediately after this header
	write(uint16(1))              // one entry
	write(uint16(orientationTag)) // tag
	write(uint16(3))              // type SHORT
	write(uint32(1))              // count
	write(uint16(orientation))    // value, packed into the first half of the field
	write(uint16(0))              // padding for the 4-byte value field
	write(uint32(0))              // no next IFD

	payload := append([]byte("Exif\x00\x00"), tiff.Bytes()...)
	segment := []byte{0xFF, 0xE1, 0, 0}
	binary.BigEndian.PutUint16(segment[2:], uint16(len(payload)+2))
	segment = append(segment, payload...)

	base := encodeJPEG(t, gradientImage(4, 2))
	out := make([]byte, 0, len(base)+len(segment))
	out = append(out, base[:2]...)
	out = append(out, segment...)
	out = append(out, base[2:]...)
	return out
}
