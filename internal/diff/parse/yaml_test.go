package parse

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

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

func TestYAML_AliasExpansion(t *testing.T) {
	bomb := func(levels int) []byte {
		var sb strings.Builder
		sb.WriteString(`a: &a ["x","x","x","x","x","x","x","x","x"]` + "\n")
		prev := "a"
		for i := 1; i < levels; i++ {
			name := string(rune('a' + i))
			fmt.Fprintf(&sb, "%s: &%s [%s]\n", name, name, strings.TrimSuffix(strings.Repeat("*"+prev+",", 9), ","))
			prev = name
		}
		return []byte(sb.String())
	}

	tests := []struct {
		name     string
		a, b     []byte
		wantErr  bool
		modified int
	}{
		{
			name: "plain anchor and alias",
			a:    []byte("base: &b {x: 1}\nother: *b\n"),
			b:    []byte("base: &b {x: 1}\nother: *b\n"),
		},
		{
			name:     "alias resolves to anchored value",
			a:        []byte("base: &b {x: 1}\nother: *b\n"),
			b:        []byte("base: &b {x: 2}\nother: *b\n"),
			modified: 2,
		},
		{
			name: "moderate expansion passes",
			a:    bomb(4),
			b:    bomb(4),
		},
		{
			name:    "billion laughs from the issue",
			a:       bomb(7),
			b:       bomb(7),
			wantErr: true,
		},
		{
			name:    "self-referencing alias",
			a:       []byte("a: &a [*a]\n"),
			b:       []byte("a: &a [*a]\n"),
			wantErr: true,
		},
		{
			name:    "one level larger",
			a:       bomb(8),
			b:       bomb(8),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			d, err := YAML(tt.a, tt.b)
			if elapsed := time.Since(start); elapsed > 5*time.Second {
				t.Errorf("took %v, want fail-fast", elapsed)
			}
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected alias expansion error")
				}
				if !strings.Contains(err.Error(), "alias expansion") {
					t.Errorf("err=%v, want alias expansion error", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if d.Modified != tt.modified {
				t.Errorf("modified=%d, want %d", d.Modified, tt.modified)
			}
		})
	}
}

func TestYAML_AliasBudget(t *testing.T) {
	// maxYAMLNodes is the total number of nodes YAML will compare, counting
	// every node reached through an alias again; one million keeps real
	// manifests far from the limit while stopping exponential expansion.
	if maxYAMLNodes != 1_000_000 {
		t.Errorf("maxYAMLNodes=%d, want 1000000", maxYAMLNodes)
	}
	// maxYAMLDepth caps recursion so a self-referencing alias errors out
	// instead of exhausting the stack; it matches the JSON parser's depth cap.
	if maxYAMLDepth != 10_000 {
		t.Errorf("maxYAMLDepth=%d, want 10000", maxYAMLDepth)
	}
}

func TestYAMLKeys(t *testing.T) {
	key := func(src string) string {
		n, err := parseYAMLDoc([]byte(src))
		if err != nil {
			t.Fatal(err)
		}
		return (&yamlState{keys: map[*yaml.Node]string{}}).key(n)
	}
	tests := []struct {
		name string
		a, b string
		same bool
	}{
		{"identical", "[1, 2]", "[1, 2]", true},
		{"mapping order ignored", "{a: 1, b: 2}", "{b: 2, a: 1}", true},
		{"sequence order matters", "[1, 2]", "[2, 1]", false},
		{"alias equals inline", "a: &x [1]\nb: *x\n", "a: [1]\nb: [1]\n", true},
		{"string vs int", "[1]", `["1"]`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := key(tt.a) == key(tt.b); got != tt.same {
				t.Errorf("key equality=%v, want %v", got, tt.same)
			}
		})
	}
}
