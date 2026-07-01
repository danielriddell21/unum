package parse

import (
	"encoding/json"
	"fmt"

	diffnode "github.com/danielriddell21/unum/internal/diff/node"
	jsonparse "github.com/danielriddell21/unum/internal/json/parse"
)

func Terraform(data []byte) (*diffnode.Diff, error) {
	var plan tfPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("parse terraform plan: %w", err)
	}

	root := &diffnode.DiffNode{
		Kind:  diffnode.Unchanged,
		Path:  ".",
		Index: -1,
	}

	var added, removed, modified int

	for _, rc := range plan.ResourceChanges {
		action := primaryAction(rc.Change.Actions)
		childPath := "." + rc.Address

		switch action {
		case "create":
			added++
			root.Children = append(root.Children, &diffnode.DiffNode{
				Kind:     diffnode.Added,
				Path:     childPath,
				Key:      rc.Address,
				Index:    -1,
				NewValue: fmt.Sprintf("(%s) %s", rc.Type, rc.Name),
			})

		case "delete":
			removed++
			root.Children = append(root.Children, &diffnode.DiffNode{
				Kind:     diffnode.Removed,
				Path:     childPath,
				Key:      rc.Address,
				Index:    -1,
				OldValue: fmt.Sprintf("(%s) %s", rc.Type, rc.Name),
			})

		case "update":
			var counts [3]int
			resourceNode, err := diffTFChange(rc, childPath, &counts)
			if err != nil {
				// Fall back to treating the whole resource as modified
				counts[2]++
				resourceNode = &diffnode.DiffNode{
					Kind:     diffnode.Modified,
					Path:     childPath,
					Key:      rc.Address,
					Index:    -1,
					OldValue: fmt.Sprintf("(%s) before", rc.Type),
					NewValue: fmt.Sprintf("(%s) after", rc.Type),
				}
			}
			added += counts[0]
			removed += counts[1]
			modified += counts[2]
			root.Children = append(root.Children, resourceNode)

		default: // no-op, read
			// skip unchanged resources
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

func diffTFChange(rc tfResourceChange, path string, counts *[3]int) (*diffnode.DiffNode, error) {
	beforeBytes, err := json.Marshal(rc.Change.Before)
	if err != nil {
		return nil, fmt.Errorf("marshal before: %w", err)
	}
	afterBytes, err := json.Marshal(rc.Change.After)
	if err != nil {
		return nil, fmt.Errorf("marshal after: %w", err)
	}

	beforeNode, err := jsonparse.Parse(beforeBytes)
	if err != nil {
		return nil, fmt.Errorf("parse before: %w", err)
	}
	afterNode, err := jsonparse.Parse(afterBytes)
	if err != nil {
		return nil, fmt.Errorf("parse after: %w", err)
	}

	inner := compareNodes(beforeNode, afterNode, path, rc.Address, -1, counts)
	inner.Path = path
	inner.Key = rc.Address
	return inner, nil
}

func primaryAction(actions []string) string {
	for _, a := range actions {
		switch a {
		case "create", "delete", "update":
			return a
		}
	}
	if len(actions) > 0 {
		return actions[0]
	}
	return "no-op"
}

type tfPlan struct {
	ResourceChanges []tfResourceChange `json:"resource_changes"`
}

type tfResourceChange struct {
	Address string   `json:"address"`
	Type    string   `json:"type"`
	Name    string   `json:"name"`
	Change  tfChange `json:"change"`
}

type tfChange struct {
	Actions []string        `json:"actions"`
	Before  json.RawMessage `json:"before"`
	After   json.RawMessage `json:"after"`
}
