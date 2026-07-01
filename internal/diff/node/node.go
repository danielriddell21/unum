package node

type ChangeKind int

const (
	Unchanged ChangeKind = iota
	Added
	Removed
	Modified
)

type Format int

const (
	FormatText Format = iota
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

type Line struct {
	Kind    ChangeKind
	OldNum  int
	NewNum  int
	Content string
}

type Hunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Lines    []Line
}

type DiffNode struct {
	Kind     ChangeKind
	Path     string
	Key      string
	Index    int
	OldValue string
	NewValue string
	Children []*DiffNode
}

type Diff struct {
	Format   Format
	FileA    string
	FileB    string
	Hunks    []Hunk
	Root     *DiffNode
	Added    int
	Removed  int
	Modified int
}
