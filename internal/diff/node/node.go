// Package node defines the core data types for diff results.
package node

// ChangeKind classifies a line or structural element.
type ChangeKind int

const (
	Unchanged ChangeKind = iota
	Added
	Removed
	Modified // structured diffs only (phases 4–6)
)

// Format identifies the detected file format.
type Format int

const (
	FormatText      Format = iota
	FormatJSON
	FormatYAML
	FormatTerraform
)

func (f Format) String() string {
	switch f {
	case FormatJSON:
		return "json"
	case FormatYAML:
		return "yaml"
	case FormatTerraform:
		return "terraform"
	default:
		return "text"
	}
}

// Line is a single line in a text diff hunk.
type Line struct {
	Kind    ChangeKind
	OldNum  int    // 0 when Added
	NewNum  int    // 0 when Removed
	Content string // raw line content, no leading +/-
}

// Hunk is a contiguous group of changes plus surrounding context.
type Hunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Lines    []Line
}

// DiffNode represents a single node in a structured diff tree (phases 4–6).
type DiffNode struct {
	Kind     ChangeKind
	Path     string // e.g. ".users[0].name"
	Key      string
	Index    int    // -1 if not an array element
	OldValue string // raw leaf value
	NewValue string
	Children []*DiffNode
}

// Diff is the top-level result returned by all parse functions.
type Diff struct {
	Format   Format
	FileA    string
	FileB    string
	Hunks    []Hunk    // populated for text diffs
	Root     *DiffNode // populated for structured diffs (phases 4–6)
	Added    int
	Removed  int
	Modified int // structured diffs only
}
