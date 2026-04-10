package parse

import (
	"fmt"

	diffnode "github.com/danielriddell21/unum/internal/diff/node"
	jsonnode "github.com/danielriddell21/unum/internal/json/node"
	jsonparse "github.com/danielriddell21/unum/internal/json/parse"
)

// JSON computes a semantic diff between two JSON byte slices.
// It walks both parse trees simultaneously and produces a DiffNode tree.
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

// compareNodes recursively compares two JSON nodes and returns a DiffNode.
// counts is [added, removed, modified].
func compareNodes(a, b *jsonnode.Node, path, key string, index int, counts *[3]int) *diffnode.DiffNode {
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

	case jsonnode.KindArray:
		dn.Kind = diffnode.Unchanged
		aLen := len(a.Children)
		bLen := len(b.Children)
		maxLen := aLen
		if bLen > maxLen {
			maxLen = bLen
		}
		for i := 0; i < maxLen; i++ {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			if i < aLen && i < bLen {
				dn.Children = append(dn.Children,
					compareNodes(a.Children[i], b.Children[i], childPath, "", i, counts))
			} else if i < bLen {
				counts[0]++
				dn.Children = append(dn.Children, &diffnode.DiffNode{
					Kind:     diffnode.Added,
					Path:     childPath,
					Index:    i,
					NewValue: nodeRepr(b.Children[i]),
				})
			} else {
				counts[1]++
				dn.Children = append(dn.Children, &diffnode.DiffNode{
					Kind:     diffnode.Removed,
					Path:     childPath,
					Index:    i,
					OldValue: nodeRepr(a.Children[i]),
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

// nodeRepr returns a compact display string for a JSON node.
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
