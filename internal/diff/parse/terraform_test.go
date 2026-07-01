package parse

import (
	"testing"

	"github.com/danielriddell21/unum/internal/diff/node"
)

func TestTerraform_Create(t *testing.T) {
	plan := []byte(`{
		"resource_changes": [{
			"address": "aws_instance.web",
			"type": "aws_instance",
			"name": "web",
			"change": {
				"actions": ["create"],
				"before": null,
				"after": {"ami": "ami-123", "instance_type": "t2.micro"}
			}
		}]
	}`)
	d, err := Terraform(plan)
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 1 || d.Removed != 0 {
		t.Errorf("added=%d removed=%d, want 1 0", d.Added, d.Removed)
	}
	if len(d.Root.Children) != 1 {
		t.Fatalf("got %d root children, want 1", len(d.Root.Children))
	}
	child := d.Root.Children[0]
	if child.Kind != node.Added {
		t.Errorf("kind=%v, want Added", child.Kind)
	}
	if child.Key != "aws_instance.web" {
		t.Errorf("key=%q, want \"aws_instance.web\"", child.Key)
	}
}

func TestTerraform_Delete(t *testing.T) {
	plan := []byte(`{
		"resource_changes": [{
			"address": "aws_instance.web",
			"type": "aws_instance",
			"name": "web",
			"change": {
				"actions": ["delete"],
				"before": {"ami": "ami-123"},
				"after": null
			}
		}]
	}`)
	d, err := Terraform(plan)
	if err != nil {
		t.Fatal(err)
	}
	if d.Removed != 1 {
		t.Errorf("removed=%d, want 1", d.Removed)
	}
	child := d.Root.Children[0]
	if child.Kind != node.Removed {
		t.Errorf("kind=%v, want Removed", child.Kind)
	}
}

func TestTerraform_Update(t *testing.T) {
	plan := []byte(`{
		"resource_changes": [{
			"address": "aws_instance.web",
			"type": "aws_instance",
			"name": "web",
			"change": {
				"actions": ["update"],
				"before": {"ami": "ami-old", "instance_type": "t2.micro"},
				"after":  {"ami": "ami-new", "instance_type": "t2.micro"}
			}
		}]
	}`)
	d, err := Terraform(plan)
	if err != nil {
		t.Fatal(err)
	}
	if d.Modified != 1 {
		t.Errorf("modified=%d, want 1 (the ami field changed)", d.Modified)
	}
	if len(d.Root.Children) != 1 {
		t.Fatalf("got %d root children, want 1", len(d.Root.Children))
	}
	resourceNode := d.Root.Children[0]
	if resourceNode.Kind != node.Unchanged {
		// The resource node wraps children; it is Unchanged at root level,
		// changed fields appear as Modified children.
		t.Errorf("resource node kind=%v, want Unchanged (changes are in children)", resourceNode.Kind)
	}
	// Find the 'ami' child that was modified
	ami := findChild(resourceNode, "ami")
	if ami == nil {
		t.Fatal("no 'ami' child in resource node")
	}
	if ami.Kind != node.Modified {
		t.Errorf("ami kind=%v, want Modified", ami.Kind)
	}
}

func TestTerraform_NoOp(t *testing.T) {
	plan := []byte(`{
		"resource_changes": [{
			"address": "aws_instance.web",
			"type": "aws_instance",
			"name": "web",
			"change": {
				"actions": ["no-op"],
				"before": {"ami": "ami-123"},
				"after":  {"ami": "ami-123"}
			}
		}]
	}`)
	d, err := Terraform(plan)
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 0 || d.Removed != 0 || d.Modified != 0 {
		t.Errorf("no-op: added=%d removed=%d modified=%d, want 0 0 0", d.Added, d.Removed, d.Modified)
	}
	if len(d.Root.Children) != 0 {
		t.Errorf("no-op: got %d children, want 0 (should be skipped)", len(d.Root.Children))
	}
}

func TestTerraform_MultipleResources(t *testing.T) {
	plan := []byte(`{
		"resource_changes": [
			{
				"address": "aws_instance.a", "type": "aws_instance", "name": "a",
				"change": {"actions": ["create"], "before": null, "after": {"x": 1}}
			},
			{
				"address": "aws_instance.b", "type": "aws_instance", "name": "b",
				"change": {"actions": ["delete"], "before": {"x": 1}, "after": null}
			},
			{
				"address": "aws_instance.c", "type": "aws_instance", "name": "c",
				"change": {"actions": ["no-op"], "before": {"x": 1}, "after": {"x": 1}}
			}
		]
	}`)
	d, err := Terraform(plan)
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 1 || d.Removed != 1 {
		t.Errorf("added=%d removed=%d, want 1 1", d.Added, d.Removed)
	}
	// Only create and delete appear as children; no-op is skipped
	if len(d.Root.Children) != 2 {
		t.Errorf("got %d children, want 2 (no-op skipped)", len(d.Root.Children))
	}
}

func TestTerraform_InvalidJSON(t *testing.T) {
	_, err := Terraform([]byte(`{broken`))
	if err == nil {
		t.Error("expected error for invalid plan JSON")
	}
}

func TestTerraform_Format(t *testing.T) {
	plan := []byte(`{"resource_changes":[]}`)
	d, err := Terraform(plan)
	if err != nil {
		t.Fatal(err)
	}
	if d.Format != node.FormatTerraform {
		t.Errorf("format=%v, want FormatTerraform", d.Format)
	}
}

func TestPrimaryAction_EmptySlice(t *testing.T) {
	// Empty actions slice → no match, no fallback → "no-op"
	got := primaryAction([]string{})
	if got != "no-op" {
		t.Errorf("primaryAction([])=%q, want \"no-op\"", got)
	}
}

func TestPrimaryAction_UnknownActionFallsBackToFirst(t *testing.T) {
	// A non-standard action should return the first element.
	got := primaryAction([]string{"replace", "read"})
	if got != "replace" {
		t.Errorf("primaryAction([replace,read])=%q, want \"replace\"", got)
	}
}

func TestPrimaryAction_KnownActions(t *testing.T) {
	for _, action := range []string{"create", "delete", "update"} {
		got := primaryAction([]string{action})
		if got != action {
			t.Errorf("primaryAction([%s])=%q, want %q", action, got, action)
		}
	}
}

func TestTerraform_UpdateFallback(t *testing.T) {
	// before contains a raw JSON fragment that cannot be parsed by jsonparse.Parse.
	plan := []byte(`{
		"resource_changes": [{
			"address": "aws_instance.web",
			"type": "aws_instance",
			"name": "web",
			"change": {
				"actions": ["update"],
				"before": "INVALID_RAW_FRAGMENT",
				"after":  {"ok": true}
			}
		}]
	}`)

	// The plan itself may or may not parse depending on how json.RawMessage
	// handles the value. Terraform() should either succeed with a fallback
	// Modified node, or fail — but never panic.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Terraform panicked on bad before: %v", r)
		}
	}()
	d, err := Terraform(plan)
	if err != nil {
		return // acceptable — outer unmarshal may reject it
	}
	if len(d.Root.Children) == 0 {
		t.Error("expected at least one child node from update fallback")
	}
}
