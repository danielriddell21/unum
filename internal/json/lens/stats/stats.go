package stats

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strconv"

	"github.com/danielriddell21/unum/internal/json/analyze"
	"github.com/danielriddell21/unum/internal/json/node"
)

const Lens = "stats"

type Stats struct {
	Count        int
	NumericCount int
	Min          float64
	Max          float64
	Mean         float64
	Stddev       float64
	P50          float64
	P95          float64
	P99          float64
}

func Get(n *node.Node) (Stats, bool) {
	nc, ok := n.GetAnnotation(Lens, "numeric_count")
	if !ok {
		return Stats{}, false
	}
	s := Stats{NumericCount: nc.(int)}
	if v, ok := n.GetAnnotation(Lens, "count"); ok {
		s.Count = v.(int)
	}
	getF := func(name string) float64 {
		if v, ok := n.GetAnnotation(Lens, name); ok {
			return v.(float64)
		}
		return 0
	}
	s.Min = getF("min")
	s.Max = getF("max")
	s.Mean = getF("mean")
	s.Stddev = getF("stddev")
	s.P50 = getF("p50")
	s.P95 = getF("p95")
	s.P99 = getF("p99")
	return s, true
}

type Analyzer struct{}

func (a Analyzer) Name() string { return Lens }

func (a Analyzer) Enabled(opts analyze.Options) bool { return opts.RunStats }

func (a Analyzer) Run(ctx context.Context, root *node.Node, _ analyze.Options) error {
	root.Walk(func(n *node.Node) bool {
		if ctx.Err() != nil {
			return false
		}
		if n.Kind == node.KindArray {
			annotateArray(n)
		}
		return true
	})
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}
	return nil
}

func annotateArray(n *node.Node) {
	var nums []float64
	for _, c := range n.Children {
		if c.Kind == node.KindNumber {
			f, err := strconv.ParseFloat(c.Raw, 64)
			if err == nil {
				nums = append(nums, f)
			}
		}
	}

	totalCount := len(n.Children)
	numCount := len(nums)

	ann := func(name string, val any) {
		n.Annotate(node.Annotation{
			Key:   node.AnnotationKey{Lens: Lens, Name: name},
			Value: val,
		})
	}

	ann("count", totalCount)
	ann("numeric_count", numCount)

	if numCount == 0 {
		return
	}

	slices.Sort(nums)

	minV := nums[0]
	maxV := nums[numCount-1]
	sum := 0.0
	for _, v := range nums {
		sum += v
	}
	mean := sum / float64(numCount)

	variance := 0.0
	for _, v := range nums {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(numCount)
	stddev := math.Sqrt(variance)

	ann("min", minV)
	ann("max", maxV)
	ann("mean", mean)
	ann("stddev", stddev)
	ann("p50", percentile(nums, 50))
	ann("p95", percentile(nums, 95))
	ann("p99", percentile(nums, 99))
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	rank := p / 100 * float64(len(sorted)-1)
	lo := int(rank)
	hi := lo + 1
	if hi >= len(sorted) {
		return sorted[len(sorted)-1]
	}
	frac := rank - float64(lo)
	return sorted[lo] + frac*(sorted[hi]-sorted[lo])
}
