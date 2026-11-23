package parse

import (
	"crypto/sha256"
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"

	diffnode "github.com/danielriddell21/unum/internal/diff/node"
)

const (
	maxYAMLNodes = 1_000_000
	maxYAMLDepth = 10_000
)

func YAML(a, b []byte) (*diffnode.Diff, error) {
	nodeA, err := parseYAMLDoc(a)
	if err != nil {
		return nil, fmt.Errorf("parse file A: %w", err)
	}
	nodeB, err := parseYAMLDoc(b)
	if err != nil {
		return nil, fmt.Errorf("parse file B: %w", err)
	}

	var counts [3]int // [added, removed, modified]
	st := &yamlState{keys: map[*yaml.Node]string{}, budget: maxYAMLNodes}
	root, err := compareYAMLNodes(nodeA, nodeB, st, ".", "", -1, &counts)
	if err != nil {
		return nil, err
	}

	d := &diffnode.Diff{
		Format:      diffnode.FormatYAML,
		Root:        root,
		Added:       counts[0],
		Removed:     counts[1],
		Modified:    counts[2],
		Approximate: st.approx,
	}
	if td, err := Text(a, b, 3); err == nil {
		d.Hunks = td.Hunks
	}
	return d, nil
}

func parseYAMLDoc(data []byte) (*yaml.Node, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("yaml unmarshal: %w", err)
	}
	// yaml.Unmarshal returns a DocumentNode; unwrap to content
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		return doc.Content[0], nil
	}
	return &doc, nil
}

func compareYAMLNodes(a, b *yaml.Node, st *yamlState, path, key string, index int, counts *[3]int) (*diffnode.DiffNode, error) { //nolint:cyclop,gocognit,dupl // NOSONAR: recursive tree comparison; branching on node kind/sequence bounds is inherent to the diff algorithm
	// only nodes reached through an alias spend budget: a plain document is bounded by its own size
	if a.Kind == yaml.AliasNode || b.Kind == yaml.AliasNode {
		st.inAlias++
		defer func() { st.inAlias-- }()
	}
	if st.inAlias > 0 {
		if st.budget--; st.budget < 0 {
			return nil, fmt.Errorf("yaml alias expansion exceeds %d nodes", maxYAMLNodes)
		}
	}
	if st.depth++; st.depth > maxYAMLDepth {
		return nil, fmt.Errorf("yaml alias expansion exceeds depth %d (recursive alias?)", maxYAMLDepth)
	}
	defer func() { st.depth-- }()

	// Resolve aliases before comparing
	a = resolveAlias(a)
	b = resolveAlias(b)

	dn := &diffnode.DiffNode{Path: path, Key: key, Index: index}

	if a.Kind != b.Kind {
		dn.Kind = diffnode.Modified
		dn.OldValue = yamlNodeRepr(a)
		dn.NewValue = yamlNodeRepr(b)
		counts[2]++
		return dn, nil
	}

	switch a.Kind {
	case yaml.MappingNode:
		dn.Kind = diffnode.Unchanged
		aMap := make(map[string]*yaml.Node, len(a.Content)/2)
		for i := 0; i+1 < len(a.Content); i += 2 {
			aMap[a.Content[i].Value] = a.Content[i+1]
		}

		seen := make(map[string]bool, len(b.Content)/2)
		for i := 0; i+1 < len(b.Content); i += 2 {
			k := b.Content[i].Value
			bVal := b.Content[i+1]
			seen[k] = true
			childPath := objPath(path, k)
			if aVal, ok := aMap[k]; ok {
				child, err := compareYAMLNodes(aVal, bVal, st, childPath, k, -1, counts)
				if err != nil {
					return nil, err
				}
				dn.Children = append(dn.Children, child)
			} else {
				counts[0]++
				dn.Children = append(dn.Children, &diffnode.DiffNode{
					Kind:     diffnode.Added,
					Path:     childPath,
					Key:      k,
					Index:    -1,
					NewValue: yamlNodeRepr(bVal),
				})
			}
		}
		for i := 0; i+1 < len(a.Content); i += 2 {
			k := a.Content[i].Value
			if !seen[k] {
				counts[1]++
				childPath := objPath(path, k)
				dn.Children = append(dn.Children, &diffnode.DiffNode{
					Kind:     diffnode.Removed,
					Path:     childPath,
					Key:      k,
					Index:    -1,
					OldValue: yamlNodeRepr(a.Content[i+1]),
				})
			}
		}

	case yaml.SequenceNode: //nolint:dupl // mirrors json.go KindArray case; parallel parsers share structure but not types
		dn.Kind = diffnode.Unchanged
		for _, op := range alignByKey(st.nodeKeys(a.Content), st.nodeKeys(b.Content), &st.approx) {
			switch {
			case op.A >= 0 && op.B >= 0:
				childPath := fmt.Sprintf("%s[%d]", path, op.B)
				child, err := compareYAMLNodes(a.Content[op.A], b.Content[op.B], st, childPath, "", op.B, counts)
				if err != nil {
					return nil, err
				}
				dn.Children = append(dn.Children, child)
			case op.B >= 0:
				counts[0]++
				dn.Children = append(dn.Children, &diffnode.DiffNode{
					Kind:     diffnode.Added,
					Path:     fmt.Sprintf("%s[%d]", path, op.B),
					Index:    op.B,
					NewValue: yamlNodeRepr(b.Content[op.B]),
				})
			default:
				counts[1]++
				dn.Children = append(dn.Children, &diffnode.DiffNode{
					Kind:     diffnode.Removed,
					Path:     fmt.Sprintf("%s[%d]", path, op.A),
					Index:    op.A,
					OldValue: yamlNodeRepr(a.Content[op.A]),
				})
			}
		}

	default: // scalar
		if a.Value == b.Value {
			dn.Kind = diffnode.Unchanged
		} else {
			dn.Kind = diffnode.Modified
			dn.OldValue = yamlScalarDisplay(a)
			dn.NewValue = yamlScalarDisplay(b)
			counts[2]++
		}
	}

	return dn, nil
}

func resolveAlias(n *yaml.Node) *yaml.Node {
	if n != nil && n.Kind == yaml.AliasNode && n.Alias != nil {
		return n.Alias
	}
	return n
}

func yamlNodeRepr(n *yaml.Node) string {
	n = resolveAlias(n)
	switch n.Kind {
	case yaml.MappingNode:
		if len(n.Content) == 0 {
			return "{}"
		}
		return fmt.Sprintf("{%d keys}", len(n.Content)/2)
	case yaml.SequenceNode:
		if len(n.Content) == 0 {
			return "[]"
		}
		return fmt.Sprintf("[%d items]", len(n.Content))
	default:
		return yamlScalarDisplay(n)
	}
}

func yamlScalarDisplay(n *yaml.Node) string {
	switch n.Tag {
	case "!!str":
		return fmt.Sprintf("%q", n.Value)
	case "!!null":
		return "null"
	default:
		return n.Value
	}
}

type yamlState struct {
	keys    map[*yaml.Node]string
	budget  int
	depth   int
	approx  bool
	inAlias int
}

func (st *yamlState) nodeKeys(nodes []*yaml.Node) []string {
	out := make([]string, len(nodes))
	for i, n := range nodes {
		out[i] = st.key(n)
	}
	return out
}

func (st *yamlState) key(n *yaml.Node) string {
	n = resolveAlias(n)
	if n.Kind != yaml.MappingNode && n.Kind != yaml.SequenceNode {
		return n.Tag + ":" + n.Value
	}
	if v, ok := st.keys[n]; ok {
		return v
	}

	// placeholder so an alias that points back at an ancestor terminates
	st.keys[n] = "cycle"

	parts := make([]string, 0, len(n.Content))
	prefix := "["
	if n.Kind == yaml.MappingNode {
		prefix = "{"
		for i := 0; i+1 < len(n.Content); i += 2 {
			parts = append(parts, lengthPrefixed(n.Content[i].Value)+st.key(n.Content[i+1]))
		}
		slices.Sort(parts)
	} else {
		for _, c := range n.Content {
			parts = append(parts, st.key(c))
		}
	}

	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(lengthPrefixed(p)))
	}
	v := prefix + string(h.Sum(nil))
	st.keys[n] = v
	return v
}
