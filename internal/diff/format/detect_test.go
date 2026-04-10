package format_test

import (
	"testing"

	"github.com/danielriddell21/unum/internal/diff/format"
	"github.com/danielriddell21/unum/internal/diff/node"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		fileA, fileB string
		want         node.Format
	}{
		{"a.json", "b.json", node.FormatJSON},
		{"a.yaml", "b.yaml", node.FormatYAML},
		{"a.yml", "b.yml", node.FormatYAML},
		{"a.tfplan", "b.tfplan", node.FormatTerraform},
		{"a.txt", "b.txt", node.FormatText},
		{"a.go", "b.go", node.FormatText},
		{"", "", node.FormatText},
	}
	for _, tt := range tests {
		got := format.Detect(tt.fileA, tt.fileB)
		if got != tt.want {
			t.Errorf("Detect(%q, %q) = %v, want %v", tt.fileA, tt.fileB, got, tt.want)
		}
	}
}

func TestDetect_FileBPrecedence(t *testing.T) {
	// fileB is checked first; if recognised, it wins over fileA.
	got := format.Detect("a.txt", "b.json")
	if got != node.FormatJSON {
		t.Errorf("got %v, want FormatJSON (fileB takes precedence)", got)
	}
	// If fileB is unrecognised, fileA is used as a fallback.
	got = format.Detect("a.json", "b.txt")
	if got != node.FormatJSON {
		t.Errorf("got %v, want FormatJSON (fileA fallback when fileB unrecognised)", got)
	}
}

func TestDetect_CaseInsensitive(t *testing.T) {
	got := format.Detect("a.JSON", "b.JSON")
	if got != node.FormatJSON {
		t.Errorf("got %v, want FormatJSON (extension case insensitive)", got)
	}
	got = format.Detect("a.YAML", "b.YAML")
	if got != node.FormatYAML {
		t.Errorf("got %v, want FormatYAML (extension case insensitive)", got)
	}
}

func TestParse_ExplicitFlag(t *testing.T) {
	tests := []struct {
		flag string
		want node.Format
	}{
		{"json", node.FormatJSON},
		{"yaml", node.FormatYAML},
		{"terraform", node.FormatTerraform},
		{"tf", node.FormatTerraform},
		{"text", node.FormatText},
	}
	for _, tt := range tests {
		got := format.Parse(tt.flag, "a.txt", "b.txt")
		if got != tt.want {
			t.Errorf("Parse(%q, ...) = %v, want %v", tt.flag, got, tt.want)
		}
	}
}

func TestParse_FlagCaseInsensitive(t *testing.T) {
	got := format.Parse("JSON", "a.txt", "b.txt")
	if got != node.FormatJSON {
		t.Errorf("got %v, want FormatJSON (flag case insensitive)", got)
	}
}

func TestParse_EmptyFlagFallsBackToDetect(t *testing.T) {
	got := format.Parse("", "a.yaml", "b.yaml")
	if got != node.FormatYAML {
		t.Errorf("got %v, want FormatYAML (empty flag should auto-detect)", got)
	}
}
