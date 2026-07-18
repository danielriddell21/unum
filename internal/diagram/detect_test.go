package diagramtool_test

import (
	"testing"

	diagramtool "github.com/danielriddell21/unum/internal/diagram"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		file string
		want diagramtool.Language
	}{
		{"diagram.mmd", diagramtool.LangMermaid},
		{"diagram.mermaid", diagramtool.LangMermaid},
		{"flow.D2", diagramtool.LangD2},
		{"graph.d2", diagramtool.LangD2},
		{"notes.txt", diagramtool.LangUnknown},
		{"", diagramtool.LangUnknown},
	}
	for _, tt := range tests {
		if got := diagramtool.DetectLanguage(tt.file); got != tt.want {
			t.Errorf("DetectLanguage(%q) = %v, want %v", tt.file, got, tt.want)
		}
	}
}

func TestParseLanguage(t *testing.T) {
	tests := []struct {
		flag, file string
		want       diagramtool.Language
	}{
		{"mermaid", "graph.d2", diagramtool.LangMermaid},
		{"mmd", "graph.d2", diagramtool.LangMermaid},
		{"d2", "diagram.mmd", diagramtool.LangD2},
		{"", "diagram.mmd", diagramtool.LangMermaid},
		{"", "graph.d2", diagramtool.LangD2},
		{"", "notes.txt", diagramtool.LangUnknown},
	}
	for _, tt := range tests {
		if got := diagramtool.ParseLanguage(tt.flag, tt.file); got != tt.want {
			t.Errorf("ParseLanguage(%q, %q) = %v, want %v", tt.flag, tt.file, got, tt.want)
		}
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		flag    string
		want    diagramtool.Format
		wantOK  bool
		wantStr string
	}{
		{"", diagramtool.FormatSVG, true, "svg"},
		{"svg", diagramtool.FormatSVG, true, "svg"},
		{"PNG", diagramtool.FormatPNG, true, "png"},
		{"drawio", diagramtool.FormatDrawio, true, "drawio"},
		{"draw.io", diagramtool.FormatDrawio, true, "drawio"},
		{"pdf", diagramtool.FormatSVG, false, "svg"},
	}
	for _, tt := range tests {
		got, ok := diagramtool.ParseFormat(tt.flag)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("ParseFormat(%q) = %v, %v; want %v, %v", tt.flag, got, ok, tt.want, tt.wantOK)
		}
		if got.String() != tt.wantStr {
			t.Errorf("ParseFormat(%q).String() = %q, want %q", tt.flag, got.String(), tt.wantStr)
		}
	}
}
