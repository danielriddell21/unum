package stats

import (
	"context"
	"math"
	"sort"
	"strconv"

	"github.com/danielriddell21/unum/internal/json/analyze"
	"github.com/danielriddell21/unum/internal/json/node"
)

// Stats is the name of the stats lens annotations.
const Lens = "stats"

// Analyzer annotates numeric array nodes with statistical summaries.
type Analyzer struct{}

func (a Analyzer) Name() string { return Lens }

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
	return ctx.Err()
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

	sort.Float64s(nums)

	min := nums[0]
	max := nums[numCount-1]
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

	ann("min", min)
	ann("max", max)
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
