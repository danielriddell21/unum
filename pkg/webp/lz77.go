package webp

const (
	minMatch = 3
	maxMatch = 4096
	hashBits = 14
	hashSize = 1 << hashBits
	maxChain = 32
)

// matcher finds LZ77 backward references over a pixel buffer using a hash
// chain, in the manner of deflate: pixels are indexed by a hash of the next
// three, and each new position is pushed onto that bucket's chain.
type matcher struct {
	argb []uint32
	head [hashSize]int32
	prev []int32
}

func newMatcher(argb []uint32) *matcher {
	m := &matcher{argb: argb, prev: make([]int32, len(argb))}
	for i := range m.head {
		m.head[i] = -1
	}
	for i := range m.prev {
		m.prev[i] = -1
	}
	return m
}

func (m *matcher) hash(i int) uint32 {
	h := m.argb[i]*2654435761 ^ m.argb[i+1]*2246822519 ^ m.argb[i+2]*3266489917
	return h >> (32 - hashBits) & (hashSize - 1)
}

// insert records position i as a candidate for future matches.
func (m *matcher) insert(i int) {
	if i+minMatch > len(m.argb) {
		return
	}
	h := m.hash(i)
	m.prev[i] = m.head[h]
	m.head[h] = int32(i)
}

// find returns the longest match for the pixels at i, or a length below
// minMatch if there is nothing worth referencing.
func (m *matcher) find(i int) (length, dist int) {
	if i+minMatch > len(m.argb) {
		return 0, 0
	}

	limit := min(len(m.argb)-i, maxMatch)
	best, bestDist := 0, 0

	for cand, chain := m.head[m.hash(i)], 0; cand >= 0 && chain < maxChain; chain++ {
		p := int(cand)
		cand = m.prev[p]

		n := 0
		for n < limit && m.argb[p+n] == m.argb[i+n] {
			n++
		}
		if n > best {
			best, bestDist = n, i-p
			if best == limit {
				break
			}
		}
	}

	if best < minMatch {
		return 0, 0
	}
	return best, bestDist
}
