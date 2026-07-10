package rendertool_test

import (
	"testing"

	rendertool "github.com/danielriddell21/unum/internal/render"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		file string
		want rendertool.Language
	}{
		{"diagram.mmd", rendertool.LangMermaid},
		{"diagram.mermaid", rendertool.LangMermaid},
		{"flow.D2", rendertool.LangD2},
		{"graph.d2", rendertool.LangD2},
		{"notes.txt", rendertool.LangUnknown},
		{"", rendertool.LangUnknown},
	}
	for _, tt := range tests {
		if got := rendertool.DetectLanguage(tt.file); got != tt.want {
			t.Errorf("DetectLanguage(%q) = %v, want %v", tt.file, got, tt.want)
		}
	}
}

func TestParseLanguage(t *testing.T) {
	tests := []struct {
		flag, file string
		want       rendertool.Language
	}{
		{"mermaid", "graph.d2", rendertool.LangMermaid},
		{"mmd", "graph.d2", rendertool.LangMermaid},
		{"d2", "diagram.mmd", rendertool.LangD2},
		{"", "diagram.mmd", rendertool.LangMermaid},
		{"", "graph.d2", rendertool.LangD2},
		{"", "notes.txt", rendertool.LangUnknown},
	}
	for _, tt := range tests {
		if got := rendertool.ParseLanguage(tt.flag, tt.file); got != tt.want {
			t.Errorf("ParseLanguage(%q, %q) = %v, want %v", tt.flag, tt.file, got, tt.want)
		}
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		flag    string
		want    rendertool.Format
		wantOK  bool
		wantStr string
	}{
		{"", rendertool.FormatSVG, true, "svg"},
		{"svg", rendertool.FormatSVG, true, "svg"},
		{"PNG", rendertool.FormatPNG, true, "png"},
		{"drawio", rendertool.FormatDrawio, true, "drawio"},
		{"draw.io", rendertool.FormatDrawio, true, "drawio"},
		{"pdf", rendertool.FormatSVG, false, "svg"},
	}
	for _, tt := range tests {
		got, ok := rendertool.ParseFormat(tt.flag)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("ParseFormat(%q) = %v, %v; want %v, %v", tt.flag, got, ok, tt.want, tt.wantOK)
		}
		if got.String() != tt.wantStr {
			t.Errorf("ParseFormat(%q).String() = %q, want %q", tt.flag, got.String(), tt.wantStr)
		}
	}
}
