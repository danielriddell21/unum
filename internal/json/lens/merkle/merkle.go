package merkle

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/danielriddell21/unum/internal/json/analyze"
	"github.com/danielriddell21/unum/internal/json/node"
)

const Lens = "merkle"

type Analyzer struct{}

func (a Analyzer) Name() string { return Lens }

func (a Analyzer) Run(ctx context.Context, root *node.Node, _ analyze.Options) error {
	root.WalkPost(func(n *node.Node) {
		if ctx.Err() != nil {
			return
		}
		h := hashNode(n)
		n.Annotate(node.Annotation{
			Key:   node.AnnotationKey{Lens: Lens, Name: "hash"},
			Value: h,
		})
	})
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}
	return nil
}

func RootHash(root *node.Node) string {
	if v, ok := root.GetAnnotation(Lens, "hash"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func hashNode(n *node.Node) string {
	h := sha256.New()

	switch n.Kind {
	case node.KindNull, node.KindBool, node.KindNumber, node.KindString:
		// Hash the type + raw value for precise fingerprinting
		_, _ = fmt.Fprintf(h, "%d:%s", n.Kind, n.Raw)

	case node.KindArray:
		// Ordered: hash of each child's hash in sequence
		_, _ = fmt.Fprintf(h, "array:")
		for _, c := range n.Children {
			if ch, ok := c.GetAnnotation(Lens, "hash"); ok {
				_, _ = fmt.Fprintf(h, "%s,", ch)
			}
		}

	case node.KindObject:
		// Sorted by key for determinism regardless of insertion order
		type kh struct {
			key  string
			hash string
		}
		pairs := make([]kh, 0, len(n.Children))
		for _, c := range n.Children {
			if ch, ok := c.GetAnnotation(Lens, "hash"); ok {
				pairs = append(pairs, kh{c.Key, ch.(string)})
			}
		}
		sort.Slice(pairs, func(i, j int) bool { return pairs[i].key < pairs[j].key })
		_, _ = fmt.Fprintf(h, "object:")
		for _, p := range pairs {
			_, _ = fmt.Fprintf(h, "%s=%s,", p.key, p.hash)
		}
	}

	return hex.EncodeToString(h.Sum(nil))
}

func Short(hash string) string {
	if len(hash) > 8 {
		return hash[:8]
	}
	return hash
}
