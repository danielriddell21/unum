package parse

import (
	"fmt"

	"gopkg.in/yaml.v3"

	diffnode "github.com/danielriddell21/unum/internal/diff/node"
)

// YAML computes a semantic diff between two YAML byte slices.
// It walks both yaml.Node trees simultaneously and produces a DiffNode tree.
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
	root := compareYAMLNodes(nodeA, nodeB, ".", "", -1, &counts)

	d := &diffnode.Diff{
		Format:   diffnode.FormatYAML,
		Root:     root,
		Added:    counts[0],
		Removed:  counts[1],
		Modified: counts[2],
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

func compareYAMLNodes(a, b *yaml.Node, path, key string, index int, counts *[3]int) *diffnode.DiffNode { //nolint:cyclop,gocognit,dupl // recursive tree comparison; branching on node kind/sequence bounds is inherent to the diff algorithm
	// Resolve aliases before comparing
	a = resolveAlias(a)
	b = resolveAlias(b)

	dn := &diffnode.DiffNode{Path: path, Key: key, Index: index}

	if a.Kind != b.Kind {
		dn.Kind = diffnode.Modified
		dn.OldValue = yamlNodeRepr(a)
		dn.NewValue = yamlNodeRepr(b)
		counts[2]++
		return dn
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
				dn.Children = append(dn.Children, compareYAMLNodes(aVal, bVal, childPath, k, -1, counts))
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
		aLen := len(a.Content)
		bLen := len(b.Content)
		maxLen := aLen
		if bLen > maxLen {
			maxLen = bLen
		}
		for i := 0; i < maxLen; i++ {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			switch {
			case i < aLen && i < bLen:
				dn.Children = append(dn.Children,
					compareYAMLNodes(a.Content[i], b.Content[i], childPath, "", i, counts))
			case i < bLen:
				counts[0]++
				dn.Children = append(dn.Children, &diffnode.DiffNode{
					Kind:     diffnode.Added,
					Path:     childPath,
					Index:    i,
					NewValue: yamlNodeRepr(b.Content[i]),
				})
			default:
				counts[1]++
				dn.Children = append(dn.Children, &diffnode.DiffNode{
					Kind:     diffnode.Removed,
					Path:     childPath,
					Index:    i,
					OldValue: yamlNodeRepr(a.Content[i]),
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

	return dn
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
