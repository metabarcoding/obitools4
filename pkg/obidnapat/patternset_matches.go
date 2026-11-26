package obidnapat

import (
	"slices"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// AppendAllMatches appends to dst, for every pattern and every strand, the
// windows that ApatPattern.AllMatches gives for that pattern (the ones that do
// not overlap a better one; the real start for a pattern with indels), by
// increasing Start, then by Pattern, then by strand (+1 before -1). Linear
// sequences only. begin and length are those of AllMatches.
func (s *PatternSet) AppendAllMatches(dst []Occurrence, seq *obiseq.BioSequence, begin, length int) []Occurrence {
	sq := seq.Sequence()
	first := len(dst)
	dst = s.collect(dst, sq, begin, length, false, true)
	raw := dst[first:]

	// grouped by pattern and strand (counting sort: the windows of one core
	// come by increasing end, which is also their nominal start)
	sc := s.scratch.Get().(*matchScratch)
	defer s.scratch.Put(sc)
	keyOf := func(o Occurrence) int { return 2*o.Pattern + int(1-o.Strand)/2 }
	sc.count = append(sc.count[:0], make([]int, 2*s.n+1)...)
	for _, o := range raw {
		sc.count[keyOf(o)+1]++
	}
	for k := 1; k < len(sc.count); k++ {
		sc.count[k] += sc.count[k-1]
	}
	sc.sorted = append(sc.sorted[:0], raw...)
	for _, o := range raw {
		k := keyOf(o)
		sc.sorted[sc.count[k]] = o
		sc.count[k]++
	}
	grouped := sc.sorted

	begin = max(begin, 0)
	out := first
	for i := 0; i < len(grouped); {
		j := i
		for j < len(grouped) && grouped[j].Pattern == grouped[i].Pattern && grouped[j].Strand == grouped[i].Strand {
			j++
		}
		o := grouped[i]
		c := s.cs[keyOf(o)]
		sc.win = sc.win[:0]
		for _, h := range grouped[i:j] {
			sc.win = append(sc.win, [3]int{h.Start, h.End, h.Errors})
		}
		sc.best = appendBestOfOverlaps(sc.best[:0], sc.win)
		for _, w := range sc.best {
			if c.indel {
				if w[2] > 0 {
					if st, ok := c.backwardStart(sq, w[1]-1, w[2], begin); ok {
						w[0] = st
					}
				}
				if w[0] < begin {
					continue
				}
			}
			dst[out] = Occurrence{o.Pattern, o.Strand, w[0], w[1], w[2]}
			out++
		}
		i = j
	}
	dst = dst[:out]
	slices.SortFunc(dst[first:], compareOccurrences)
	return dst
}

// matchScratch holds the buffers of one call of AppendAllMatches.
type matchScratch struct {
	count  []int
	sorted []Occurrence
	win    [][3]int
	best   [][3]int
}
