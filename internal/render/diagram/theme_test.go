package diagram_test

import (
	"testing"

	"github.com/danielriddell21/unum/internal/render/diagram"
)

func TestThemeByName(t *testing.T) {
	tests := []struct {
		name     string
		wantDark bool
	}{
		{"cyber", true},
		{"dracula", true},
		{"clean", false},
		{"solarized", false},
		{"", true}, // unknown → default dark
	}
	for _, tt := range tests {
		got := diagram.ThemeByName(tt.name)
		if got.Dark != tt.wantDark {
			t.Errorf("ThemeByName(%q).Dark = %v, want %v", tt.name, got.Dark, tt.wantDark)
		}
		if got.Background == "" || got.Accent == "" {
			t.Errorf("ThemeByName(%q) missing colors: %+v", tt.name, got)
		}
	}
}
