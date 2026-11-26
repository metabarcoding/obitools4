package obidnapat

// Searches that allow insertions and deletions: Wu and Manber's extension of the
// bit-parallel search, the one obiapat's ManberIndel implements (agrep). With the
// same convention as scan.go (bit i of r_j: the first i+1 positions of the
// pattern are matched with at most j errors, the empty prefix being always
// matched), a byte b of the text gives, from the old states o_j and the new ones
// n_j:
//
//	n_0 = ((o_0 << 1) | 1) & mask[b]
//	n_j = (((o_j << 1) | 1) & mask[b])
//	    | ( o_{j-1}                // insertion: the byte is an extra one
//	      | (o_{j-1} << 1) | 1     // substitution
//	      | (n_{j-1} << 1) )       // deletion: a position of the pattern is skipped
//	      & allow
//
// and the initial state is o_j = 2^j - 1: j leading positions deleted. (obiapat
// also adds the empty prefix to the deletions of the levels above 1; the
// substitution term already holds it, so it changes nothing.) The end of a window is reported where the last bit
// of a level is set, with the smallest such level.
//
// As in obiapat the reported window is the nominal one: [pos-len+1, pos+1) for a
// match ending at pos, whatever the insertions and deletions that matched, so
// its start may be a few bases off (BestMatch realigns it). A window must lie in
// the region of the search (obiapat let the ones that start before it through,
// sometimes before the sequence).

// initIndel sets the initial state of a search with indels.
func (c *compiled) initIndel(st *state) {
	full := c.end | (c.end - 1)
	for j := 0; j <= c.levels; j++ {
		st[j] = (1<<uint(j) - 1) & full
	}
}

// runIndel scans text, whose first byte stands for the position base of the
// (possibly unrolled) sequence, and appends the windows whose last base is in it.
func (c *compiled) runIndel(dst [][3]int, text []byte, base int, st *state) [][3]int {
	if out, ok := c.runIndelUnrolled(dst, text, base, st); ok {
		return out
	}
	return c.runIndelN(dst, text, base, st)
}

// runIndelN is the kernel for any number of levels, the state in an array.
func (c *compiled) runIndelN(dst [][3]int, text []byte, base int, st *state) [][3]int {
	m, end, allow, off, n, top := &c.mask, c.end, c.allow, c.offset(base), len(c.positions), c.levels
	r := *st
	for j, b := range text {
		mk := m[b]
		prevOld := r[0] // the old state of the level below
		prevNew := ((prevOld << 1) | 1) & mk
		r[0] = prevNew
		for l := 1; l <= top; l++ {
			cur := r[l&(maxLevels-1)]
			nw := (((cur << 1) | 1) & mk) | ((prevOld | (prevOld << 1) | 1 | (prevNew << 1)) & allow)
			r[l&(maxLevels-1)] = nw
			prevOld, prevNew = cur, nw
		}
		if r[top&(maxLevels-1)]&end != 0 {
			e := 0
			for r[e&(maxLevels-1)]&end == 0 {
				e++
			}
			dst = append(dst, [3]int{off + j, off + j + n, e})
		}
	}
	*st = r
	return dst
}

// searchIndel is search for a pattern with indels. begin and length delimit the
// region as in search: for a linear sequence the whole (nominal) window must lie
// in it; for a circular one its start must, and it may cross the junction.
func (c *compiled) searchIndel(dst [][3]int, seq []byte, begin, length int, circular bool) [][3]int {
	if begin < 0 {
		begin = 0
	}
	first := len(dst)
	dst = c.indelHits(dst, seq, begin, length, circular)

	// the windows that start before the region come first
	k := first
	for _, w := range dst[first:] {
		if w[0] >= begin {
			dst[k] = w
			k++
		}
	}
	return dst[:k]
}

// indelHits appends the nominal windows of every end that the forward search
// reports in the region, those whose nominal start falls before the region
// included (a window with deletions is shorter than the pattern: its real start can
// lie in the region when its nominal start does not). begin must not be negative.
func (c *compiled) indelHits(dst [][3]int, seq []byte, begin, length int, circular bool) [][3]int {
	n, m := len(seq), len(c.positions)
	if begin >= n || m > n && circular {
		return dst
	}
	limit := n
	if length >= 0 && length < n-begin {
		limit = begin + length
	}
	to := limit // the unrolled positions read: [begin, to)
	if circular {
		to = limit - 1 + m
	}
	if to-begin < 1 {
		return dst
	}

	noisy := c.needsCheck(hasN(seq))
	var st state
	c.initIndel(&st)
	for v := begin; v < to; {
		p := v % n
		take := min(to-v, n-p)
		if noisy {
			take = min(take, noiseChunk)
		}
		mark := len(dst)
		dst = c.runIndel(dst, seq[p:p+take], v, &st)
		if noisy {
			dst = dropNoise(dst, mark, seq, circular)
		}
		v += take
	}
	return dst
}
