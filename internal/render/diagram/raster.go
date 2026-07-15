package diagram

import (
	"context"
	"fmt"
	"regexp"

	resvg "github.com/kanrichan/resvg-go"
	"oss.terrastruct.com/d2/d2renderers/d2fonts"
)

const d2PNGScale = 2

var (
	d2FontMono = regexp.MustCompile(`d2-\d+-font-mono`)
	d2FontAny  = regexp.MustCompile(`d2-\d+-font-\w+`)
)

func SVGToPNG(svg []byte) ([]byte, error) {
	w, h := SVGSize(svg)
	if w == 0 || h == 0 {
		w, h = 800, 600
	}
	// d2 references fonts by a hashed WOFF @font-face name resvg can't resolve;
	// rewrite them to the bundled TTF families loaded below.
	svg = d2FontMono.ReplaceAll(svg, []byte("Source Code Pro"))
	svg = d2FontAny.ReplaceAll(svg, []byte("Source Sans Pro"))

	ctx, err := resvg.NewContext(context.Background())
	if err != nil {
		return nil, fmt.Errorf("resvg init: %w", err)
	}
	defer ctx.Close() //nolint:errcheck // best-effort teardown
	r, err := ctx.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("resvg renderer: %w", err)
	}
	defer r.Close() //nolint:errcheck // best-effort teardown

	d2fonts.FontFaces.Range(func(_ d2fonts.Font, ttf []byte) bool {
		_ = r.LoadFontData(ttf)
		return true
	})
	png, err := r.RenderWithSize(svg, uint32(w*d2PNGScale), uint32(h*d2PNGScale))
	if err != nil {
		return nil, fmt.Errorf("resvg render: %w", err)
	}
	return png, nil
}
