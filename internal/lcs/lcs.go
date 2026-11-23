package lcs

type Op struct {
	A int
	B int
}

type Result struct {
	Ops         []Op
	Approximate bool
}

const (
	cellLimit = 250_000
	workLimit = 1 << 25
)

func Align(a, b []string) []Op {
	return Compute(a, b).Ops
}

func Compute(a, b []string) Result {
	return compute(a, b, workLimit)
}

func compute(a, b []string, budget int) Result {
	// Small inputs keep the quadratic table: it is cheap at this size and its
	// tie-breaking is what existing callers have been built against.
	if table := buildTable(a, b); table != nil {
		return Result{Ops: walkTable(a, b, table)}
	}
	if ops, ok := alignLinear(a, b, budget); ok {
		return Result{Ops: ops}
	}
	return Result{Ops: positional(len(a), len(b)), Approximate: true}
}

func walkTable(a, b []string, table [][]int) []Op {
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

func alignLinear(a, b []string, budget int) ([]Op, bool) {
	ai, bi := intern(a, b)
	// A line that never occurs on the other side cannot be part of any common
	// subsequence, so dropping it before the search keeps the result exact while
	// shrinking heavily rewritten inputs to the lines that can actually match.
	aIdx, bIdx := shared(ai, bi), shared(bi, ai)
	aKept, bKept := pick(ai, aIdx), pick(bi, bIdx)

	off := (len(aKept)+len(bKept)+1)/2 + 1
	m := &myers{
		a:      aKept,
		b:      bKept,
		vf:     make([]int, 2*off+1),
		vb:     make([]int, 2*off+1),
		off:    off,
		budget: budget,
	}
	if !m.diff(span{aHi: len(aKept), bHi: len(bKept)}) {
		return nil, false
	}
	for i, mt := range m.matches {
		m.matches[i] = Op{A: aIdx[mt.A], B: bIdx[mt.B]}
	}
	return opsFromMatches(m.matches, len(a), len(b)), true
}

func intern(a, b []string) ([]int, []int) {
	ids := make(map[string]int, len(a))
	convert := func(lines []string) []int {
		out := make([]int, len(lines))
		for i, l := range lines {
			id, ok := ids[l]
			if !ok {
				id = len(ids)
				ids[l] = id
			}
			out[i] = id
		}
		return out
	}
	return convert(a), convert(b)
}

func shared(ids, other []int) []int {
	// intern numbers lines densely across both sides, so every id is below
	// the combined length.
	present := make([]bool, len(ids)+len(other))
	for _, id := range other {
		present[id] = true
	}
	idx := make([]int, 0, len(ids))
	for i, id := range ids {
		if present[id] {
			idx = append(idx, i)
		}
	}
	return idx
}

func pick(ids, idx []int) []int {
	out := make([]int, len(idx))
	for i, at := range idx {
		out[i] = ids[at]
	}
	return out
}

func opsFromMatches(matches []Op, aLen, bLen int) []Op {
	ops := make([]Op, 0, aLen+bLen-len(matches))
	i, j := 0, 0
	// Every gap between matches is emitted as its removals then its additions,
	// the same shape the table walk produces for an in-place replacement.
	gap := func(aEnd, bEnd int) {
		for ; i < aEnd; i++ {
			ops = append(ops, Op{A: i, B: -1})
		}
		for ; j < bEnd; j++ {
			ops = append(ops, Op{A: -1, B: j})
		}
	}
	for _, mt := range matches {
		gap(mt.A, mt.B)
		ops = append(ops, mt)
		i++
		j++
	}
	gap(aLen, bLen)
	return ops
}

type span struct {
	aLo, aHi, bLo, bHi int
}

func (s span) n() int { return s.aHi - s.aLo }
func (s span) m() int { return s.bHi - s.bLo }

type myers struct {
	a, b    []int
	vf, vb  []int
	off     int
	budget  int
	matches []Op
}

type frontier struct {
	v              []int
	kStart, kEnd   int
	reverse        bool
	checksOverlaps bool
}

func (m *myers) diff(s span) bool {
	for s.aLo < s.aHi && s.bLo < s.bHi && m.a[s.aLo] == m.b[s.bLo] {
		m.matches = append(m.matches, Op{A: s.aLo, B: s.bLo})
		s.aLo++
		s.bLo++
	}
	suffix := 0
	for s.aLo < s.aHi && s.bLo < s.bHi && m.a[s.aHi-1] == m.b[s.bHi-1] {
		s.aHi--
		s.bHi--
		suffix++
	}

	if s.n() > 0 && s.m() > 0 {
		x, y, ok := m.bisect(s)
		if !ok {
			return false
		}
		if !m.diff(span{aLo: s.aLo, aHi: x, bLo: s.bLo, bHi: y}) {
			return false
		}
		if !m.diff(span{aLo: x, aHi: s.aHi, bLo: y, bHi: s.bHi}) {
			return false
		}
	}

	for k := range suffix {
		m.matches = append(m.matches, Op{A: s.aHi + k, B: s.bHi + k})
	}
	return true
}

func (m *myers) bisect(s span) (int, int, bool) {
	// Myers' middle snake (1986, section 4b): walk forward from the start and
	// backward from the end until the frontiers meet, then split there. Each
	// half of the split is independently optimal, so recursing stays minimal.
	maxD := (s.n() + s.m() + 1) / 2
	for k := -maxD - 1; k <= maxD+1; k++ {
		m.vf[m.off+k] = -1
		m.vb[m.off+k] = -1
	}
	m.vf[m.off+1] = 0
	m.vb[m.off+1] = 0

	odd := (s.n()-s.m())%2 != 0
	fwd := &frontier{v: m.vf, checksOverlaps: odd}
	bwd := &frontier{v: m.vb, reverse: true, checksOverlaps: !odd}
	for d := 0; ; d++ {
		if m.budget < 0 {
			return 0, 0, false
		}
		if x, y, found := m.extend(s, d, maxD, fwd, bwd); found {
			return s.aLo + x, s.bLo + y, true
		}
		if x, y, found := m.extend(s, d, maxD, bwd, fwd); found {
			return s.aLo + x, s.bLo + y, true
		}
	}
}

func (m *myers) extend(s span, d, maxD int, f, other *frontier) (int, int, bool) {
	n, mm := s.n(), s.m()
	for k := -d + f.kStart; k <= d-f.kEnd; k += 2 {
		i := m.off + k
		x := f.v[i-1] + 1
		if k == -d || (k != d && f.v[i-1] < f.v[i+1]) {
			x = f.v[i+1]
		}
		start := x
		x = m.snake(s, x, x-k, f.reverse)
		y := x - k
		m.budget -= 1 + x - start
		f.v[i] = x

		switch {
		case x > n:
			f.kEnd += 2
		case y > mm:
			f.kStart += 2
		case f.checksOverlaps:
			if sx, sy, ok := m.overlap(s, k, x, y, maxD, f, other); ok {
				return sx, sy, true
			}
		}
	}
	return 0, 0, false
}

func (m *myers) snake(s span, x, y int, reverse bool) int {
	n, mm := s.n(), s.m()
	if reverse {
		for x < n && y < mm && m.a[s.aHi-1-x] == m.b[s.bHi-1-y] {
			x++
			y++
		}
		return x
	}
	for x < n && y < mm && m.a[s.aLo+x] == m.b[s.bLo+y] {
		x++
		y++
	}
	return x
}

func (m *myers) overlap(s span, k, x, y, maxD int, f, other *frontier) (int, int, bool) {
	kOther := s.n() - s.m() - k
	if kOther < -maxD-1 || kOther > maxD+1 {
		return 0, 0, false
	}
	xOther := other.v[m.off+kOther]
	// The two frontiers meet once their progress along the shared diagonal
	// covers the whole of A between them.
	if xOther == -1 || x+xOther < s.n() {
		return 0, 0, false
	}
	if f.reverse {
		return xOther, xOther - kOther, true
	}
	return x, y, true
}
