package diagram

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed assets/mermaid.min.js
var mermaidJS string

func (b *Browser) RenderMermaid(source string) ([]byte, error) {
	html := `<!doctype html><html><head><meta charset="utf-8"><script>` +
		mermaidJS + `</script></head><body><div id="container"></div></body></html>`
	page, err := b.page(html)
	if err != nil {
		return nil, err
	}
	defer page.Close() //nolint:errcheck // page discarded with the browser

	if _, err := page.Eval(`() => mermaid.initialize({ startOnLoad: false })`); err != nil {
		return nil, fmt.Errorf("init mermaid: %w", err)
	}
	res, err := page.Eval(`async (src) => {
		const { svg } = await mermaid.render('unum-diagram', src);
		return svg;
	}`, source)
	if err != nil {
		return nil, fmt.Errorf("render mermaid: %w", err)
	}
	svg := res.Value.Str()
	if !strings.Contains(svg, "<svg") {
		return nil, fmt.Errorf("mermaid produced no svg output")
	}
	return []byte(svg), nil
}
