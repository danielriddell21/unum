package stats_test

import (
	"context"
	"testing"

	"github.com/danielriddell21/unum/internal/json/analyze"
	"github.com/danielriddell21/unum/internal/json/lens/stats"
	"github.com/danielriddell21/unum/internal/json/parse"
)

func TestStatsMinMaxMean(t *testing.T) {
	root, err := parse.Parse([]byte(`{"scores": [10, 20, 30]}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := (stats.Analyzer{}).Run(context.Background(), root, analyze.Options{}); err != nil {
		t.Fatal(err)
	}

	arr := root.Children[0]

	getFloat := func(name string) (float64, bool) {
		v, ok := arr.GetAnnotation(stats.Lens, name)
		if !ok {
			return 0, false
		}
		f, ok := v.(float64)
		return f, ok
	}

	if v, ok := getFloat("min"); !ok || v != 10 {
		t.Errorf("min: got %v, want 10", v)
	}
	if v, ok := getFloat("max"); !ok || v != 30 {
		t.Errorf("max: got %v, want 30", v)
	}
	if v, ok := getFloat("mean"); !ok || v != 20 {
		t.Errorf("mean: got %v, want 20", v)
	}
}

func TestStatsCount(t *testing.T) {
	root, err := parse.Parse([]byte(`[1, "hello", 3, null, 5]`))
	if err != nil {
		t.Fatal(err)
	}
	if err := (stats.Analyzer{}).Run(context.Background(), root, analyze.Options{}); err != nil {
		t.Fatal(err)
	}

	total, _ := root.GetAnnotation(stats.Lens, "count")
	numCount, _ := root.GetAnnotation(stats.Lens, "numeric_count")

	if total.(int) != 5 {
		t.Errorf("count: got %v, want 5", total)
	}
	if numCount.(int) != 3 {
		t.Errorf("numeric_count: got %v, want 3", numCount)
	}
}

func TestStatsEmptyArray(t *testing.T) {
	root, err := parse.Parse([]byte(`[]`))
	if err != nil {
		t.Fatal(err)
	}
	if err := (stats.Analyzer{}).Run(context.Background(), root, analyze.Options{}); err != nil {
		t.Fatal(err)
	}

	if _, ok := root.GetAnnotation(stats.Lens, "min"); ok {
		t.Error("expected no min annotation on empty array")
	}
}

func TestStatsNonNumericArray(t *testing.T) {
	root, err := parse.Parse([]byte(`["a", "b", "c"]`))
	if err != nil {
		t.Fatal(err)
	}
	if err := (stats.Analyzer{}).Run(context.Background(), root, analyze.Options{}); err != nil {
		t.Fatal(err)
	}

	nc, _ := root.GetAnnotation(stats.Lens, "numeric_count")
	if nc.(int) != 0 {
		t.Errorf("numeric_count: got %v, want 0", nc)
	}
	if _, ok := root.GetAnnotation(stats.Lens, "min"); ok {
		t.Error("expected no min annotation for non-numeric array")
	}
}

