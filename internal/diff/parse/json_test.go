package parse

import (
	"testing"

	"github.com/danielriddell21/unum/internal/diff/node"
)

func TestJSON_Identical(t *testing.T) {
	a := []byte(`{"name":"alice","age":30}`)
	d, err := JSON(a, a)
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 0 || d.Removed != 0 || d.Modified != 0 {
		t.Errorf("added=%d removed=%d modified=%d, want 0 0 0", d.Added, d.Removed, d.Modified)
	}
	if d.Root == nil {
		t.Fatal("Root is nil")
	}
	if d.Root.Kind != node.Unchanged {
		t.Errorf("root kind=%v, want Unchanged", d.Root.Kind)
	}
}

func TestJSON_AddedKey(t *testing.T) {
	a := []byte(`{"name":"alice"}`)
	b := []byte(`{"name":"alice","age":30}`)
	d, err := JSON(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 1 {
		t.Errorf("added=%d, want 1", d.Added)
	}
	child := findChild(d.Root, "age")
	if child == nil {
		t.Fatal("no child with key 'age'")
	}
	if child.Kind != node.Added {
		t.Errorf("age kind=%v, want Added", child.Kind)
	}
	if child.NewValue != "30" {
		t.Errorf("age NewValue=%q, want \"30\"", child.NewValue)
	}
}

func TestJSON_RemovedKey(t *testing.T) {
	a := []byte(`{"name":"alice","age":30}`)
	b := []byte(`{"name":"alice"}`)
	d, err := JSON(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Removed != 1 {
		t.Errorf("removed=%d, want 1", d.Removed)
	}
	child := findChild(d.Root, "age")
	if child == nil {
		t.Fatal("no child with key 'age'")
	}
	if child.Kind != node.Removed {
		t.Errorf("age kind=%v, want Removed", child.Kind)
	}
}

func TestJSON_ModifiedScalar(t *testing.T) {
	a := []byte(`{"version":"1.0"}`)
	b := []byte(`{"version":"2.0"}`)
	d, err := JSON(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Modified != 1 {
		t.Errorf("modified=%d, want 1", d.Modified)
	}
	child := findChild(d.Root, "version")
	if child == nil {
		t.Fatal("no child with key 'version'")
	}
	if child.Kind != node.Modified {
		t.Errorf("version kind=%v, want Modified", child.Kind)
	}
	if child.OldValue != `"1.0"` || child.NewValue != `"2.0"` {
		t.Errorf("version old=%q new=%q, want \"1.0\" \"2.0\"", child.OldValue, child.NewValue)
	}
}

func TestJSON_TypeChange(t *testing.T) {
	// Changing from string to number is a type mismatch — treated as Modified.
	a := []byte(`{"count":"5"}`)
	b := []byte(`{"count":5}`)
	d, err := JSON(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Modified != 1 {
		t.Errorf("modified=%d, want 1", d.Modified)
	}
}

func TestJSON_NestedObject(t *testing.T) {
	a := []byte(`{"server":{"host":"localhost","port":8080}}`)
	b := []byte(`{"server":{"host":"prod.example.com","port":8080}}`)
	d, err := JSON(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Modified != 1 {
		t.Errorf("modified=%d, want 1", d.Modified)
	}
	server := findChild(d.Root, "server")
	if server == nil {
		t.Fatal("no 'server' child")
	}
	host := findChild(server, "host")
	if host == nil {
		t.Fatal("no 'host' grandchild")
	}
	if host.Kind != node.Modified {
		t.Errorf("host kind=%v, want Modified", host.Kind)
	}
}

func TestJSON_ArrayAddedElement(t *testing.T) {
	a := []byte(`{"items":[1,2]}`)
	b := []byte(`{"items":[1,2,3]}`)
	d, err := JSON(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 1 {
		t.Errorf("added=%d, want 1", d.Added)
	}
}

func TestJSON_ArrayRemovedElement(t *testing.T) {
	a := []byte(`{"items":[1,2,3]}`)
	b := []byte(`{"items":[1,2]}`)
	d, err := JSON(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Removed != 1 {
		t.Errorf("removed=%d, want 1", d.Removed)
	}
}

func TestJSON_RootLevelPath_NoDoubleDot(t *testing.T) {
	// Regression: root-level keys must produce ".key", not "..key".
	a := []byte(`{"version":"1.0"}`)
	b := []byte(`{"version":"2.0"}`)
	d, err := JSON(a, b)
	if err != nil {
		t.Fatal(err)
	}
	child := findChild(d.Root, "version")
	if child == nil {
		t.Fatal("no 'version' child")
	}
	if child.Path != ".version" {
		t.Errorf("path=%q, want \".version\" (double-dot regression)", child.Path)
	}
}

func TestJSON_PopulatesHunks(t *testing.T) {
	// JSON diff should also populate Hunks so web UI can offer unified/split view.
	a := []byte(`{"version":"1.0"}`)
	b := []byte(`{"version":"2.0"}`)
	d, err := JSON(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Hunks) == 0 {
		t.Error("Hunks are empty; JSON diff should populate both Root and Hunks")
	}
}

func TestJSON_InvalidInput(t *testing.T) {
	_, err := JSON([]byte(`{broken`), []byte(`{}`))
	if err == nil {
		t.Error("expected error for invalid JSON in file A")
	}
	_, err = JSON([]byte(`{}`), []byte(`{broken`))
	if err == nil {
		t.Error("expected error for invalid JSON in file B")
	}
}

// findChild returns the first direct child of n with the given key, or nil.
func findChild(n *node.DiffNode, key string) *node.DiffNode {
	if n == nil {
		return nil
	}
	for _, c := range n.Children {
		if c.Key == key {
			return c
		}
	}
	return nil
}
