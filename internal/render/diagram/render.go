package diagram

import "fmt"

func NeedsBrowser(lang, format string) bool {
	return lang == "mermaid" || format == "png"
}

func Render(lang, format, source string, b *Browser) (out []byte, contentType string, err error) {
	var svg []byte
	var d2 *D2Diagram
	switch lang {
	case "d2":
		d, err := RenderD2(source)
		if err != nil {
			return nil, "", err
		}
		d2, svg = d, d.SVG
	case "mermaid":
		if b == nil {
			return nil, "", fmt.Errorf("mermaid rendering needs a browser")
		}
		s, err := b.RenderMermaid(source)
		if err != nil {
			return nil, "", err
		}
		svg = s
	default:
		return nil, "", fmt.Errorf("unknown language %q", lang)
	}

	switch format {
	case "png":
		if b == nil {
			return nil, "", fmt.Errorf("png output needs a browser")
		}
		png, err := b.SVGToPNG(svg)
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

func drawio(d2 *D2Diagram, svg []byte) ([]byte, error) {
	if d2 != nil {
		return d2.Drawio()
	}
	return DrawioFromSVG(svg)
}
