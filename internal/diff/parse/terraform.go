package parse

import (
	"encoding/json"
	"fmt"

	diffnode "github.com/danielriddell21/unum/internal/diff/node"
	jsonparse "github.com/danielriddell21/unum/internal/json/parse"
	"github.com/danielriddell21/unum/pkg/terraform"
)

func Terraform(data []byte) (*diffnode.Diff, error) {
	plan, err := terraform.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse terraform plan: %w", err)
	}

	root := &diffnode.DiffNode{
		Kind:  diffnode.Unchanged,
		Path:  ".",
		Index: -1,
	}

	var added, removed, modified int

	for _, rc := range plan.Changes {
		childPath := "." + rc.Address

		switch rc.Action {
		case terraform.ActionCreate:
			added++
			root.Children = append(root.Children, &diffnode.DiffNode{
				Kind:     diffnode.Added,
				Path:     childPath,
				Key:      rc.Address,
				Index:    -1,
				NewValue: fmt.Sprintf("(%s) %s", rc.Type, rc.Name),
			})

		case terraform.ActionDelete:
			removed++
			root.Children = append(root.Children, &diffnode.DiffNode{
				Kind:     diffnode.Removed,
				Path:     childPath,
				Key:      rc.Address,
				Index:    -1,
				OldValue: fmt.Sprintf("(%s) %s", rc.Type, rc.Name),
			})

		case terraform.ActionUpdate, terraform.ActionReplace:
			modified++
			var counts [3]int
			resourceNode := diffTFResource(rc, childPath, &counts)
			root.Children = append(root.Children, resourceNode)

		default: // no-op, read — skip unchanged resources
		}
	}

	return &diffnode.Diff{
		Format:   diffnode.FormatTerraform,
		Root:     root,
		Added:    added,
		Removed:  removed,
		Modified: modified,
	}, nil
}

func diffTFResource(rc terraform.ResourceChange, path string, counts *[3]int) *diffnode.DiffNode {
	fallback := func() *diffnode.DiffNode {
		counts[2]++
		return &diffnode.DiffNode{
			Kind:     diffnode.Modified,
			Path:     path,
			Key:      rc.Address,
			Index:    -1,
			OldValue: fmt.Sprintf("(%s) before", rc.Type),
			NewValue: fmt.Sprintf("(%s) after", rc.Type),
		}
	}

	beforeBytes, err := json.Marshal(rc.Before)
	if err != nil {
		return fallback()
	}
	afterBytes, err := json.Marshal(rc.After)
	if err != nil {
		return fallback()
	}

	beforeNode, err := jsonparse.Parse(beforeBytes)
	if err != nil {
		return fallback()
	}
	afterNode, err := jsonparse.Parse(afterBytes)
	if err != nil {
		return fallback()
	}

	inner := compareNodes(beforeNode, afterNode, path, rc.Address, -1, counts)
	inner.Path = path
	inner.Key = rc.Address
	return inner
}
