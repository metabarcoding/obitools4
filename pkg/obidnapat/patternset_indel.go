package obidnapat

import "math/bits"

// The words of the patterns with insertions and deletions. Their recurrence is
// that of scan_indel.go with the start bit of every core in place of 1 (the same
// substitution as for the words of the substitutions):
//
//	n_0 = ((o_0 << 1) | start) & mk
//	n_j = (((o_j << 1) | start) & mk)
//	    | ((o_{j-1} | ((o_{j-1} << 1) | start) | (n_{j-1} << 1)) & allow)
//
// and the initial state is that of every core, o_j = 2^j - 1 within its bits. The
// bit that the deletion term pushes from the last position of a core into the
// first one of the next core is of no consequence: the substitution term already
// sets it, under the same allow bit.
//
// A word reports the nominal window [pos-len+1, pos+1) of every end (as
// compiled.indelHits does), with the smallest level that holds the end.

// initIndel sets the initial state of a word of cores with indels.
func (w *word) initIndel(st *state) {
	for j := 0; j <= w.levels; j++ {
		for k := range w.members {
			m := &w.members[k]
			full := m.c.end | (m.c.end - 1)
			st[j] |= ((1<<uint(j) - 1) & full) << m.shift
		}
	}
}

// scanIndelWords appends the nominal windows that the words with indels find in
// the region [begin, begin+length) of a linear sequence. A window whose nominal
// start precedes the region is kept when keepEarly (a window with deletions is
// shorter than the pattern: its real start can lie in the region).
func (s *PatternSet) scanIndelWords(dst []Occurrence, seq []byte, begin, length int, keepEarly bool) []Occurrence {
	n := len(seq)
	begin = max(begin, 0)
	if len(s.iwords) == 0 || begin >= n {
		return dst
	}
	limit := n
	if length >= 0 && length < n-begin {
		limit = begin + length
	}
	if limit-begin < 1 {
		return dst
	}
	first := len(dst)
	f := windows{seq: seq, noisy: true}
	for k := range f.to {
		f.from[k], f.to[k] = begin, limit
	}
	text := seq[begin:limit]
	for i := 0; i < len(s.iwords); {
		// the words of the same levels advance together
		var ws [maxBatchWords]*word
		var sts [maxBatchWords]state
		var sp [maxBatchWords]*state
		k := 0
		for ; i < len(s.iwords) && k < min(batchFor(s.iwords[i].levels), maxBatchWords) && (k == 0 || s.iwords[i].levels == ws[0].levels); i++ {
			ws[k], sp[k] = &s.iwords[i], &sts[k]
			ws[k].initIndel(sp[k])
			k++
		}
		if out, ok := runIndelWords(dst, text, begin, ws[:k], sp[:k], &f); ok && k > 1 {
			dst = out
			continue
		}
		for j := 0; j < k; j++ {
			dst = ws[j].runIndel(dst, text, begin, sp[j], &f)
		}
	}
	if !keepEarly {
		k := first
		for _, o := range dst[first:] {
			if o.Start >= begin {
				dst[k] = o
				k++
			}
		}
		dst = dst[:k]
	}
	return dst
}

// runIndel scans text, whose first byte stands for the position pos, with the
// kernel written out for the levels of the word if there is one (gen_kernels.go).
func (w *word) runIndel(dst []Occurrence, text []byte, pos int, st *state, f *windows) []Occurrence {
	if out, ok := w.runIndelUnrolled(dst, text, pos, st, f); ok {
		return out
	}
	return w.runIndelN(dst, text, pos, st, f)
}

// runIndelN is the kernel for any number of levels, with the state in an array.
func (w *word) runIndelN(dst []Occurrence, text []byte, pos int, st *state, f *windows) []Occurrence {
	top, start, end, allow := w.levels, w.start, w.end, w.allow
	r := *st
	for j, b := range text {
		mk := w.mask[b]
		prevOld := r[0]
		prevNew := ((prevOld << 1) | start) & mk
		r[0] = prevNew
		for l := 1; l <= top; l++ {
			cur := r[l&(maxLevels-1)]
			nw := (((cur << 1) | start) & mk) | ((prevOld | ((prevOld << 1) | start) | (prevNew << 1)) & allow)
			r[l&(maxLevels-1)] = nw
			prevOld, prevNew = cur, nw
		}
		if hit := r[top&(maxLevels-1)] & end; hit != 0 {
			dst = w.emitIndel(dst, pos+j, hit, &r, f)
		}
	}
	*st = r
	return dst
}

// emitIndel appends the nominal windows that end at the position pos: hit has the
// end bits of the cores whose last level is set, r is the state after the byte.
// The windows that hold no more real bases than errors are dropped (noise.go).
func (w *word) emitIndel(dst []Occurrence, pos int, hit uint64, r *state, f *windows) []Occurrence {
	for ; hit != 0; hit &= hit - 1 {
		m := &w.members[w.owner[bits.TrailingZeros64(hit)]]
		e := 0
		for r[e&(maxLevels-1)]&m.endBit == 0 {
			e++
		}
		n := len(m.c.positions)
		st := pos - n + 1
		if realBases(f.seq, st, st+n, false) <= e {
			continue
		}
		dst = append(dst, Occurrence{m.pattern, m.strand, st, st + n, e})
	}
	return dst
}
