package tui

import (
	"path/filepath"
	"strings"
)

func replaceExt(path, ext string) string {
	if path == "" {
		return path
	}
	old := filepath.Ext(path)
	return strings.TrimSuffix(path, old) + ext
}
