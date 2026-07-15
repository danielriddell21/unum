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

func newFlowGraph() *flowGraph {
	return &flowGraph{labels: map[string]string{}, seen: map[string]bool{}}
}

func (g *flowGraph) add(id string) {
	if id != "" && !g.seen[id] {
		g.seen[id] = true
		g.order = append(g.order, id)
	}
}

func (g *flowGraph) edge(src, dst, label string) {
	g.add(src)
	g.add(dst)
	g.edges = append(g.edges, mermaidEdge{Src: src, Dst: dst, Label: label})
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
			g.edge(n.From, n.To, n.Label)
		case *ast.Subgraph:
			g.walk(n.Statements)
		}
	}
}

func (g *flowGraph) d2() string {
	if len(g.order) == 0 {
		return ""
	}
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

func MermaidToD2(source string) (string, error) {
	diagram, err := mmparse.Parse(source)
	if err != nil {
		return "", fmt.Errorf("parse mermaid: %w", err)
	}
	switch n := diagram.(type) {
	case *ast.Flowchart:
		g := newFlowGraph()
		g.walk(n.Statements)
		return g.d2(), nil
	case *ast.SequenceDiagram:
		return sequenceD2(n), nil
	case *ast.StateDiagram:
		return stateD2(n), nil
	case *ast.ERDiagram:
		return erD2(n), nil
	default:
		// Chart and timeline types have no d2 graph equivalent.
		return "", nil
	}
}

func mermaidFlowchartD2(source string) (string, error) {
	fc, err := mmparse.ParseFlowchart(source)
	if err != nil {
		return "", fmt.Errorf("parse mermaid flowchart: %w", err)
	}
	g := newFlowGraph()
	g.walk(fc.Statements)
	return g.d2(), nil
}

type seqCollector struct {
	order  []string
	labels map[string]string
	seen   map[string]bool
	msgs   []mermaidEdge
}

func (c *seqCollector) add(id, alias string) {
	if id == "" {
		return
	}
	if !c.seen[id] {
		c.seen[id] = true
		c.order = append(c.order, id)
	}
	if alias != "" {
		c.labels[id] = alias
	}
}

func (c *seqCollector) walk(stmts []ast.SeqStmt) {
	for _, s := range stmts {
		switch n := s.(type) {
		case *ast.Participant:
			c.add(n.ID, n.Alias)
		case *ast.Message:
			c.add(n.From, "")
			c.add(n.To, "")
			c.msgs = append(c.msgs, mermaidEdge{Src: n.From, Dst: n.To, Label: n.Text})
		case *ast.Loop:
			c.walk(n.Statements)
		case *ast.Alt:
			for _, cond := range n.Conditions {
				c.walk(cond.Statements)
			}
		}
	}
}

func sequenceD2(sd *ast.SequenceDiagram) string {
	c := &seqCollector{labels: map[string]string{}, seen: map[string]bool{}}
	c.walk(sd.Statements)
	if len(c.order) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("shape: sequence_diagram\n")
	for _, id := range c.order {
		label := c.labels[id]
		if label == "" {
			label = id
		}
		fmt.Fprintf(&b, "%s: %q\n", d2ID(id), label)
	}
	for _, m := range c.msgs {
		if m.Label != "" {
			fmt.Fprintf(&b, "%s -> %s: %q\n", d2ID(m.Src), d2ID(m.Dst), m.Label)
		} else {
			fmt.Fprintf(&b, "%s -> %s\n", d2ID(m.Src), d2ID(m.Dst))
		}
	}
	return b.String()
}

func stateD2(sd *ast.StateDiagram) string {
	g := newFlowGraph()
	var walk func(stmts []ast.StateStmt)
	walk = func(stmts []ast.StateStmt) {
		for _, s := range stmts {
			switch n := s.(type) {
			case *ast.State:
				g.add(n.ID)
				if n.Description != "" {
					g.labels[n.ID] = n.Description
				}
				if len(n.Nested) > 0 {
					walk(n.Nested)
				}
			case *ast.Transition:
				g.edge(n.From, n.To, n.Label)
			case *ast.StartState:
				g.labels["start"] = "●"
				g.edge("start", n.To, "")
			case *ast.EndState:
				g.labels["end"] = "◉"
				g.edge(n.From, "end", "")
			}
		}
	}
	walk(sd.Statements)
	return g.d2()
}

func erD2(erd *ast.ERDiagram) string {
	g := newFlowGraph()
	for _, e := range erd.Entities {
		g.add(e.Name)
		if e.Alias != "" {
			g.labels[e.Name] = e.Alias
		}
	}
	for _, r := range erd.Relationships {
		g.edge(r.From, r.To, r.Label)
	}
	return g.d2()
}

func d2ID(id string) string {
	if d2IDSafe.MatchString(id) {
		return id
	}
	return fmt.Sprintf("%q", id)
}
