package node

import "testing"

func TestFormat_String(t *testing.T) {
	tests := []struct {
		name string
		f    Format
		want string
	}{
		{"text default", FormatText, "text"},
		{"json", FormatJSON, "json"},
		{"yaml", FormatYAML, "yaml"},
		{"terraform", FormatTerraform, "terraform"},
		{"unknown falls back to text", Format(99), "text"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.f.String(); got != tc.want {
				t.Errorf("Format(%d).String()=%q, want %q", tc.f, got, tc.want)
			}
		})
	}
}
