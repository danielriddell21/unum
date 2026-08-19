package parse

import (
	"testing"

	"github.com/danielriddell21/unum/internal/diff/node"
)

func TestYAML_Identical(t *testing.T) {
	a := []byte("name: alice\nage: 30\n")
	d, err := YAML(a, a)
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 0 || d.Removed != 0 || d.Modified != 0 {
		t.Errorf("added=%d removed=%d modified=%d, want 0 0 0", d.Added, d.Removed, d.Modified)
	}
	if d.Root.Kind != node.Unchanged {
		t.Errorf("root kind=%v, want Unchanged", d.Root.Kind)
	}
}

func TestYAML_AddedKey(t *testing.T) {
	a := []byte("name: alice\n")
	b := []byte("name: alice\nrole: admin\n")
	d, err := YAML(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 1 {
		t.Errorf("added=%d, want 1", d.Added)
	}
	child := findChild(d.Root, "role")
	if child == nil {
		t.Fatal("no child with key 'role'")
	}
	if child.Kind != node.Added {
		t.Errorf("role kind=%v, want Added", child.Kind)
	}
}

func TestYAML_RemovedKey(t *testing.T) {
	a := []byte("name: alice\nrole: admin\n")
	b := []byte("name: alice\n")
	d, err := YAML(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Removed != 1 {
		t.Errorf("removed=%d, want 1", d.Removed)
	}
	child := findChild(d.Root, "role")
	if child == nil {
		t.Fatal("no 'role' child in diff")
	}
	if child.Kind != node.Removed {
		t.Errorf("role kind=%v, want Removed", child.Kind)
	}
}

func TestYAML_ModifiedScalar(t *testing.T) {
	a := []byte("host: localhost\n")
	b := []byte("host: prod.example.com\n")
	d, err := YAML(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Modified != 1 {
		t.Errorf("modified=%d, want 1", d.Modified)
	}
	child := findChild(d.Root, "host")
	if child == nil {
		t.Fatal("no 'host' child")
	}
	if child.Kind != node.Modified {
		t.Errorf("host kind=%v, want Modified", child.Kind)
	}
}

func TestYAML_SequenceAddedElement(t *testing.T) {
	a := []byte("ports:\n  - 80\n  - 443\n")
	b := []byte("ports:\n  - 80\n  - 443\n  - 8080\n")
	d, err := YAML(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 1 {
		t.Errorf("added=%d, want 1", d.Added)
	}
}

func TestYAML_SequenceRemovedElement(t *testing.T) {
	a := []byte("ports:\n  - 80\n  - 443\n  - 8080\n")
	b := []byte("ports:\n  - 80\n  - 443\n")
	d, err := YAML(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Removed != 1 {
		t.Errorf("removed=%d, want 1", d.Removed)
	}
}

func TestYAML_NullScalarDisplay(t *testing.T) {
	// A null value in YAML should display as "null", not as an empty string.
	a := []byte("value: something\n")
	b := []byte("value: null\n")
	d, err := YAML(a, b)
	if err != nil {
		t.Fatal(err)
	}
	child := findChild(d.Root, "value")
	if child == nil {
		t.Fatal("no 'value' child")
	}
	if child.NewValue != "null" {
		t.Errorf("null display=%q, want \"null\"", child.NewValue)
	}
}

func TestYAML_StringScalarQuoted(t *testing.T) {
	// !!str values should be displayed with quotes so they are distinguishable
	// from numbers/booleans.
	a := []byte("label: old\n")
	b := []byte("label: new\n")
	d, err := YAML(a, b)
	if err != nil {
		t.Fatal(err)
	}
	child := findChild(d.Root, "label")
	if child == nil {
		t.Fatal("no 'label' child")
	}
	if child.Kind != node.Modified {
		t.Fatalf("label kind=%v, want Modified", child.Kind)
	}
	// yamlScalarDisplay wraps !!str values in Go's %q quotes
	if child.OldValue != `"old"` {
		t.Errorf("OldValue=%q, want `\"old\"`", child.OldValue)
	}
	if child.NewValue != `"new"` {
		t.Errorf("NewValue=%q, want `\"new\"`", child.NewValue)
	}
}

func TestYAML_PopulatesHunks(t *testing.T) {
	a := []byte("version: 1.0\n")
	b := []byte("version: 2.0\n")
	d, err := YAML(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Hunks) == 0 {
		t.Error("Hunks are empty; YAML diff should populate both Root and Hunks")
	}
}

func TestYAML_InvalidInput(t *testing.T) {
	_, err := YAML([]byte("key: [unclosed"), []byte("key: value\n"))
	if err == nil {
		t.Error("expected error for invalid YAML in file A")
	}
}

func TestYAML_RootLevelPath_NoDoubleDot(t *testing.T) {
	a := []byte("version: 1.0\n")
	b := []byte("version: 2.0\n")
	d, err := YAML(a, b)
	if err != nil {
		t.Fatal(err)
	}
	child := findChild(d.Root, "version")
	if child == nil {
		t.Fatal("no 'version' child")
	}
	if child.Path != ".version" {
		t.Errorf("path=%q, want \".version\"", child.Path)
	}
}

func TestYAML_SequenceAlignment(t *testing.T) {
	tests := []struct {
		name                                 string
		a, b                                 string
		wantAdded, wantRemoved, wantModified int
	}{
		{
			name:      "insert at front shifts nothing",
			a:         "ports:\n  - 443\n  - 8080\n",
			b:         "ports:\n  - 80\n  - 443\n  - 8080\n",
			wantAdded: 1,
		},
		{
			name:      "insert in the middle shifts nothing",
			a:         "ports:\n  - 80\n  - 8080\n",
			b:         "ports:\n  - 80\n  - 443\n  - 8080\n",
			wantAdded: 1,
		},
		{
			name:        "removal from the middle shifts nothing",
			a:           "ports:\n  - 80\n  - 443\n  - 8080\n",
			b:           "ports:\n  - 80\n  - 8080\n",
			wantRemoved: 1,
		},
		{
			name:         "in-place edit stays a modification",
			a:            "ports:\n  - 80\n  - 443\n",
			b:            "ports:\n  - 80\n  - 8443\n",
			wantModified: 1,
		},
		{
			name: "reorder is unchanged",
			a:    "ports:\n  - 8080\n  - 80\n  - 443\n",
			b:    "ports:\n  - 80\n  - 443\n  - 8080\n",
		},
		{
			name:      "mapping elements align on content",
			a:         "scopes:\n  - value: a\n  - value: c\n",
			b:         "scopes:\n  - value: a\n  - value: b\n  - value: c\n",
			wantAdded: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := YAML([]byte(tt.a), []byte(tt.b))
			if err != nil {
				t.Fatal(err)
			}
			if d.Added != tt.wantAdded || d.Removed != tt.wantRemoved || d.Modified != tt.wantModified {
				t.Errorf("added=%d removed=%d modified=%d, want %d %d %d",
					d.Added, d.Removed, d.Modified, tt.wantAdded, tt.wantRemoved, tt.wantModified)
			}
		})
	}
}

func TestYAML_NestedSequenceElementsAlign(t *testing.T) {
	a := []byte("matrix:\n  - [1, 2]\n  - [5, 6]\n")
	b := []byte("matrix:\n  - [1, 2]\n  - [3, 4]\n  - [5, 6]\n")
	d, err := YAML(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 1 || d.Removed != 0 || d.Modified != 0 {
		t.Errorf("added=%d removed=%d modified=%d, want 1 0 0", d.Added, d.Removed, d.Modified)
	}
}
