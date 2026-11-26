package obidnapat

// A vgroup runs several words side by side, one per lane of a SIMD vector, with
// the recurrence of word.run2: the limit of 64 positions in all disappears, a
// long core has a lane of its own. The lanes of a vector run the same number of
// levels, so a group gathers words of the same levels. Prototype: levels = 2.
//
// The vector kernel only tells, every vectorBlock bytes, which lanes have seen an
// end; the occurrences of these lanes are then produced by the scalar kernel of
// their word from the state saved at the start of the block, so that they are
// exactly the scalar ones.

const (
	vectorLevels = 2  // the number of levels the vector kernel is written for
	vectorBlock  = 64 // bytes between two looks at the ends
)

type vgroup struct {
	lanes int
	words []int    // lanes indices in PatternSet.words
	masks []uint64 // masks[b*lanes+i]: mask of the byte b in the lane i
	start []uint64 // per lane: start, end and allow of the word
	end   []uint64
	allow []uint64
}

// vectorize gathers the words of vectorLevels levels by groups of one vector;
// the words that do not fill a group keep their scalar kernel.
func (s *PatternSet) vectorize() {
	lanes := vectorLanes()
	if lanes < 2 {
		return
	}
	var idx []int
	for i := range s.words {
		if s.words[i].levels == vectorLevels {
			idx = append(idx, i)
		}
	}
	for ; len(idx) >= lanes; idx = idx[lanes:] {
		g := vgroup{lanes: lanes, words: idx[:lanes:lanes], masks: make([]uint64, 256*lanes),
			start: make([]uint64, lanes), end: make([]uint64, lanes), allow: make([]uint64, lanes)}
		for l, wi := range g.words {
			w := &s.words[wi]
			w.vector = true
			for b := range w.mask {
				g.masks[b*lanes+l] = w.mask[b]
			}
			g.start[l], g.end[l], g.allow[l] = w.start, w.end, w.allow
		}
		s.groups = append(s.groups, g)
	}
}

// scan runs the group over the part of the sequence where one of its words can
// match, and appends the occurrences.
func (g *vgroup) scan(s *PatternSet, dst []Occurrence, seq []byte, noisy bool, begin, length int, circular bool) []Occurrence {
	n, lanes := len(seq), g.lanes
	fs := make([]windows, lanes)
	lo, hi := -1, -1
	for l, wi := range g.words {
		f, l0, h0 := s.words[wi].windows(seq, noisy, begin, length, circular)
		if l0 < 0 {
			continue
		}
		fs[l] = f
		if lo < 0 || l0 < lo {
			lo = l0
		}
		hi = max(hi, h0)
	}
	if lo < 0 {
		return dst
	}

	st := make([]uint64, 3*lanes)    // r0, r1, r2 of every lane
	saved := make([]uint64, 3*lanes) // the state at the start of the block
	seen := make([]uint64, lanes)
	for v := lo; v < hi; {
		p := v % n
		take := min(hi-v, n-p)
		text := seq[p : p+take]
		for off := 0; off < len(text); off += vectorBlock {
			blk := text[off:min(off+vectorBlock, len(text))]
			copy(saved, st)
			g.block(st, blk, seen)
			for l := 0; l < lanes; l++ {
				if seen[l] == 0 {
					continue
				}
				var rs state
				rs[0], rs[1], rs[2] = saved[l], saved[lanes+l], saved[2*lanes+l]
				dst = s.words[g.words[l]].run(dst, blk, v+off, &rs, &fs[l])
			}
		}
		v += take
	}
	return dst
}
