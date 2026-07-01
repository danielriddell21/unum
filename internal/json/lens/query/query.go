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
	input := root.ToAny()

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
