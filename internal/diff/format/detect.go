// Package format detects the file format for diff operations.
package format

import (
	"path/filepath"
	"strings"

	"github.com/danielriddell21/unum/internal/diff/node"
)

// Detect infers the diff format from file extensions.
// The extension of fileB (the "new" file) takes precedence.
// Returns FormatText for unrecognised extensions.
func Detect(fileA, fileB string) node.Format {
	for _, f := range []string{fileB, fileA} {
		switch strings.ToLower(filepath.Ext(f)) {
		case ".json":
			return node.FormatJSON
		case ".yaml", ".yml":
			return node.FormatYAML
		case ".tfplan":
			return node.FormatTerraform
		}
	}
	return node.FormatText
}

// Parse converts a flag value to a Format, defaulting to auto-detection.
func Parse(flag, fileA, fileB string) node.Format {
	switch strings.ToLower(flag) {
	case "json":
		return node.FormatJSON
	case "yaml":
		return node.FormatYAML
	case "terraform", "tf":
		return node.FormatTerraform
	case "text":
		return node.FormatText
	}
	return Detect(fileA, fileB)
}
