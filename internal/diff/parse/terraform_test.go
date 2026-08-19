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

func TestTerraform_Replace(t *testing.T) {
	// A replace (delete,create) is diffed like an update: field changes show
	// up as Modified children rather than a whole-resource add/remove.
	plan := []byte(`{
		"resource_changes": [{
			"address": "aws_db.main", "type": "aws_db", "name": "main",
			"change": {
				"actions": ["delete","create"],
				"before": {"engine": "postgres"},
				"after":  {"engine": "mysql"}
			}
		}]
	}`)
	d, err := Terraform(plan)
	if err != nil {
		t.Fatal(err)
	}
	if d.Modified != 1 {
		t.Errorf("modified=%d, want 1", d.Modified)
	}
	if len(d.Root.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(d.Root.Children))
	}
	engine := findChild(d.Root.Children[0], "engine")
	if engine == nil || engine.Kind != node.Modified {
		t.Errorf("engine child = %v, want Modified", engine)
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

func scopePlan(before, after string) []byte {
	return []byte(`{
		"resource_changes": [{
			"address": "auth0_resource_server.api",
			"type": "auth0_resource_server",
			"name": "api",
			"change": {
				"actions": ["update"],
				"before": {"identifier": "https://api.example.com", "scopes": ` + before + `},
				"after":  {"identifier": "https://api.example.com", "scopes": ` + after + `}
			}
		}]
	}`)
}

func TestTerraform_ScopeInsertion(t *testing.T) {
	scope := func(v string) string {
		return `{"value":"` + v + `","description":"` + v + `"}`
	}
	before := "[" + scope("create:users") + "," + scope("delete:users") + "," + scope("read:users") + "]"
	after := "[" + scope("admin:all") + "," + scope("create:users") + "," + scope("delete:users") + "," + scope("read:users") + "]"

	d, err := Terraform(scopePlan(before, after))
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 1 || d.Removed != 0 || d.Modified != 0 {
		t.Errorf("added=%d removed=%d modified=%d, want 1 0 0", d.Added, d.Removed, d.Modified)
	}

	scopes := findChild(d.Root.Children[0], "scopes")
	if scopes == nil {
		t.Fatal("no 'scopes' child in resource node")
	}
	if len(scopes.Children) != 4 {
		t.Fatalf("got %d scope children, want 4", len(scopes.Children))
	}
	for i, child := range scopes.Children {
		want := node.Unchanged
		if i == 0 {
			want = node.Added
		}
		if child.Kind != want {
			t.Errorf("scopes[%d] kind=%v, want %v", i, child.Kind, want)
		}
	}
}

func TestTerraform_ScopeReorderIsUnchanged(t *testing.T) {
	before := `["read:users","create:users","delete:users"]`
	after := `["create:users","delete:users","read:users"]`

	d, err := Terraform(scopePlan(before, after))
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 0 || d.Removed != 0 || d.Modified != 0 {
		t.Errorf("added=%d removed=%d modified=%d, want 0 0 0", d.Added, d.Removed, d.Modified)
	}
	scopes := findChild(d.Root.Children[0], "scopes")
	if scopes == nil {
		t.Fatal("no 'scopes' child in resource node")
	}
	for i, child := range scopes.Children {
		if child.Kind != node.Unchanged {
			t.Errorf("scopes[%d] kind=%v, want Unchanged", i, child.Kind)
		}
	}
}

func TestTerraform_ScopeRemovalAndEdit(t *testing.T) {
	d, err := Terraform(scopePlan(`["create:users","delete:users","read:users"]`, `["create:users","read:users"]`))
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 0 || d.Removed != 1 || d.Modified != 0 {
		t.Errorf("removal: added=%d removed=%d modified=%d, want 0 1 0", d.Added, d.Removed, d.Modified)
	}

	d, err = Terraform(scopePlan(`["create:users","read:users"]`, `["create:users","read:reports"]`))
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 0 || d.Removed != 0 || d.Modified != 1 {
		t.Errorf("edit: added=%d removed=%d modified=%d, want 0 0 1", d.Added, d.Removed, d.Modified)
	}
}
