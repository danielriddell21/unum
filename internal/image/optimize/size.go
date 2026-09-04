package optimize

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var byteUnits = []struct {
	suffix string
	mult   float64
}{
	{"kib", 1 << 10},
	{"mib", 1 << 20},
	{"gib", 1 << 30},
	{"kb", 1e3},
	{"mb", 1e6},
	{"gb", 1e9},
	{"k", 1e3},
	{"m", 1e6},
	{"g", 1e9},
	{"b", 1},
}

var sizeUnits = []string{"kB", "MB", "GB", "TB", "PB"}

func HumanBytes(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d B", n)
	}
	v := float64(n)
	i := -1
	for v >= 1000 && i < len(sizeUnits)-1 {
		v /= 1000
		i++
	}
	// 999999 B divides to 1000.0 kB, which reads worse than 1.0 MB.
	if v >= 999.5 && i < len(sizeUnits)-1 {
		v /= 1000
		i++
	}
	if v < 10 {
		return fmt.Sprintf("%.1f %s", v, sizeUnits[i])
	}
	return fmt.Sprintf("%.0f %s", v, sizeUnits[i])
}

func ParseBytes(s string) (int, error) {
	t := strings.ToLower(strings.TrimSpace(s))
	if t == "" {
		return 0, fmt.Errorf("empty size")
	}

	mult := 1.0
	for _, u := range byteUnits {
		if rest, ok := strings.CutSuffix(t, u.suffix); ok {
			t = strings.TrimSpace(rest)
			mult = u.mult
			break
		}
	}

	v, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q (want e.g. 200kb, 1.5mb, 500000)", s)
	}
	if v <= 0 {
		return 0, fmt.Errorf("size must be positive, got %q", s)
	}

	n := math.Round(v * mult)
	if n > math.MaxInt32 {
		return 0, fmt.Errorf("size %q is too large", s)
	}
	return int(n), nil
}

func ParseScale(s string) (float64, error) {
	t := strings.TrimSpace(s)
	if t == "" {
		return 1, nil
	}

	pct := false
	if rest, ok := strings.CutSuffix(t, "%"); ok {
		t = strings.TrimSpace(rest)
		pct = true
	}

	v, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid scale %q (want e.g. 50%% or 0.5)", s)
	}
	if pct {
		v /= 100
	}
	if v <= 0 || v > 1 {
		return 0, fmt.Errorf("must be between 0 and 1 (or 1%% and 100%%), got %q", s)
	}
	return v, nil
}

func Saving(before, after int) float64 {
	if before <= 0 {
		return 0
	}
	return 1 - float64(after)/float64(before)
}
