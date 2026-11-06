package obidnapat

import (
	"math/bits"
	"slices"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// AppendAll appends to dst the occurrences of the patterns in seq, by increasing
// Start, then by Pattern, then by strand (+1 before -1). Restricted to one
// pattern and one strand, they are the matches of ApatPattern.AppendAllIndex.
// begin and length are those of the methods of ApatPattern.
func (s *PatternSet) AppendAll(dst []Occurrence, seq *obiseq.BioSequence, begin, length int) []Occurrence {
	return s.appendAll(dst, seq.Sequence(), begin, length, false)
}

// AppendAllCircular is AppendAll on a circular sequence.
func (s *PatternSet) AppendAllCircular(dst []Occurrence, seq *obiseq.BioSequence, begin, length int) []Occurrence {
	return s.appendAll(dst, seq.Sequence(), begin, length, true)
}

// FindAll returns the occurrences of the patterns in seq, nil if there is none.
func (s *PatternSet) FindAll(seq *obiseq.BioSequence, begin, length int) []Occurrence {
	return nilIfNone(s.AppendAll(nil, seq, begin, length))
}

// FindAllCircular is FindAll on a circular sequence.
func (s *PatternSet) FindAllCircular(seq *obiseq.BioSequence, begin, length int) []Occurrence {
	return nilIfNone(s.AppendAllCircular(nil, seq, begin, length))
}

func nilIfNone(o []Occurrence) []Occurrence {
	if len(o) == 0 {
		return nil
	}
	return o
}

func compareOccurrences(a, b Occurrence) int {
	switch {
	case a.Start != b.Start:
		return a.Start - b.Start
	case a.Pattern != b.Pattern:
		return a.Pattern - b.Pattern
	}
	return int(b.Strand) - int(a.Strand) // +1 first
}

func (s *PatternSet) appendAll(dst []Occurrence, seq []byte, begin, length int, circular bool) []Occurrence {
	first := len(dst)
	dst = s.collect(dst, seq, begin, length, circular, false)
	// The words come one after the other, and the ends of the windows in the
	// order of the scan: the occurrences are sorted by their start here.
	slices.SortFunc(dst[first:], compareOccurrences)
	return dst
}

// collect appends the occurrences, in no particular order. keepEarly: see
// scanIndelWords.
func (s *PatternSet) collect(dst []Occurrence, seq []byte, begin, length int, circular, keepEarly bool) []Occurrence {
	noisy := hasN(seq) // the windows are then checked (see noise.go)
	if circular {
		for _, m := range s.imems { // the words with indels read linear sequences only
			for _, loc := range m.c.search(nil, seq, begin, length, true) {
				dst = append(dst, Occurrence{m.pattern, m.strand, loc[0], loc[1], loc[2]})
			}
		}
	} else {
		dst = s.scanIndelWords(dst, seq, begin, length, keepEarly)
	}
	for _, m := range s.solo {
		for _, loc := range m.c.search(nil, seq, begin, length, circular) {
			dst = append(dst, Occurrence{m.pattern, m.strand, loc[0], loc[1], loc[2]})
		}
	}
	for _, b := range s.batches {
		dst = s.scanBatch(dst, b, seq, noisy, begin, length, circular)
	}
	for i := range s.groups {
		dst = s.groups[i].scan(s, dst, seq, noisy, begin, length, circular)
	}
	return dst
}

// scanBatch runs the words of a batch (the same number of levels) together over
// the part of the sequence where one of their cores can match, and appends the
// occurrences.
func (s *PatternSet) scanBatch(dst []Occurrence, idx []int, seq []byte, noisy bool, begin, length int, circular bool) []Occurrence {
	n := len(seq)
	var ws [maxBatchWords]*word
	var fs [maxBatchWords]windows
	var fp [maxBatchWords]*windows
	var sts [maxBatchWords]state
	var sp [maxBatchWords]*state
	lo, hi := -1, -1
	for k, wi := range idx {
		ws[k], fp[k], sp[k] = &s.words[wi], &fs[k], &sts[k]
		f, l0, h0 := ws[k].windows(seq, noisy, begin, length, circular)
		if l0 < 0 {
			continue
		}
		fs[k] = f
		if lo < 0 || l0 < lo {
			lo = l0
		}
		hi = max(hi, h0)
	}
	if lo < 0 {
		return dst
	}
	k := len(idx)
	for v := lo; v < hi; {
		p := v % n
		take := min(hi-v, n-p)
		text := seq[p : p+take]
		if k == 1 {
			dst = ws[0].run(dst, text, v, sp[0], fp[0])
		} else if out, ok := runWords(dst, text, v, ws[:k], sp[:k], fp[:k]); ok {
			dst = out
		} else {
			for j := 0; j < k; j++ {
				dst = ws[j].run(dst, text, v, sp[j], fp[j])
			}
		}
		v += take
	}
	return dst
}

// maxBatchWords is the largest batchSize (maxBatch of gen_kernels.go).
const maxBatchWords = 3

// windows computes, for a sequence of n bytes, where each member of the word is
// searched, and the part [lo, hi) of the unrolled sequence that holds them all
// (lo < 0: nothing to search).
func (w *word) windows(seq []byte, noisy bool, begin, length int, circular bool) (f windows, lo, hi int) {
	n := len(seq)
	f.seq, f.noisy, f.circular = seq, noisy, circular
	// f.from[k], f.to[k]: the part of the unrolled sequence where the core of the
	// member k is searched (as in compiled.search); to[k] <= from[k] when the
	// member cannot match.
	lo, hi = -1, -1
	for k := range w.members {
		c := w.members[k].c
		b, last, ok := c.span(n, begin, length, circular)
		if !ok {
			continue
		}
		f.from[k], f.to[k] = b+c.left, last+c.left+c.coreLen
		if lo < 0 || f.from[k] < lo {
			lo = f.from[k]
		}
		hi = max(hi, f.to[k])
	}
	return f, lo, hi
}

// windows is the part of the unrolled sequence where each member of a word is
// searched, for one call.
type windows struct {
	from, to [wordBits]int
	seq      []byte // the sequence, to check the windows (see noise.go)
	noisy    bool   // the sequence holds an N
	circular bool
}

// run scans text, whose first byte stands for the position pos of the unrolled
// sequence, with the kernel written out for the levels of the word if there is
// one (gen_kernels.go), and appends the occurrences.
func (w *word) run(dst []Occurrence, text []byte, pos int, st *state, f *windows) []Occurrence {
	if out, ok := w.runUnrolled(dst, text, pos, st, f); ok {
		return out
	}
	return w.runN(dst, text, pos, st, f)
}

// runN is the kernel for any number of levels, with the state in an array.
func (w *word) runN(dst []Occurrence, text []byte, pos int, st *state, f *windows) []Occurrence {
	top, start, end, allow := w.levels, w.start, w.end, w.allow
	r := *st // a local copy: the compiler keeps it on the stack, in one piece
	for j, b := range text {
		mk := w.mask[b]
		prev := r[0]
		r[0] = ((prev << 1) | start) & mk
		for l := 1; l <= top; l++ {
			cur := r[l&(maxLevels-1)]
			r[l&(maxLevels-1)] = (((cur << 1) | start) & mk) | (((prev << 1) | start) & allow)
			prev = cur
		}
		if hit := r[top&(maxLevels-1)] & end; hit != 0 {
			dst = w.emit(dst, pos+j, hit, &r, f)
		}
	}
	*st = r
	return dst
}

// emit appends the windows that end at the position pos of the unrolled
// sequence: hit has the end bits of the cores whose last level is set, r is the
// state after the byte.
func (w *word) emit(dst []Occurrence, pos int, hit uint64, r *state, f *windows) []Occurrence {
	for ; hit != 0; hit &= hit - 1 {
		bit := bits.TrailingZeros64(hit)
		k := int(w.owner[bit])
		m := &w.members[k]
		c := m.c
		if pos >= f.to[k] || pos-c.coreLen+1 < f.from[k] {
			continue // outside the part where this core is searched
		}
		e := 0
		for r[e&(maxLevels-1)]&m.endBit == 0 {
			e++
		}
		if e > c.levels {
			continue
		}
		st := c.offset(pos)
		if (f.noisy || e >= len(c.positions)) && realBases(f.seq, st, st+len(c.positions), f.circular) <= e {
			continue // matched by its errors and its N alone
		}
		dst = append(dst, Occurrence{m.pattern, m.strand, st, st + len(c.positions), e})
	}
	return dst
}
