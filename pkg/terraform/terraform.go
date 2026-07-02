// Package terraform parses "terraform show -json" plan output and renders it as
// human-readable, terraform plan-style text.
//
// It is the shared engine behind unum's diff command and the tf-plan-summary
// GitHub Action. Decode a plan once with [Parse]; then count changes with
// [Plan.AddCount], [Plan.ChangeCount], [Plan.DestroyCount] and
// [Plan.ReplaceCount], list the changed resources with [Plan.Changed], or
// render the diff body with [Plan.RenderDiff].
//
// The package emits plain, uncoloured text and imports only the standard
// library: a caller adds any heading, footer and colour it wants.
package terraform

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// Action is the primary change Terraform will apply to a resource. The zero
// value is [ActionNoOp].
type Action int

const (
	// ActionNoOp means the resource is unchanged.
	ActionNoOp Action = iota
	// ActionCreate means the resource will be created.
	ActionCreate
	// ActionRead means the resource will be read during apply.
	ActionRead
	// ActionUpdate means the resource will be updated in place.
	ActionUpdate
	// ActionDelete means the resource will be destroyed.
	ActionDelete
	// ActionReplace means the resource will be destroyed and recreated.
	ActionReplace
)

// String returns the lower-case name of the action, such as "create" or
// "replace".
func (a Action) String() string {
	switch a {
	case ActionCreate:
		return "create"
	case ActionRead:
		return "read"
	case ActionUpdate:
		return "update"
	case ActionDelete:
		return "delete"
	case ActionReplace:
		return "replace"
	default:
		return "no-op"
	}
}

func (a Action) phrase() string {
	switch a {
	case ActionCreate:
		return "will be created"
	case ActionRead:
		return "will be read during apply"
	case ActionUpdate:
		return "will be updated in-place"
	case ActionDelete:
		return "will be destroyed"
	case ActionReplace:
		return "must be replaced"
	default:
		return "will not change"
	}
}

// ResourceChange is a single resource's planned change, as found in the
// resource_changes array of a plan.
type ResourceChange struct {
	// Address is the full resource address, such as
	// "module.web.aws_instance.this".
	Address string
	// Type is the resource type, such as "aws_instance".
	Type string
	// Name is the resource name, such as "this".
	Name string
	// Action is the primary change Terraform will apply to the resource.
	Action Action
	// Before is the prior state, or nil when the resource is being created.
	Before map[string]any
	// After is the planned state, or nil when the resource is being destroyed.
	After map[string]any
	// AfterUnknown mirrors the shape of After, holding true where a value is
	// only known after apply.
	AfterUnknown map[string]any
	// BeforeSensitive mirrors the shape of Before, holding true where a value
	// is sensitive.
	BeforeSensitive map[string]any
	// AfterSensitive mirrors the shape of After, holding true where a value is
	// sensitive.
	AfterSensitive map[string]any
}

// Plan is a parsed Terraform plan, as returned by [Parse].
type Plan struct {
	// Changes holds every resource change in the plan, in file order.
	Changes []ResourceChange
}

// Parse decodes "terraform show -json" output into a [Plan]. It returns an
// error only when the top-level JSON is malformed; individual resources with an
// unexpected shape degrade gracefully rather than failing the whole parse.
func Parse(data []byte) (*Plan, error) {
	var raw rawPlan
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("decode plan json: %w", err)
	}

	p := &Plan{Changes: make([]ResourceChange, 0, len(raw.ResourceChanges))}
	for _, rc := range raw.ResourceChanges {
		typ, name := rc.Type, rc.Name
		if typ == "" || name == "" {
			typ, name = splitAddress(rc.Address)
		}
		p.Changes = append(p.Changes, ResourceChange{
			Address:         rc.Address,
			Type:            typ,
			Name:            name,
			Action:          actionFrom(rc.Change.Actions),
			Before:          asObject(rc.Change.Before),
			After:           asObject(rc.Change.After),
			AfterUnknown:    asObject(rc.Change.AfterUnknown),
			BeforeSensitive: asObject(rc.Change.BeforeSensitive),
			AfterSensitive:  asObject(rc.Change.AfterSensitive),
		})
	}
	return p, nil
}

// Count returns the number of resources whose primary [Action] is action.
func (p *Plan) Count(action Action) int {
	n := 0
	for _, rc := range p.Changes {
		if rc.Action == action {
			n++
		}
	}
	return n
}

// AddCount returns the number of resources to create; it is shorthand for
// Count([ActionCreate]).
func (p *Plan) AddCount() int { return p.Count(ActionCreate) }

// ChangeCount returns the number of resources to update in place; it is
// shorthand for Count([ActionUpdate]).
func (p *Plan) ChangeCount() int { return p.Count(ActionUpdate) }

// DestroyCount returns the number of resources to destroy; it is shorthand for
// Count([ActionDelete]).
func (p *Plan) DestroyCount() int { return p.Count(ActionDelete) }

// ReplaceCount returns the number of resources to destroy and recreate; it is
// shorthand for Count([ActionReplace]).
func (p *Plan) ReplaceCount() int { return p.Count(ActionReplace) }

// HasChanges reports whether the plan contains at least one create, update,
// delete or replace action. It is equivalent to len([Plan.Changed]) > 0.
func (p *Plan) HasChanges() bool { return len(p.Changed()) > 0 }

// Changed returns the resources with a visible change — create, update, delete
// or replace — in plan order. No-op and read resources are omitted.
func (p *Plan) Changed() []ResourceChange {
	out := make([]ResourceChange, 0, len(p.Changes))
	for _, rc := range p.Changes {
		switch rc.Action {
		case ActionCreate, ActionUpdate, ActionDelete, ActionReplace:
			out = append(out, rc)
		}
	}
	return out
}

type rawPlan struct {
	ResourceChanges []rawResourceChange `json:"resource_changes"`
}

type rawResourceChange struct {
	Address string    `json:"address"`
	Type    string    `json:"type"`
	Name    string    `json:"name"`
	Change  rawChange `json:"change"`
}

type rawChange struct {
	Actions         []string        `json:"actions"`
	Before          json.RawMessage `json:"before"`
	After           json.RawMessage `json:"after"`
	AfterUnknown    json.RawMessage `json:"after_unknown"`
	BeforeSensitive json.RawMessage `json:"before_sensitive"`
	AfterSensitive  json.RawMessage `json:"after_sensitive"`
}

func actionFrom(actions []string) Action {
	switch {
	case len(actions) == 2 && slices.Contains(actions, "create") && slices.Contains(actions, "delete"):
		return ActionReplace
	case slices.Contains(actions, "create"):
		return ActionCreate
	case slices.Contains(actions, "delete"):
		return ActionDelete
	case slices.Contains(actions, "update"):
		return ActionUpdate
	case slices.Contains(actions, "read"):
		return ActionRead
	default:
		return ActionNoOp
	}
}

func asObject(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return m
}

func splitAddress(address string) (typ, name string) {
	i := strings.LastIndex(address, ".")
	if i < 0 {
		return address, ""
	}
	name = address[i+1:]
	rest := address[:i]
	if j := strings.LastIndex(rest, "."); j >= 0 {
		return rest[j+1:], name
	}
	return rest, name
}
