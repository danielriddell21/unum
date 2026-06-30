package query

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"

	"github.com/danielriddell21/unum/internal/json/node"
)

func Execute(root *node.Node, expr string) (string, error) {
	if strings.TrimSpace(expr) == "" {
		return "", fmt.Errorf("empty jq expression")
	}

	q, err := gojq.Parse(expr)
	if err != nil {
		return "", fmt.Errorf("jq parse: %w", err)
	}

	// Convert node tree to Go any (gojq operates on any)
	input := nodeToAny(root)

	iter := q.Run(input)

	var results []string
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return "", fmt.Errorf("jq: %w", err)
		}
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", fmt.Errorf("jq marshal: %w", err)
		}
		results = append(results, string(b))
	}

	return strings.Join(results, "\n"), nil
}

func nodeToAny(n *node.Node) any {
	switch n.Kind {
	case node.KindNull:
		return nil
	case node.KindBool:
		return n.Raw == "true"
	case node.KindNumber:
		f := json.Number(n.Raw)
		// gojq handles json.Number
		return f
	case node.KindString:
		var s string
		_ = json.Unmarshal([]byte(n.Raw), &s)
		return s
	case node.KindArray:
		arr := make([]any, len(n.Children))
		for i, c := range n.Children {
			arr[i] = nodeToAny(c)
		}
		return arr
	case node.KindObject:
		m := make(map[string]any, len(n.Children))
		for _, c := range n.Children {
			m[c.Key] = nodeToAny(c)
		}
		return m
	}
	return nil
}
