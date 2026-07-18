package diagramtool

import (
	"path/filepath"
	"strings"
)

type Language int

const (
	LangUnknown Language = iota
	LangMermaid
	LangD2
)

func (l Language) String() string {
	switch l {
	case LangMermaid:
		return "mermaid"
	case LangD2:
		return "d2"
	default:
		return "unknown"
	}
}

type Format int

const (
	FormatSVG Format = iota
	FormatPNG
	FormatDrawio
)

func (f Format) String() string {
	switch f {
	case FormatPNG:
		return "png"
	case FormatDrawio:
		return "drawio"
	default:
		return "svg"
	}
}

func DetectLanguage(file string) Language {
	switch strings.ToLower(filepath.Ext(file)) {
	case ".mmd", ".mermaid":
		return LangMermaid
	case ".d2":
		return LangD2
	}
	return LangUnknown
}

func ParseLanguage(flag, file string) Language {
	switch strings.ToLower(flag) {
	case "mermaid", "mmd":
		return LangMermaid
	case "d2":
		return LangD2
	}
	return DetectLanguage(file)
}

func ParseFormat(flag string) (Format, bool) {
	switch strings.ToLower(flag) {
	case "", "svg":
		return FormatSVG, true
	case "png":
		return FormatPNG, true
	case "drawio", "draw.io":
		return FormatDrawio, true
	}
	return FormatSVG, false
}
