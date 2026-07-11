package diagram

import "fmt"

func NeedsBrowser(lang, format string) bool {
	return lang == "d2" && format == "png"
}

func Render(lang, format, source string, b *Browser, t *Theme) (out []byte, contentType string, err error) {
	var svg []byte
	var d2 *D2Diagram
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
	default:
		return nil, "", fmt.Errorf("unknown language %q", lang)
	}

	switch format {
	case "png":
		png, err := toPNG(svg, d2, b)
		if err != nil {
			return nil, "", err
		}
		return png, "image/png", nil
	case "drawio":
		out, err := drawio(d2, svg)
		if err != nil {
			return nil, "", err
		}
		return out, "application/xml", nil
	default:
		return svg, "image/svg+xml", nil
	}
}

func toPNG(svg []byte, d2 *D2Diagram, b *Browser) ([]byte, error) {
	if d2 != nil {
		if b == nil {
			return nil, fmt.Errorf("d2 png output needs a browser")
		}
		return b.SVGToPNG(svg)
	}
	return MermaidSVGToPNG(svg)
}

func drawio(d2 *D2Diagram, svg []byte) ([]byte, error) {
	if d2 != nil {
		return d2.Drawio()
	}
	return DrawioFromSVG(svg)
}
