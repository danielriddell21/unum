package optimize

import (
	"math"
	"testing"
)

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want string
	}{
		{"zero", 0, "0 B"},
		{"bytes", 512, "512 B"},
		{"just under a kilobyte", 999, "999 B"},
		{"exactly a kilobyte", 1000, "1.0 kB"},
		{"small kilobytes keep a decimal", 2400, "2.4 kB"},
		{"large kilobytes drop the decimal", 82000, "82 kB"},
		{"rounds up into megabytes", 999999, "1.0 MB"},
		{"megabytes", 2400000, "2.4 MB"},
		{"gigabytes", 3200000000, "3.2 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HumanBytes(tt.in); got != tt.want {
				t.Errorf("HumanBytes(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseBytes(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    int
		wantErr bool
	}{
		{"plain number", "500000", 500000, false},
		{"kilobytes", "200kb", 200000, false},
		{"short kilobytes", "200k", 200000, false},
		{"megabytes", "1.5mb", 1500000, false},
		{"kibibytes", "1kib", 1024, false},
		{"mebibytes", "2mib", 2097152, false},
		{"bytes suffix", "900b", 900, false},
		{"uppercase", "200KB", 200000, false},
		{"surrounding space", "  200 kb  ", 200000, false},
		{"empty", "", 0, true},
		{"not a number", "big", 0, true},
		{"zero", "0", 0, true},
		{"negative", "-5kb", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBytes(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseBytes(%q) = %d, want error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseBytes(%q): unexpected error %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ParseBytes(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseScale(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    float64
		wantErr bool
	}{
		{"empty means unchanged", "", 1, false},
		{"percentage", "50%", 0.5, false},
		{"fraction", "0.5", 0.5, false},
		{"full size as percentage", "100%", 1, false},
		{"full size as fraction", "1", 1, false},
		{"percentage with space", " 25 % ", 0.25, false},
		{"above one rejected", "2", 0, true},
		{"above 100 percent rejected", "150%", 0, true},
		{"zero rejected", "0", 0, true},
		{"negative rejected", "-0.5", 0, true},
		{"not a number", "half", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseScale(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseScale(%q) = %v, want error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseScale(%q): unexpected error %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ParseScale(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestSaving(t *testing.T) {
	tests := []struct {
		name          string
		before, after int
		want          float64
	}{
		{"halved", 1000, 500, 0.5},
		{"unchanged", 1000, 1000, 0},
		{"grew", 1000, 1200, -0.2},
		{"no source bytes", 0, 500, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Saving(tt.before, tt.after); math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("Saving(%d, %d) = %v, want %v", tt.before, tt.after, got, tt.want)
			}
		})
	}
}
