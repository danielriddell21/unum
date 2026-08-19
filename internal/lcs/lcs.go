package lcs

type Op struct {
	A int
	B int
}

const cellLimit = 250_000

func Align(a, b []string) []Op {
	table := buildTable(a, b)
	if table == nil {
		return positional(len(a), len(b))
	}

	ops := make([]Op, 0, max(len(a), len(b)))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			ops = append(ops, Op{A: i, B: j})
			i++
			j++
		case table[i+1][j] >= table[i][j+1]:
			ops = append(ops, Op{A: i, B: -1})
			i++
		default:
			ops = append(ops, Op{A: -1, B: j})
			j++
		}
	}
	for ; i < len(a); i++ {
		ops = append(ops, Op{A: i, B: -1})
	}
	for ; j < len(b); j++ {
		ops = append(ops, Op{A: -1, B: j})
	}
	return ops
}

func buildTable(a, b []string) [][]int {
	if len(a) > cellLimit || len(b) > cellLimit {
		return nil
	}
	if len(a) > 0 && len(b) > cellLimit/len(a) {
		return nil
	}

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

func positional(aLen, bLen int) []Op {
	ops := make([]Op, 0, max(aLen, bLen))
	for i := range max(aLen, bLen) {
		switch {
		case i < aLen && i < bLen:
			ops = append(ops, Op{A: i, B: i})
		case i < bLen:
			ops = append(ops, Op{A: -1, B: i})
		default:
			ops = append(ops, Op{A: i, B: -1})
		}
	}
	return ops
}
