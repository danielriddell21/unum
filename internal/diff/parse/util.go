package parse

func objPath(parent, key string) string {
	if parent == "." {
		return "." + key
	}
	return parent + "." + key
}

type alignOp struct {
	a int
	b int
}

const alignCellLimit = 250_000

func alignByKey(aKeys, bKeys []string) []alignOp {
	if ops := alignReordered(aKeys, bKeys); ops != nil {
		return ops
	}
	if len(aKeys)*len(bKeys) > alignCellLimit {
		return positionalAlign(len(aKeys), len(bKeys))
	}

	table := lcsTable(aKeys, bKeys)
	ops := make([]alignOp, 0, max(len(aKeys), len(bKeys)))
	i, j := 0, 0
	for i < len(aKeys) && j < len(bKeys) {
		switch {
		case aKeys[i] == bKeys[j]:
			ops = append(ops, alignOp{a: i, b: j})
			i++
			j++
		case table[i+1][j] >= table[i][j+1]:
			ops = append(ops, alignOp{a: i, b: -1})
			i++
		default:
			ops = append(ops, alignOp{a: -1, b: j})
			j++
		}
	}
	for ; i < len(aKeys); i++ {
		ops = append(ops, alignOp{a: i, b: -1})
	}
	for ; j < len(bKeys); j++ {
		ops = append(ops, alignOp{a: -1, b: j})
	}
	return pairReplacements(ops)
}

func alignReordered(aKeys, bKeys []string) []alignOp {
	if len(aKeys) != len(bKeys) || len(aKeys) == 0 {
		return nil
	}

	pool := make(map[string][]int, len(aKeys))
	for i, k := range aKeys {
		pool[k] = append(pool[k], i)
	}
	ops := make([]alignOp, 0, len(bKeys))
	for j, k := range bKeys {
		free := pool[k]
		if len(free) == 0 {
			return nil
		}
		pool[k] = free[1:]
		ops = append(ops, alignOp{a: free[0], b: j})
	}
	return ops
}

func positionalAlign(aLen, bLen int) []alignOp {
	ops := make([]alignOp, 0, max(aLen, bLen))
	for i := range max(aLen, bLen) {
		switch {
		case i < aLen && i < bLen:
			ops = append(ops, alignOp{a: i, b: i})
		case i < bLen:
			ops = append(ops, alignOp{a: -1, b: i})
		default:
			ops = append(ops, alignOp{a: i, b: -1})
		}
	}
	return ops
}

func lcsTable(a, b []string) [][]int {
	table := make([][]int, len(a)+1)
	for i := range table {
		table[i] = make([]int, len(b)+1)
	}
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				table[i][j] = table[i+1][j+1] + 1
				continue
			}
			table[i][j] = max(table[i+1][j], table[i][j+1])
		}
	}
	return table
}

func pairReplacements(ops []alignOp) []alignOp {
	out := make([]alignOp, 0, len(ops))
	for i := 0; i < len(ops); {
		if ops[i].b >= 0 {
			out = append(out, ops[i])
			i++
			continue
		}
		// A run of removals followed directly by a run of additions is an
		// in-place replacement: pair them so the elements diff as modified.
		removeStart := i
		for i < len(ops) && ops[i].b < 0 {
			i++
		}
		addStart := i
		for i < len(ops) && ops[i].a < 0 {
			i++
		}
		out = append(out, mergeRuns(ops[removeStart:addStart], ops[addStart:i])...)
	}
	return out
}

func mergeRuns(removed, added []alignOp) []alignOp {
	paired := min(len(removed), len(added))
	out := make([]alignOp, 0, len(removed)+len(added))
	for i := range paired {
		out = append(out, alignOp{a: removed[i].a, b: added[i].b})
	}
	out = append(out, removed[paired:]...)
	return append(out, added[paired:]...)
}
