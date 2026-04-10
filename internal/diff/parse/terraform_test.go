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
