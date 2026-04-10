package web

import (
	"testing"

	"github.com/danielriddell21/unum/internal/diff/node"
)

func textDiff() *node.Diff {
	return &node.Diff{
		Format: node.FormatText,
		FileA:  "a.txt",
		FileB:  "b.txt",
		Added:  2,
		Removed: 1,
		Hunks: []node.Hunk{
			{
				OldStart: 1, OldCount: 3,
				NewStart: 1, NewCount: 4,
				Lines: []node.Line{
					{Kind: node.Unchanged, OldNum: 1, NewNum: 1, Content: "same"},
					{Kind: node.Removed, OldNum: 2, Content: "old line"},
					{Kind: node.Added, NewNum: 2, Content: "new line 1"},
					{Kind: node.Added, NewNum: 3, Content: "new line 2"},
					{Kind: node.Unchanged, OldNum: 3, NewNum: 4, Content: "same"},
				},
			},
		},
	}
}

func jsonDiff() *node.Diff {
	return &node.Diff{
		Format:   node.FormatJSON,
		FileA:    "a.json",
		FileB:    "b.json",
		Added:    1,
		Removed:  0,
		Modified: 1,
		Root: &node.DiffNode{
			Kind: node.Unchanged,
			Children: []*node.DiffNode{
				{Kind: node.Modified, Path: ".version", Key: "version", OldValue: `"1.0"`, NewValue: `"2.0"`},
				{Kind: node.Added, Path: ".region", Key: "region", NewValue: `"us-east-1"`},
			},
		},
		Hunks: []node.Hunk{
			{
				OldStart: 1, OldCount: 1, NewStart: 1, NewCount: 1,
				Lines: []node.Line{
					{Kind: node.Removed, OldNum: 1, Content: `"version": "1.0"`},
					{Kind: node.Added, NewNum: 1, Content: `"version": "2.0"`},
				},
			},
		},
	}
}

func terraformDiff() *node.Diff {
	return &node.Diff{
		Format:  node.FormatTerraform,
		FileA:   "plan.json",
		FileB:   "plan.json",
		Added:   1,
		Removed: 0,
		Root: &node.DiffNode{
			Kind: node.Unchanged,
			Children: []*node.DiffNode{
				{Kind: node.Added, Path: ".aws_instance.web", Key: "aws_instance.web"},
			},
		},
	}
}

// ─── buildPayload ─────────────────────────────────────────────────────────────

func TestBuildPayload_TextDiff(t *testing.T) {
	d := textDiff()
	p, err := buildPayload(d)
	if err != nil {
		t.Fatal(err)
	}

	if p.FileA != "a.txt" || p.FileB != "b.txt" {
		t.Errorf("files: got %q %q", p.FileA, p.FileB)
	}
	if p.Format != "text" {
		t.Errorf("format=%q, want \"text\"", p.Format)
	}
	if p.Added != 2 || p.Removed != 1 {
		t.Errorf("added=%d removed=%d, want 2 1", p.Added, p.Removed)
	}
	if len(p.Hunks) != 1 {
		t.Fatalf("hunks=%d, want 1", len(p.Hunks))
	}
	if len(p.Hunks[0].Lines) != 5 {
		t.Errorf("hunk lines=%d, want 5", len(p.Hunks[0].Lines))
	}
	if p.Tree != nil {
		t.Error("text diff should have nil Tree")
	}
}

func TestBuildPayload_JSONDiff(t *testing.T) {
	d := jsonDiff()
	p, err := buildPayload(d)
	if err != nil {
		t.Fatal(err)
	}

	if p.Format != "json" {
		t.Errorf("format=%q, want \"json\"", p.Format)
	}
	if p.Tree == nil {
		t.Fatal("json diff should have non-nil Tree")
	}
	if len(p.Hunks) == 0 {
		t.Error("json diff with hunks should populate Hunks")
	}
	if len(p.Tree.Children) != 2 {
		t.Errorf("tree children=%d, want 2", len(p.Tree.Children))
	}
}

func TestBuildPayload_TreeNodeKinds(t *testing.T) {
	d := jsonDiff()
	p, err := buildPayload(d)
	if err != nil {
		t.Fatal(err)
	}

	modifiedChild := p.Tree.Children[0]
	if modifiedChild.Kind != "modified" {
		t.Errorf("child[0] kind=%q, want \"modified\"", modifiedChild.Kind)
	}
	if modifiedChild.OldValue != `"1.0"` || modifiedChild.NewValue != `"2.0"` {
		t.Errorf("child[0] values: old=%q new=%q", modifiedChild.OldValue, modifiedChild.NewValue)
	}

	addedChild := p.Tree.Children[1]
	if addedChild.Kind != "added" {
		t.Errorf("child[1] kind=%q, want \"added\"", addedChild.Kind)
	}
}

func TestBuildPayload_TerraformDiff(t *testing.T) {
	d := terraformDiff()
	p, err := buildPayload(d)
	if err != nil {
		t.Fatal(err)
	}

	if p.Format != "terraform" {
		t.Errorf("format=%q, want \"terraform\"", p.Format)
	}
	if p.Tree == nil {
		t.Fatal("terraform diff should have non-nil Tree")
	}
	if len(p.Hunks) != 0 {
		t.Errorf("terraform diff should have no hunks, got %d", len(p.Hunks))
	}
}

func TestBuildPayload_HunkLineKinds(t *testing.T) {
	d := textDiff()
	p, err := buildPayload(d)
	if err != nil {
		t.Fatal(err)
	}

	lines := p.Hunks[0].Lines
	kinds := make([]string, len(lines))
	for i, l := range lines {
		kinds[i] = l.Kind
	}

	want := []string{"unchanged", "removed", "added", "added", "unchanged"}
	for i, k := range want {
		if kinds[i] != k {
			t.Errorf("line[%d] kind=%q, want %q", i, kinds[i], k)
		}
	}
}
