package format

import (
	"path/filepath"
	"strings"

	"github.com/danielriddell21/unum/internal/diff/node"
)

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
