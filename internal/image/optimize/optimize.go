package optimize

import (
	"bytes"
	"fmt"
	"image"
	"sync"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

type Source struct {
	Name   string
	Image  image.Image
	Format Format
	Width  int
	Height int
	Bytes  int
}

type Options struct {
	Quality   int
	Scale     float64
	MaxWidth  int
	MaxHeight int
	Format    Format
}

type Result struct {
	Data    []byte
	Format  Format
	Quality int
	Width   int
	Height  int
	Bytes   int
}

type Step struct {
	Quality int `json:"quality"`
	Bytes   int `json:"bytes"`
}

var DefaultLadder = []int{95, 90, 85, 80, 75, 65, 55, 40}

func Decode(name string, data []byte) (Source, error) {
	img, decoded, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return Source{}, fmt.Errorf("unsupported or corrupt image: %w", err)
	}

	format := formatFromDecoded(decoded)
	// Go's jpeg encoder drops EXIF, so bake the rotation into the pixels or a
	// tagged phone photo comes back out sideways.
	if format == FormatJPEG {
		img = applyOrientation(img, jpegOrientation(data))
	}

	b := img.Bounds()
	return Source{
		Name:   name,
		Image:  img,
		Format: format,
		Width:  b.Dx(),
		Height: b.Dy(),
		Bytes:  len(data),
	}, nil
}

func ResolveOutputFormat(explicit, src Format) Format {
	if explicit != FormatUnknown {
		return explicit
	}
	if src.CanEncode() {
		return src
	}
	return FormatPNG
}

func (o Options) prepare(src Source) (image.Image, Format, error) {
	format := ResolveOutputFormat(o.Format, src.Format)
	if !format.CanEncode() {
		return nil, FormatUnknown, fmt.Errorf("cannot write %s: output format must be jpeg, png, or gif", format)
	}
	w, h := TargetDimensions(src.Width, src.Height, o.Scale, o.MaxWidth, o.MaxHeight)
	return Resample(src.Image, w, h), format, nil
}

func Optimize(src Source, opts Options) (Result, error) {
	img, format, err := opts.prepare(src)
	if err != nil {
		return Result{}, err
	}

	quality := ClampQuality(opts.Quality)
	data, err := Encode(img, format, quality)
	if err != nil {
		return Result{}, err
	}
	return newResult(data, img, format, quality), nil
}

func Ladder(src Source, opts Options, qualities []int) ([]Step, error) {
	img, format, err := opts.prepare(src)
	if err != nil {
		return nil, err
	}

	steps := make([]Step, len(qualities))
	errs := make([]error, len(qualities))

	// Each rung is an independent encode of the same immutable image, so the
	// whole ladder costs about as much wall time as its slowest rung.
	var wg sync.WaitGroup
	for i, q := range qualities {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := Encode(img, format, q)
			if err != nil {
				errs[i] = err
				return
			}
			steps[i] = Step{Quality: ClampQuality(q), Bytes: len(data)}
		}()
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return steps, nil
}

func FitTo(src Source, opts Options, budget int) (Result, error) {
	img, format, err := opts.prepare(src)
	if err != nil {
		return Result{}, err
	}

	var best Result
	for lo, hi := 1, 100; lo <= hi; {
		mid := (lo + hi) / 2
		data, err := Encode(img, format, mid)
		if err != nil {
			return Result{}, err
		}
		if len(data) <= budget {
			best = newResult(data, img, format, mid)
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}

	// Nothing fit the budget — hand back the smallest we can make so the caller
	// can report the miss with a real file rather than nothing.
	if best.Data == nil {
		data, err := Encode(img, format, 1)
		if err != nil {
			return Result{}, err
		}
		best = newResult(data, img, format, 1)
	}
	return best, nil
}

func newResult(data []byte, img image.Image, format Format, quality int) Result {
	b := img.Bounds()
	return Result{
		Data:    data,
		Format:  format,
		Quality: quality,
		Width:   b.Dx(),
		Height:  b.Dy(),
		Bytes:   len(data),
	}
}
