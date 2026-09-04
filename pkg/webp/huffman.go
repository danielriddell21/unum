package webp

import "sort"

// maxCodeLength is the longest Huffman code VP8L permits.
const maxCodeLength = 15

// huffman is a canonical Huffman code over a fixed alphabet.
//
// A symbol with length 0 is absent from the code. The single-symbol case is
// represented with length 0 as well, matching the decoder's "simple code"
// handling, where the symbol costs no bits at all.
type huffman struct {
	lengths []uint8
	codes   []uint32
	// symbols present in the code, ascending.
	present []int
}

// buildHuffman produces a canonical code whose lengths never exceed limit.
func buildHuffman(freq []int, limit int) huffman {
	h := huffman{
		lengths: make([]uint8, len(freq)),
		codes:   make([]uint32, len(freq)),
	}
	for sym, f := range freq {
		if f > 0 {
			h.present = append(h.present, sym)
		}
	}

	switch len(h.present) {
	case 0:
		// The decoder rejects an empty code, so keep symbol 0 as a
		// zero-bit placeholder.
		h.present = []int{0}
		return h
	case 1:
		// One symbol needs no bits to distinguish it.
		return h
	}

	h.lengths = packageMerge(freq, limit)
	h.assignCodes()
	return h
}

// assignCodes fills in canonical codes, matching the decoder's
// codeLengthsToCodes: shorter lengths first, ascending symbol order within a
// length.
func (h *huffman) assignCodes() {
	var histogram [maxCodeLength + 1]uint32
	for _, l := range h.lengths {
		histogram[l]++
	}

	var code uint32
	var next [maxCodeLength + 1]uint32
	for l := 1; l <= maxCodeLength; l++ {
		code = (code + histogram[l-1]) << 1
		next[l] = code
	}

	for sym, l := range h.lengths {
		if l > 0 {
			h.codes[sym] = next[l]
			next[l]++
		}
	}
}

// pmNode is a coin in the package-merge construction: either a leaf standing
// for one symbol, or a package of two cheaper items.
type pmNode struct {
	weight      int
	sym         int
	left, right int
}

// packageMerge computes optimal Huffman code lengths subject to a maximum
// length, using the package-merge algorithm. A plain Huffman tree can exceed
// the 15-bit ceiling VP8L imposes on skewed distributions, which the decoder
// rejects outright.
func packageMerge(freq []int, limit int) []uint8 {
	lengths := make([]uint8, len(freq))

	type leaf struct{ weight, sym int }
	var leaves []leaf
	for sym, f := range freq {
		if f > 0 {
			leaves = append(leaves, leaf{f, sym})
		}
	}
	if len(leaves) < 2 {
		return lengths
	}
	sort.Slice(leaves, func(i, j int) bool {
		if leaves[i].weight != leaves[j].weight {
			return leaves[i].weight < leaves[j].weight
		}
		return leaves[i].sym < leaves[j].sym
	})

	pool := make([]pmNode, 0, len(leaves)*2*limit)
	base := make([]int, len(leaves))
	for i, l := range leaves {
		pool = append(pool, pmNode{weight: l.weight, sym: l.sym, left: -1, right: -1})
		base[i] = len(pool) - 1
	}

	prev := base
	for level := 1; level < limit; level++ {
		packaged := make([]int, 0, len(prev)/2)
		for i := 0; i+1 < len(prev); i += 2 {
			pool = append(pool, pmNode{
				weight: pool[prev[i]].weight + pool[prev[i+1]].weight,
				sym:    -1,
				left:   prev[i],
				right:  prev[i+1],
			})
			packaged = append(packaged, len(pool)-1)
		}
		prev = mergeByWeight(pool, base, packaged)
	}

	// The first 2n-2 items of the final list select each symbol once per level
	// it survives to, which is exactly its code length.
	for _, idx := range prev[:2*len(leaves)-2] {
		countLeaves(pool, idx, lengths)
	}
	return lengths
}

// mergeByWeight merges two weight-ordered lists into one.
func mergeByWeight(pool []pmNode, a, b []int) []int {
	out := make([]int, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if pool[a[i]].weight <= pool[b[j]].weight {
			out = append(out, a[i])
			i++
			continue
		}
		out = append(out, b[j])
		j++
	}
	return append(append(out, a[i:]...), b[j:]...)
}

// countLeaves increments the code length of every symbol beneath idx.
func countLeaves(pool []pmNode, idx int, lengths []uint8) {
	stack := []int{idx}
	for len(stack) > 0 {
		n := pool[stack[len(stack)-1]]
		stack = stack[:len(stack)-1]
		if n.sym >= 0 {
			lengths[n.sym]++
			continue
		}
		stack = append(stack, n.left, n.right)
	}
}
