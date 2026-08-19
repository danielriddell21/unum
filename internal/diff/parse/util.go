package parse

import "github.com/danielriddell21/unum/internal/lcs"

func objPath(parent, key string) string {
	if parent == "." {
		return "." + key
	}
	return parent + "." + key
}

func alignByKey(aKeys, bKeys []string) []lcs.Op {
	if ops := alignReordered(aKeys, bKeys); ops != nil {
		return ops
	}
	return pairReplacements(lcs.Align(aKeys, bKeys))
}

func alignReordered(aKeys, bKeys []string) []lcs.Op {
	if len(aKeys) != len(bKeys) || len(aKeys) == 0 {
		return nil
	}

	pool := make(map[string][]int, len(aKeys))
	for i, k := range aKeys {
		pool[k] = append(pool[k], i)
	}
	ops := make([]lcs.Op, 0, len(bKeys))
	for j, k := range bKeys {
		free := pool[k]
		if len(free) == 0 {
			return nil
		}
		pool[k] = free[1:]
		ops = append(ops, lcs.Op{A: free[0], B: j})
	}
	return ops
}

func pairReplacements(ops []lcs.Op) []lcs.Op {
	out := make([]lcs.Op, 0, len(ops))
	for i := 0; i < len(ops); {
		if ops[i].B >= 0 {
			out = append(out, ops[i])
			i++
			continue
		}
		// A run of removals followed directly by a run of additions is an
		// in-place replacement: pair them so the elements diff as modified.
		removeStart := i
		for i < len(ops) && ops[i].B < 0 {
			i++
		}
		addStart := i
		for i < len(ops) && ops[i].A < 0 {
			i++
		}
		out = append(out, mergeRuns(ops[removeStart:addStart], ops[addStart:i])...)
	}
	return out
}

func mergeRuns(removed, added []lcs.Op) []lcs.Op {
	paired := min(len(removed), len(added))
	out := make([]lcs.Op, 0, len(removed)+len(added))
	for i := range paired {
		out = append(out, lcs.Op{A: removed[i].A, B: added[i].B})
	}
	out = append(out, removed[paired:]...)
	return append(out, added[paired:]...)
}
