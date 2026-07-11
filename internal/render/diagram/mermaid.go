package diagram

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

//go:embed assets/mermaid.min.js
var mermaidJS string

var d2IDSafe = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

func (b *Browser) RenderMermaid(source string, t *Theme) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	html := `<!doctype html><html><head><meta charset="utf-8"><script>` +
		mermaidJS + `</script></head><body><div id="container"></div></body></html>`
	page, err := b.page(html)
	if err != nil {
		return nil, err
	}
	defer page.Close() //nolint:errcheck // page discarded with the browser

	config := map[string]any{"startOnLoad": false}
	if t != nil {
		config = t.mermaidConfig()
	}
	if _, err := page.Eval(`(cfg) => mermaid.initialize(cfg)`, config); err != nil {
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

type mermaidGraph struct {
	Nodes []struct{ ID, Label string }       `json:"nodes"`
	Edges []struct{ Src, Dst, Label string } `json:"edges"`
}

// MermaidToD2 renders a mermaid flowchart, reads its logical graph out of the
// DOM, and re-emits it as equivalent d2 source. Returns an empty string for
// diagram types that are not node/edge flowcharts.
func (b *Browser) MermaidToD2(source string) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	html := `<!doctype html><html><head><meta charset="utf-8"><script>` +
		mermaidJS + `</script></head><body><div id="container"></div></body></html>`
	page, err := b.page(html)
	if err != nil {
		return "", err
	}
	defer page.Close() //nolint:errcheck // page discarded with the browser

	if _, err := page.Eval(`() => mermaid.initialize({ startOnLoad: false })`); err != nil {
		return "", fmt.Errorf("init mermaid: %w", err)
	}
	res, err := page.Eval(`async (src) => {
		const { svg } = await mermaid.render('unum-diagram', src);
		const c = document.getElementById('container');
		c.innerHTML = svg;
		const nodes = [...c.querySelectorAll('.node')].map(n => {
			const m = (n.id || '').match(/flowchart-(.+)-\d+$/);
			return { ID: m ? m[1] : n.id, Label: (n.textContent || '').trim() };
		});
		const labels = [...c.querySelectorAll('.edgeLabels .edgeLabel')]
			.filter(e => !e.querySelector('.edgeLabel'))
			.map(e => (e.textContent || '').trim());
		const edges = [...c.querySelectorAll('[id*="-L_"]')].map(p => {
			const m = (p.id || '').match(/-L_(.+)_(.+)_\d+$/);
			return m ? { Src: m[1], Dst: m[2] } : null;
		}).filter(Boolean);
		edges.forEach((e, i) => { e.Label = labels[i] || ''; });
		return JSON.stringify({ nodes, edges });
	}`, source)
	if err != nil {
		return "", fmt.Errorf("extract mermaid graph: %w", err)
	}

	var g mermaidGraph
	if err := json.Unmarshal([]byte(res.Value.Str()), &g); err != nil {
		return "", fmt.Errorf("decode mermaid graph: %w", err)
	}
	if len(g.Nodes) == 0 || len(g.Edges) == 0 {
		return "", nil // not a flowchart we can convert
	}
	return graphToD2(g), nil
}

func graphToD2(g mermaidGraph) string {
	var b strings.Builder
	for _, n := range g.Nodes {
		fmt.Fprintf(&b, "%s: %q\n", d2ID(n.ID), n.Label)
	}
	for _, e := range g.Edges {
		if e.Label != "" {
			fmt.Fprintf(&b, "%s -> %s: %q\n", d2ID(e.Src), d2ID(e.Dst), e.Label)
		} else {
			fmt.Fprintf(&b, "%s -> %s\n", d2ID(e.Src), d2ID(e.Dst))
		}
	}
	return b.String()
}

func d2ID(id string) string {
	if d2IDSafe.MatchString(id) {
		return id
	}
	return fmt.Sprintf("%q", id)
}
