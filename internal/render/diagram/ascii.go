package diagram

import (
	"context"
	"fmt"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2ascii"
	"oss.terrastruct.com/d2/d2renderers/d2ascii/charset"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"oss.terrastruct.com/util-go/go2"
)

func RenderD2ASCII(source string) (string, error) {
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return "", fmt.Errorf("text ruler: %w", err)
	}
	layoutResolver := func(string) (d2graph.LayoutGraph, error) {
		return d2elklayout.DefaultLayout, nil
	}
	renderOpts := &d2svg.RenderOpts{Pad: go2.Pointer(int64(0))}
	compileOpts := &d2lib.CompileOptions{LayoutResolver: layoutResolver, Ruler: ruler}

	ctx := log.WithDefault(context.Background())
	target, _, err := d2lib.Compile(ctx, source, compileOpts, renderOpts)
	if err != nil {
		return "", fmt.Errorf("compile d2: %w", err)
	}
	out, err := d2ascii.NewASCIIartist().Render(ctx, target, &d2ascii.RenderOpts{Charset: charset.Unicode})
	if err != nil {
		return "", fmt.Errorf("render d2 ascii: %w", err)
	}
	return string(out), nil
}
