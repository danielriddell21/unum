package engine

import "fmt"

func Render(lang, format, source string, t *Theme) (out []byte, contentType string, err error) {
	var svg []byte
	var d2 *D2Diagram
	var mermaidSrc string
	switch lang {
	case "d2":
		d, err := RenderD2(source, t)
		if err != nil {
			return nil, "", err
		}
		d2, svg = d, d.SVG
	case "mermaid":
		s, err := RenderMermaid(source, t)
		if err != nil {
			return nil, "", err
		}
		svg = s
		mermaidSrc = source
	default:
		return nil, "", fmt.Errorf("unknown language %q", lang)
	}

	switch format {
	case "png":
		png, err := toPNG(d2, svg)
		if err != nil {
			return nil, "", err
		}
		return png, "image/png", nil
	case "drawio":
		out, err := drawio(d2, svg, mermaidSrc, t)
		if err != nil {
			return nil, "", err
		}
		return out, "application/xml", nil
	default:
		return svg, "image/svg+xml", nil
	}
}

func toPNG(d2 *D2Diagram, svg []byte) ([]byte, error) {
	if d2 != nil {
		return SVGToPNG(svg)
	}
	return MermaidSVGToPNG(svg)
}

func drawio(d2 *D2Diagram, svg []byte, mermaidSrc string, t *Theme) ([]byte, error) {
	if d2 != nil {
		return d2.Drawio()
	}
	return MermaidDrawio(mermaidSrc, svg, t)
}
