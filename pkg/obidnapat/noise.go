package obidnapat

import "bytes"

// A window is not reported when it holds no more real bases (not N) than
// errors: a pattern must not be matched by its errors alone. The wildcards of
// the text are free in the kernels (mask['n'] accepts everything), so a run of N
// would otherwise give one occurrence per position.
//
// The rule is applied to every occurrence as the kernels produce it, before any
// grouping by overlap, so that a rejected window cannot hide its neighbours. It
// costs nothing for a sequence without N (hasN is one vectorised scan), and,
// for the other ones, the text is cut in chunks of noiseChunk bytes after each
// of which the rejected windows are dropped in place: the buffer of the
// occurrences never holds more than a chunk of them.

// needsCheck tells whether the rule can reject a window of c: when the sequence
// holds an N, when the pattern allows at least as many errors as it has
// positions (then a window of real bases can be all errors), or when the
// pattern allows indels (the nominal window of a match with deletions can hang
// over the start of the sequence, and its missing bases are not real ones).
func (c *compiled) needsCheck(noisy bool) bool {
	return noisy || c.indel || c.levels >= len(c.positions)
}

// noiseChunk is the number of bytes scanned between two filterings of the
// occurrences, for a sequence that holds N.
const noiseChunk = 4096

// hasN tells whether seq holds an N, in either case.
func hasN(seq []byte) bool {
	return bytes.IndexByte(seq, 'n') >= 0 || bytes.IndexByte(seq, 'N') >= 0
}

// realBases counts the bases of seq[start:end] that are not N. The window is
// read circularly when circular (end may go past len(seq)); otherwise the part
// that falls outside the sequence (a nominal window with deletions) is ignored.
func realBases(seq []byte, start, end int, circular bool) int {
	n := len(seq)
	if circular {
		count := 0
		for i := start; i < end; i++ {
			if b := seq[((i%n)+n)%n]; b != 'n' && b != 'N' {
				count++
			}
		}
		return count
	}
	count := 0
	for _, b := range seq[max(start, 0):min(end, n)] {
		if b != 'n' && b != 'N' {
			count++
		}
	}
	return count
}

// dropNoise removes, among the windows dst[mark:], those that hold no more real
// bases than errors. The windows are [3]int{start, end, errors}.
func dropNoise(dst [][3]int, mark int, seq []byte, circular bool) [][3]int {
	k := mark
	for _, w := range dst[mark:] {
		if realBases(seq, w[0], w[1], circular) > w[2] {
			dst[k] = w
			k++
		}
	}
	return dst[:k]
}
