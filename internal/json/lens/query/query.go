package query

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"

	"github.com/danielriddell21/unum/internal/json/node"
)

const MaxOutputBytes = 10 << 20

var (
	ErrTimeout         = errors.New("query timed out")
	ErrOutputTruncated = errors.New("output truncated")
)

func Execute(ctx context.Context, root *node.Node, expr string) (string, error) {
	return ExecuteAny(ctx, root.ToAny(), expr)
}

func ExecuteAny(ctx context.Context, input any, expr string) (string, error) {
	if strings.TrimSpace(expr) == "" {
		return "", fmt.Errorf("empty jq expression")
	}

	q, err := gojq.Parse(expr)
	if err != nil {
		return "", fmt.Errorf("jq parse: %w", err)
	}

	iter := q.RunWithContext(ctx, input)

	var out strings.Builder
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if errors.Is(err, context.DeadlineExceeded) {
				return "", ErrTimeout
			}
			return "", fmt.Errorf("jq: %w", err)
		}
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", fmt.Errorf("jq marshal: %w", err)
		}
		sep := 0
		if out.Len() > 0 {
			sep = 1
		}
		if out.Len()+sep+len(b) > MaxOutputBytes {
			return "", fmt.Errorf("%w at %d bytes", ErrOutputTruncated, MaxOutputBytes)
		}
		if sep > 0 {
			out.WriteByte('\n')
		}
		out.Write(b)
	}

	return out.String(), nil
}
