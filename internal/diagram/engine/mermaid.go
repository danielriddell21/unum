package engine

import (
	"fmt"

	gm "github.com/zkrebbekx/go-mermaid"
	gmraster "github.com/zkrebbekx/go-mermaid/raster"
)

const mermaidPNGScale = 3.0

func RenderMermaid(source string, t *Theme) ([]byte, error) {
	var opts []gm.Option
	if t != nil {
		opts = append(opts, gm.WithCustomTheme("unum", t.goMermaidPalette()))
	}
	svg, err := gm.Render(source, opts...)
	if err != nil {
		return nil, fmt.Errorf("render mermaid: %w", err)
	}
	return svg, nil
}

func MermaidSVGToPNG(svg []byte) ([]byte, error) {
	png, err := gmraster.RasterizeSVG(svg, mermaidPNGScale)
	if err != nil {
		return nil, fmt.Errorf("rasterise mermaid png: %w", err)
	}
	return png, nil
}

func (t *Theme) goMermaidPalette() gm.Palette {
	return gm.Palette{
		// Fill nodes with the canvas colour (border-only, like the d2 render) so
		// the two engines look like one tool and labels keep full contrast.
		Background: t.Background,
		NodeFill:   t.Background,
		NodeStroke: t.Accent,
		Text:       t.Text,
		Edge:       t.Accent,
	}
}
