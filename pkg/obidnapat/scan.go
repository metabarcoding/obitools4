package obidnapat

//go:generate go run gen_kernels.go

// The search is the bit-parallel one of Baeza-Yates and Gonnet, extended to
// substitutions by Wu and Manber: bit i of the word r_j says that the first i+1
// positions of the core are matched by the text read so far with at most j
// errors. For every byte of the text,
//
//	r_0' = ((r_0 << 1) | 1) & mask[b]
//	r_j' = (((r_j << 1) | 1) & mask[b]) | (((r_{j-1} << 1) | 1) & allow)
//
// The first term extends a match, the second one spends an error on a
// substitution, which is forbidden at the positions marked with # (allow). A
// window ends here when the last bit of r_j is set; its number of errors is the
// smallest j for which it is. Since r_j contains r_{j-1}, the last level alone
// tells whether there is a window.
//
// The levels up to maxUnrolled (gen_kernels.go) are written out, so that their
// states stay in registers; the other ones use an array.

// state is the words r_0..r_levels of a scan, kept between the segments of a
// circular sequence.
type state [maxLevels]uint64

// run scans text, whose first byte stands for the position base of the
// (possibly unrolled) sequence, and appends the windows whose core ends in it.
func (c *compiled) run(dst [][3]int, text []byte, base int, st *state) [][3]int {
	if c.levels == 0 {
		return c.run0(dst, text, base, st)
	}
	if out, ok := c.runUnrolled(dst, text, base, st); ok {
		return out
	}
	return c.runN(dst, text, base, st)
}

// offset converts the index of a byte of the text, where a core window ends, in
// the start of the whole pattern.
func (c *compiled) offset(base int) int { return base - c.coreLen + 1 - c.left }

func (c *compiled) run0(dst [][3]int, text []byte, base int, st *state) [][3]int {
	r, m, end, off, n := st[0], &c.mask, c.end, c.offset(base), len(c.positions)
	for j, b := range text {
		r = ((r << 1) | 1) & m[b]
		if r&end != 0 {
			dst = append(dst, [3]int{off + j, off + j + n, 0})
		}
	}
	st[0] = r
	return dst
}

func (c *compiled) runN(dst [][3]int, text []byte, base int, st *state) [][3]int {
	m, end, allow, off, n, top := &c.mask, c.end, c.allow, c.offset(base), len(c.positions), c.levels
	r := *st // a local copy: the compiler keeps it on the stack, in one piece
	for j, b := range text {
		mk := m[b]
		prev := r[0]
		r[0] = ((prev << 1) | 1) & mk
		// ascending order, with the previous value of the level below kept in prev
		for l := 1; l <= top; l++ {
			cur := r[l&(maxLevels-1)]
			r[l&(maxLevels-1)] = (((cur << 1) | 1) & mk) | (((prev << 1) | 1) & allow)
			prev = cur
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

// span gives, for a sequence of n bytes, the first and the last start of a
// window of the pattern (begin is brought back to 0 if negative); ok is false
// when no window can exist.
func (c *compiled) span(n, begin, length int, circular bool) (first, last int, ok bool) {
	m := len(c.positions)
	if begin < 0 {
		begin = 0
	}
	if begin >= n || m > n && circular {
		return 0, 0, false
	}
	limit := n
	if length >= 0 && length < n-begin {
		limit = begin + length
	}
	last = limit - m
	if circular {
		last = limit - 1
	}
	return begin, last, last >= begin
}

// search appends to dst the windows of the pattern found in seq. A window is
// reported by its start, its end (start + Len(), which may go past the end of a
// circular sequence) and its number of errors; the windows come by increasing
// start.
//
// begin and length delimit the region where the pattern is looked for (length
// < 0: up to the end of the sequence). For a linear sequence the whole window
// must lie in the region. For a circular one the start must, and the window may
// cross the junction. Each place is reported once.
func (c *compiled) search(dst [][3]int, seq []byte, begin, length int, circular bool) [][3]int {
	m := len(c.positions)
	begin, lastStart, ok := c.span(len(seq), begin, length, circular)
	if !ok {
		return dst
	}
	if c.coreLen == 0 { // nothing but wildcards: every window matches
		for s := begin; s <= lastStart; s++ {
			dst = append(dst, [3]int{s, s + m, 0})
		}
		return dst
	}

	// the core text, in positions of the unrolled sequence
	var st state
	n := len(seq)
	from, to := begin+c.left, lastStart+c.left+c.coreLen
	if c.bndm && !circular {
		return c.runBNDM(dst, seq[from:to], from)
	}
	for v := from; v < to; {
		p := v % n
		take := min(to-v, n-p)
		dst = c.run(dst, seq[p:p+take], v, &st)
		v += take
	}
	return dst
}
