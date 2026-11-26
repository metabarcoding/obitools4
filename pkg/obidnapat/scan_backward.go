package obidnapat

// The forward search of scan_indel.go knows where a window with errors ends
// (exactly) and how many errors it has, but not where it starts: its start is
// the nominal one. The start is found here without any alignment matrix, by the
// same bit-parallel search run backwards from the end, on the window of at most
// len+errors bases that precedes it.
//
// Bit j of the level k says that the pattern from its position j to its last one
// (bit len: nothing of it) is aligned with the bases read so far, which end at the
// end of the window, with at most k errors. A base read gives, from the old states
// o and the new ones n:
//
//	n_0 = (o_0 >> 1) & mask[b]
//	n_k = ((o_k >> 1) & mask[b])
//	    | ( (o_{k-1} >> 1)         // substitution
//	      | (n_{k-1} >> 1) )       // deletion of position j, nothing read
//	      & allow
//	    | (o_{k-1} & gap)          // insertion between positions j-1 and j
//
// where gap = allow << 1: an error after a position is forbidden if it is marked
// with #, and nothing is inserted before the first position (the first base read
// is then the start of the window). The initial state holds the deletions of the
// last positions. The start is the first base, going left, after which bit 0 of
// the level of the window is set: the latest start of an alignment of the whole
// pattern with that many errors.
//
// The forward search lets the first positions of the pattern be deleted at the
// start of the region without regard to #; this one does not, so for a pattern
// with # in its first positions it may find no start (ok false) at a window the
// forward search reported.

// backwardStart looks for the start of the window of the pattern that ends with
// the base seq[pos], with e errors, no further left than lo. Several starts can
// give an alignment with that many errors (an error at the edge of the window may
// be a deletion or a substitution): it returns the latest one, the shortest window.
func (c *compiled) backwardStart(seq []byte, pos, e, lo int) (start int, ok bool) {
	m := len(c.positions)
	lowest := max(lo, pos-(m+e)+1, 0)
	empty, gap, allow := c.end<<1, c.allow<<1, c.allow

	var r state
	r[0] = empty
	for k := 1; k <= e; k++ {
		r[k] = r[k-1] | ((r[k-1] >> 1) & allow)
	}
	for t := pos; t >= lowest; t-- {
		mk := c.mask[seq[t]]
		oldPrev := r[0]
		newPrev := (oldPrev >> 1) & mk
		r[0] = newPrev
		for k := 1; k <= e; k++ {
			cur := r[k&(maxLevels-1)]
			nw := ((cur >> 1) & mk) | (((oldPrev >> 1) | (newPrev >> 1)) & allow) | (oldPrev & gap)
			r[k&(maxLevels-1)] = nw
			oldPrev, newPrev = cur, nw
		}
		if r[e&(maxLevels-1)]&1 != 0 {
			return t, true
		}
	}
	return 0, false
}
