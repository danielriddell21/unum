package webp

// clToken is one entry in the run-length encoding of a Huffman code's own
// code lengths, which the bit-stream carries as a nested Huffman code.
type clToken struct {
	sym       int
	extra     uint32
	extraBits uint
}

// codeLengthRLE compresses a run of code lengths using the specification's
// repeat codes: 16 repeats the previous non-zero length, 17 and 18 repeat runs
// of zeroes.
func codeLengthRLE(lengths []uint8) []clToken {
	var out []clToken

	for i := 0; i < len(lengths); {
		l := lengths[i]
		run := 1
		for i+run < len(lengths) && lengths[i+run] == l {
			run++
		}
		i += run

		if l == 0 {
			out = appendZeroRun(out, run)
			continue
		}
		out = append(out, clToken{sym: int(l)})
		out = appendRepeatRun(out, l, run-1)
	}
	return out
}

func appendZeroRun(out []clToken, run int) []clToken {
	for run > 0 {
		switch {
		case run < 3:
			for range run {
				out = append(out, clToken{sym: 0})
			}
			run = 0
		case run <= 10:
			out = append(out, clToken{sym: 17, extra: uint32(run - 3), extraBits: 3})
			run = 0
		default:
			n := min(run, 138)
			out = append(out, clToken{sym: 18, extra: uint32(n - 11), extraBits: 7})
			run -= n
		}
	}
	return out
}

func appendRepeatRun(out []clToken, l uint8, run int) []clToken {
	for run > 0 {
		if run < 3 {
			for range run {
				out = append(out, clToken{sym: int(l)})
			}
			return out
		}
		n := min(run, 6)
		out = append(out, clToken{sym: 16, extra: uint32(n - 3), extraBits: 2})
		run -= n
	}
	return out
}

// canWriteSimple reports whether a code fits the bit-stream's compact form for
// one or two symbols, which needs no code-length table at all.
func canWriteSimple(h huffman) bool {
	if len(h.present) > 2 {
		return false
	}
	for _, s := range h.present {
		if s > 255 {
			return false
		}
	}
	return true
}

func writeHuffmanTree(w *bitWriter, h huffman) {
	if canWriteSimple(h) {
		w.writeBits(1, 1)
		w.writeBits(uint32(len(h.present)-1), 1)

		if first := h.present[0]; first < 2 {
			w.writeBits(0, 1)
			w.writeBits(uint32(first), 1)
		} else {
			w.writeBits(1, 1)
			w.writeBits(uint32(first), 8)
		}
		if len(h.present) == 2 {
			w.writeBits(uint32(h.present[1]), 8)
		}
		return
	}
	writeNormalTree(w, h)
}

// declaredLengths returns the code lengths as they must appear in the
// bit-stream. A code with a single symbol spends no bits on it, but the
// decoder rejects a table where every length is zero, so that one symbol is
// declared as length 1 even though nothing is emitted for it.
func declaredLengths(h huffman) []uint8 {
	out := make([]uint8, len(h.lengths))
	copy(out, h.lengths)
	if len(h.present) == 1 {
		out[h.present[0]] = 1
	}
	return out
}

func writeNormalTree(w *bitWriter, h huffman) {
	w.writeBits(0, 1)

	rle := codeLengthRLE(declaredLengths(h))
	clFreq := make([]int, len(codeLengthCodeOrder))
	for _, r := range rle {
		clFreq[r.sym]++
	}
	// Each code-length code length is written in three bits, so seven is the
	// most the format can express.
	clHuff := buildHuffman(clFreq, 7)
	clDeclared := declaredLengths(clHuff)

	nCodes := len(codeLengthCodeOrder)
	for nCodes > 4 && clDeclared[codeLengthCodeOrder[nCodes-1]] == 0 {
		nCodes--
	}
	w.writeBits(uint32(nCodes-4), 4)
	for i := range nCodes {
		w.writeBits(uint32(clDeclared[codeLengthCodeOrder[i]]), 3)
	}
	w.writeBits(0, 1) // read code lengths for the whole alphabet

	for _, r := range rle {
		w.writeCode(clHuff.codes[r.sym], uint(clHuff.lengths[r.sym]))
		w.writeBits(r.extra, r.extraBits)
	}
}

func writeSymbol(w *bitWriter, h huffman, sym int) {
	w.writeCode(h.codes[sym], uint(h.lengths[sym]))
}

// tokenize turns the pixel buffer into literals, backward references and
// colour-cache hits.
func tokenize(argb []uint32, ccBits uint) []token {
	var cache []uint32
	if ccBits > 0 {
		cache = make([]uint32, 1<<ccBits)
	}
	shift := 32 - ccBits

	// Every emitted pixel enters the cache, copies included — the decoder
	// catches its own cache up the same way before each lookup.
	remember := func(p uint32) {
		if cache != nil {
			cache[p*colorCacheMultiplier>>shift] = p
		}
	}

	m := newMatcher(argb)
	tokens := make([]token, 0, len(argb))

	for i := 0; i < len(argb); {
		if length, dist := m.find(i); length >= minMatch {
			tokens = append(tokens, token{kind: tokenCopy, length: length, dist: dist})
			for j := range length {
				m.insert(i + j)
				remember(argb[i+j])
			}
			i += length
			continue
		}

		p := argb[i]
		if cache != nil {
			if idx := p * colorCacheMultiplier >> shift; cache[idx] == p {
				tokens = append(tokens, token{kind: tokenCache, cache: idx})
				m.insert(i)
				remember(p)
				i++
				continue
			}
		}

		tokens = append(tokens, token{kind: tokenLiteral, argb: p})
		m.insert(i)
		remember(p)
		i++
	}
	return tokens
}

type histogramSet struct {
	green, red, blue, alpha, dist []int
}

func newHistograms(tokens []token, ccBits uint, dc distanceCodes) histogramSet {
	greenSize := nLiteralCodes + nLengthCodes
	if ccBits > 0 {
		greenSize += 1 << ccBits
	}

	h := histogramSet{
		green: make([]int, greenSize),
		red:   make([]int, nLiteralCodes),
		blue:  make([]int, nLiteralCodes),
		alpha: make([]int, nLiteralCodes),
		dist:  make([]int, nDistanceCodes),
	}

	for _, t := range tokens {
		switch t.kind {
		case tokenLiteral:
			h.green[t.argb>>8&0xff]++
			h.red[t.argb>>16&0xff]++
			h.blue[t.argb&0xff]++
			h.alpha[t.argb>>24&0xff]++
		case tokenCopy:
			lsym, _, _ := prefixEncode(t.length)
			h.green[nLiteralCodes+lsym]++
			dsym, _, _ := prefixEncode(dc.code(t.dist))
			h.dist[dsym]++
		case tokenCache:
			h.green[nLiteralCodes+nLengthCodes+int(t.cache)]++
		}
	}
	return h
}
