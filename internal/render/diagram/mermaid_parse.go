package diagram

import (
	"fmt"
	"regexp"
	"strings"

	mmparse "github.com/sammcj/mermaid-check"
	"github.com/sammcj/mermaid-check/ast"
)

var d2IDSafe = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

type mermaidEdge struct{ Src, Dst, Label string }

type flowGraph struct {
	order  []string
	labels map[string]string
	seen   map[string]bool
	edges  []mermaidEdge
}

func (g *flowGraph) add(id string) {
	if id != "" && !g.seen[id] {
		g.seen[id] = true
		g.order = append(g.order, id)
	}
}

func (g *flowGraph) walk(stmts []ast.Statement) {
	for _, s := range stmts {
		switch n := s.(type) {
		case *ast.NodeDef:
			g.add(n.ID)
			if n.Label != "" {
				g.labels[n.ID] = n.Label
			}
		case *ast.Link:
			g.add(n.From)
			g.add(n.To)
			g.edges = append(g.edges, mermaidEdge{Src: n.From, Dst: n.To, Label: n.Label})
		case *ast.Subgraph:
			g.walk(n.Statements)
		}
	}
}

func (g *flowGraph) d2() string {
	var b strings.Builder
	for _, id := range g.order {
		label := g.labels[id]
		if label == "" {
			label = id
		}
		fmt.Fprintf(&b, "%s: %q\n", d2ID(id), label)
	}
	for _, e := range g.edges {
		if e.Label != "" {
			fmt.Fprintf(&b, "%s -> %s: %q\n", d2ID(e.Src), d2ID(e.Dst), e.Label)
		} else {
			fmt.Fprintf(&b, "%s -> %s\n", d2ID(e.Src), d2ID(e.Dst))
		}
	}
	return b.String()
}

// MermaidToD2 parses a mermaid flowchart natively (no browser) and re-emits it
// as equivalent d2 source. Returns an empty string for diagram types that are
// not node/edge flowcharts.
func MermaidToD2(source string) (string, error) {
	fc, err := mmparse.ParseFlowchart(source)
	if err != nil {
		return "", fmt.Errorf("parse mermaid flowchart: %w", err)
	}
	g := &flowGraph{labels: map[string]string{}, seen: map[string]bool{}}
	g.walk(fc.Statements)
	if len(g.order) == 0 {
		return "", nil
	}
	return g.d2(), nil
}

func d2ID(id string) string {
	if d2IDSafe.MatchString(id) {
		return id
	}
	return fmt.Sprintf("%q", id)
}
