package diagram

import (
	"context"
	"fmt"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2target"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"oss.terrastruct.com/util-go/go2"
)

type D2Diagram struct {
	SVG    []byte
	target *d2target.Diagram
}

func (d *D2Diagram) Drawio() ([]byte, error) {
	return drawioFromD2(d.target)
}

func RenderD2(source string, t *Theme) (*D2Diagram, error) {
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, fmt.Errorf("text ruler: %w", err)
	}
	layoutResolver := func(string) (d2graph.LayoutGraph, error) {
		return d2dagrelayout.DefaultLayout, nil
	}
	renderOpts := &d2svg.RenderOpts{Pad: go2.Pointer(int64(8))}
	if t != nil {
		renderOpts.ThemeID = go2.Pointer(t.d2Base())
		renderOpts.ThemeOverrides = t.d2Overrides()
	}
	compileOpts := &d2lib.CompileOptions{LayoutResolver: layoutResolver, Ruler: ruler}

	ctx := log.WithDefault(context.Background())
	target, _, err := d2lib.Compile(ctx, source, compileOpts, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("compile d2: %w", err)
	}
	svg, err := d2svg.Render(target, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("render d2 svg: %w", err)
	}
	return &D2Diagram{SVG: svg, target: target}, nil
}
