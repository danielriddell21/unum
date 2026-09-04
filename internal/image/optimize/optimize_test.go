package optimize

import (
	"bytes"
	"image"
	"testing"
)

func mustEncode(t *testing.T, img image.Image, format Format, quality int) []byte {
	t.Helper()
	data, err := Encode(img, format, quality)
	if err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	return data
}

func mustDecode(t *testing.T, name string, data []byte) Source {
	t.Helper()
	src, err := Decode(name, data)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return src
}

func TestDecode(t *testing.T) {
	img := gradientImage(40, 30)

	tests := []struct {
		name   string
		data   []byte
		format Format
	}{
		{"jpeg", mustEncode(t, img, FormatJPEG, 90), FormatJPEG},
		{"png", mustEncode(t, img, FormatPNG, 100), FormatPNG},
		{"gif", mustEncode(t, img, FormatGIF, 90), FormatGIF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, err := Decode("fixture"+tt.format.Ext(), tt.data)
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			if src.Format != tt.format {
				t.Errorf("Format = %v, want %v", src.Format, tt.format)
			}
			if src.Width != 40 || src.Height != 30 {
				t.Errorf("dimensions = %dx%d, want 40x30", src.Width, src.Height)
			}
			if src.Bytes != len(tt.data) {
				t.Errorf("Bytes = %d, want %d", src.Bytes, len(tt.data))
			}
			if src.Name != "fixture"+tt.format.Ext() {
				t.Errorf("Name = %q, want the name passed in", src.Name)
			}
		})
	}
}

func TestDecodeRejectsNonImages(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"text", []byte("this is not an image")},
		{"empty", nil},
		{"truncated png", []byte{0x89, 'P', 'N', 'G'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Decode("bad", tt.data); err == nil {
				t.Error("expected an error decoding a non-image")
			}
		})
	}
}

func TestOptimizeScalesAndReencodes(t *testing.T) {
	src := mustDecode(t, "in.jpg", mustEncode(t, gradientImage(200, 100), FormatJPEG, 95))

	got, err := Optimize(src, Options{Quality: 60, Scale: 0.5})
	if err != nil {
		t.Fatalf("Optimize: %v", err)
	}

	if got.Width != 100 || got.Height != 50 {
		t.Errorf("dimensions = %dx%d, want 100x50", got.Width, got.Height)
	}
	if got.Format != FormatJPEG {
		t.Errorf("Format = %v, want jpeg carried over from the source", got.Format)
	}
	if got.Quality != 60 {
		t.Errorf("Quality = %d, want 60", got.Quality)
	}
	if got.Bytes != len(got.Data) || got.Bytes == 0 {
		t.Errorf("Bytes = %d, does not match the %d bytes returned", got.Bytes, len(got.Data))
	}
	if got.Bytes >= src.Bytes {
		t.Errorf("optimized output is %d bytes, no smaller than the %d byte source", got.Bytes, src.Bytes)
	}
}

func TestOptimizeConvertsFormat(t *testing.T) {
	src := mustDecode(t, "in.png", mustEncode(t, gradientImage(64, 64), FormatPNG, 100))

	got, err := Optimize(src, Options{Quality: 80, Format: FormatJPEG})
	if err != nil {
		t.Fatalf("Optimize: %v", err)
	}

	if got.Format != FormatJPEG {
		t.Fatalf("Format = %v, want jpeg", got.Format)
	}
	if !bytes.HasPrefix(got.Data, []byte{0xFF, 0xD8, 0xFF}) {
		t.Error("output is not a jpeg")
	}
}

func TestOptimizeAppliesMaxWidth(t *testing.T) {
	src := mustDecode(t, "in.png", mustEncode(t, gradientImage(400, 200), FormatPNG, 100))

	got, err := Optimize(src, Options{Quality: 100, MaxWidth: 100})
	if err != nil {
		t.Fatalf("Optimize: %v", err)
	}

	if got.Width != 100 || got.Height != 50 {
		t.Errorf("dimensions = %dx%d, want 100x50 with the aspect ratio kept", got.Width, got.Height)
	}
}

func TestOptimizeRejectsUnwritableFormat(t *testing.T) {
	src := mustDecode(t, "in.png", mustEncode(t, gradientImage(16, 16), FormatPNG, 100))

	if _, err := Optimize(src, Options{Quality: 80, Format: FormatWebP}); err == nil {
		t.Error("expected an error asking for a format we cannot write")
	}
}

func TestLadder(t *testing.T) {
	src := mustDecode(t, "in.jpg", mustEncode(t, noiseImage(96, 96), FormatJPEG, 95))

	steps, err := Ladder(src, Options{}, DefaultLadder)
	if err != nil {
		t.Fatalf("Ladder: %v", err)
	}

	if len(steps) != len(DefaultLadder) {
		t.Fatalf("got %d rungs, want %d", len(steps), len(DefaultLadder))
	}
	for i, step := range steps {
		if step.Quality != DefaultLadder[i] {
			t.Errorf("rung %d quality = %d, want %d", i, step.Quality, DefaultLadder[i])
		}
		if step.Bytes <= 0 {
			t.Errorf("rung %d reported %d bytes", i, step.Bytes)
		}
	}

	// The ladder runs high quality to low, so sizes should trend downwards.
	if steps[0].Bytes <= steps[len(steps)-1].Bytes {
		t.Errorf("top rung (%d bytes) should be larger than the bottom rung (%d bytes)",
			steps[0].Bytes, steps[len(steps)-1].Bytes)
	}
}

func TestLadderRespectsScale(t *testing.T) {
	src := mustDecode(t, "in.jpg", mustEncode(t, noiseImage(128, 128), FormatJPEG, 95))

	full, err := Ladder(src, Options{}, []int{80})
	if err != nil {
		t.Fatalf("Ladder full: %v", err)
	}
	half, err := Ladder(src, Options{Scale: 0.5}, []int{80})
	if err != nil {
		t.Fatalf("Ladder half: %v", err)
	}

	if half[0].Bytes >= full[0].Bytes {
		t.Errorf("scaled ladder rung is %d bytes, not smaller than the full-size %d bytes",
			half[0].Bytes, full[0].Bytes)
	}
}

func TestLadderPropagatesEncodeErrors(t *testing.T) {
	src := mustDecode(t, "in.png", mustEncode(t, gradientImage(16, 16), FormatPNG, 100))

	if _, err := Ladder(src, Options{Format: FormatBMP}, DefaultLadder); err == nil {
		t.Error("expected an error for a format we cannot write")
	}
}

func TestFitToFindsTheBestQualityUnderBudget(t *testing.T) {
	src := mustDecode(t, "in.jpg", mustEncode(t, noiseImage(160, 160), FormatJPEG, 98))
	budget := src.Bytes / 3

	got, err := FitTo(src, Options{}, budget)
	if err != nil {
		t.Fatalf("FitTo: %v", err)
	}

	if got.Bytes > budget {
		t.Errorf("result is %d bytes, over the %d byte budget", got.Bytes, budget)
	}
	if got.Quality < 1 || got.Quality > 100 {
		t.Errorf("Quality = %d, want 1-100", got.Quality)
	}

	// One quality step up should break the budget, or we left savings unused.
	if got.Quality < 100 {
		bigger, err := Encode(Resample(src.Image, src.Width, src.Height), got.Format, got.Quality+1)
		if err != nil {
			t.Fatalf("encode one step up: %v", err)
		}
		if len(bigger) <= budget {
			t.Errorf("quality %d also fits the budget at %d bytes, so %d was not the best fit",
				got.Quality+1, len(bigger), got.Quality)
		}
	}
}

func TestFitToReturnsSmallestPossibleWhenBudgetUnreachable(t *testing.T) {
	src := mustDecode(t, "in.jpg", mustEncode(t, noiseImage(128, 128), FormatJPEG, 95))

	got, err := FitTo(src, Options{}, 10)
	if err != nil {
		t.Fatalf("FitTo: %v", err)
	}

	if got.Quality != 1 {
		t.Errorf("Quality = %d, want the floor of 1 when nothing fits", got.Quality)
	}
	if len(got.Data) == 0 {
		t.Error("expected a best-effort image even when the budget cannot be met")
	}
}

func TestFitToGeneratesADecodableImage(t *testing.T) {
	src := mustDecode(t, "in.jpg", mustEncode(t, noiseImage(96, 96), FormatJPEG, 95))

	got, err := FitTo(src, Options{}, src.Bytes/2)
	if err != nil {
		t.Fatalf("FitTo: %v", err)
	}

	if _, _, err := image.Decode(bytes.NewReader(got.Data)); err != nil {
		t.Fatalf("result does not decode: %v", err)
	}
}
