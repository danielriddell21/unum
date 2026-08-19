package parse

import (
	"fmt"
	"slices"
	"strings"

	diffnode "github.com/danielriddell21/unum/internal/diff/node"
	jsonnode "github.com/danielriddell21/unum/internal/json/node"
	jsonparse "github.com/danielriddell21/unum/internal/json/parse"
)

func JSON(a, b []byte) (*diffnode.Diff, error) {
	rootA, err := jsonparse.Parse(a)
	if err != nil {
		return nil, fmt.Errorf("parse file A: %w", err)
	}
	rootB, err := jsonparse.Parse(b)
	if err != nil {
		return nil, fmt.Errorf("parse file B: %w", err)
	}

	var counts [3]int // [added, removed, modified]
	root := compareNodes(rootA, rootB, ".", "", -1, &counts)

	d := &diffnode.Diff{
		Format:   diffnode.FormatJSON,
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

func compareNodes(a, b *jsonnode.Node, path, key string, index int, counts *[3]int) *diffnode.DiffNode { //nolint:cyclop,gocognit,dupl // NOSONAR: recursive tree comparison; branching on node kind/array bounds is inherent to the diff algorithm
	dn := &diffnode.DiffNode{Path: path, Key: key, Index: index}

	if a.Kind != b.Kind {
		dn.Kind = diffnode.Modified
		dn.OldValue = nodeRepr(a)
		dn.NewValue = nodeRepr(b)
		counts[2]++
		return dn
	}

	switch a.Kind {
	case jsonnode.KindObject:
		dn.Kind = diffnode.Unchanged
		aMap := make(map[string]*jsonnode.Node, len(a.Children))
		for _, c := range a.Children {
			aMap[c.Key] = c
		}
		seen := make(map[string]bool, len(b.Children))

		// Process b's children in order (preserved + added)
		for _, bc := range b.Children {
			seen[bc.Key] = true
			childPath := objPath(path, bc.Key)
			if ac, ok := aMap[bc.Key]; ok {
				dn.Children = append(dn.Children, compareNodes(ac, bc, childPath, bc.Key, -1, counts))
			} else {
				counts[0]++
				dn.Children = append(dn.Children, &diffnode.DiffNode{
					Kind:     diffnode.Added,
					Path:     childPath,
					Key:      bc.Key,
					Index:    -1,
					NewValue: nodeRepr(bc),
				})
			}
		}
		// Append keys removed from a (in a's original order)
		for _, ac := range a.Children {
			if !seen[ac.Key] {
				counts[1]++
				childPath := objPath(path, ac.Key)
				dn.Children = append(dn.Children, &diffnode.DiffNode{
					Kind:     diffnode.Removed,
					Path:     childPath,
					Key:      ac.Key,
					Index:    -1,
					OldValue: nodeRepr(ac),
				})
			}
		}

	case jsonnode.KindArray: //nolint:dupl // mirrors yaml.go SequenceNode case; parallel parsers share structure but not types
		dn.Kind = diffnode.Unchanged
		for _, op := range alignByKey(jsonNodeKeys(a.Children), jsonNodeKeys(b.Children)) {
			switch {
			case op.A >= 0 && op.B >= 0:
				childPath := fmt.Sprintf("%s[%d]", path, op.B)
				dn.Children = append(dn.Children,
					compareNodes(a.Children[op.A], b.Children[op.B], childPath, "", op.B, counts))
			case op.B >= 0:
				counts[0]++
				dn.Children = append(dn.Children, &diffnode.DiffNode{
					Kind:     diffnode.Added,
					Path:     fmt.Sprintf("%s[%d]", path, op.B),
					Index:    op.B,
					NewValue: nodeRepr(b.Children[op.B]),
				})
			default:
				counts[1]++
				dn.Children = append(dn.Children, &diffnode.DiffNode{
					Kind:     diffnode.Removed,
					Path:     fmt.Sprintf("%s[%d]", path, op.A),
					Index:    op.A,
					OldValue: nodeRepr(a.Children[op.A]),
				})
			}
		}

	default: // leaf: bool, number, string, null
		if a.Raw == b.Raw {
			dn.Kind = diffnode.Unchanged
		} else {
			dn.Kind = diffnode.Modified
			dn.OldValue = a.Raw
			dn.NewValue = b.Raw
			counts[2]++
		}
	}

	return dn
}

func nodeRepr(n *jsonnode.Node) string {
	switch n.Kind {
	case jsonnode.KindObject:
		if len(n.Children) == 0 {
			return "{}"
		}
		return fmt.Sprintf("{%d keys}", len(n.Children))
	case jsonnode.KindArray:
		if len(n.Children) == 0 {
			return "[]"
		}
		return fmt.Sprintf("[%d items]", len(n.Children))
	default:
		return n.Raw
	}
}

func jsonNodeKeys(nodes []*jsonnode.Node) []string {
	keys := make([]string, len(nodes))
	for i, n := range nodes {
		keys[i] = jsonNodeKey(n)
	}
	return keys
}

func jsonNodeKey(n *jsonnode.Node) string {
	switch n.Kind {
	case jsonnode.KindObject:
		parts := make([]string, 0, len(n.Children))
		for _, c := range n.Children {
			parts = append(parts, c.Key+":"+jsonNodeKey(c))
		}
		slices.Sort(parts)
		return "{" + strings.Join(parts, ",") + "}"
	case jsonnode.KindArray:
		parts := make([]string, 0, len(n.Children))
		for _, c := range n.Children {
			parts = append(parts, jsonNodeKey(c))
		}
		return "[" + strings.Join(parts, ",") + "]"
	default:
		return n.Raw
	}
}
