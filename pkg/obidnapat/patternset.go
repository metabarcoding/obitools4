package obidnapat

import (
	"fmt"
	"slices"
)

// wordBits is the number of positions a packed word can hold. It is only the
// capacity of the machine word: the limit of a pattern remains MaxCoreLength.
const wordBits = 64

// Occurrence is a window of a pattern of a PatternSet. Start and End are those
// of ApatPattern (End is not included, and may go past the end of a circular
// sequence).
type Occurrence struct {
	Pattern int  // index of the pattern in the slice given to MakePatternSet
	Strand  int8 // +1: the pattern itself, -1: its reverse complement
	Start   int
	End     int
	Errors  int
}

// member is a core packed in a word.
type member struct {
	c       *compiled
	pattern int
	strand  int8
	shift   uint
	endBit  uint64 // bit of the last position of the core in the word
}

// word holds several cores side by side in 64 bits. Its recurrence is that of
// scan.go with start in place of 1: start has the first bit of every core, which
// is forced to 1 at every step, so that what a core pushes out of its last bit
// into the first one of the next core is of no consequence.
type word struct {
	mask    [256]uint64
	start   uint64
	end     uint64
	allow   uint64
	levels  int // error levels: the largest one of the members
	members []member
	owner   [wordBits]int8 // owner[i]: index in members of the core that ends at bit i
	vector  bool           // run by a vgroup (patternset_vec.go), not on its own
}

// PatternSet looks for several patterns, and optionally their reverse
// complements, in one pass over the sequence. It is immutable and can be used by
// several goroutines at once.
type PatternSet struct {
	words   []word
	groups  []vgroup
	batches [][]int  // indices of the words advanced in one loop: same levels, at most batchSize
	solo    []member // searched on their own: the patterns made of wildcards only (no core), and the strict ones that BNDM reads faster than a word
	n       int
}

// MakePatternSet compiles the search of patterns. With bothStrands, the reverse
// complement of every pattern is searched as well (a palindrome is found once
// per strand).
func MakePatternSet(patterns []ApatPattern, bothStrands bool) (*PatternSet, error) {
	return makePatternSet(patterns, bothStrands, false)
}

// makePatternSet is MakePatternSet; vector tells whether the words of 2 levels
// may run on SIMD vectors (when the program is built for it). That prototype is
// off by default: advancing several words in one loop is faster (see the notes).
func makePatternSet(patterns []ApatPattern, bothStrands, vector bool) (*PatternSet, error) {
	var entries []member
	for i, p := range patterns {
		if p.c == nil {
			return nil, fmt.Errorf("obidnapat: pattern %d of the set is not compiled", i)
		}
		entries = append(entries, member{c: p.c, pattern: i, strand: 1})
		if bothStrands {
			rc, err := p.ReverseComplement()
			if err != nil {
				return nil, fmt.Errorf("obidnapat: reverse complement of pattern %d: %w", i, err)
			}
			entries = append(entries, member{c: rc.c, pattern: i, strand: -1})
		}
	}

	s := &PatternSet{n: len(patterns)}

	// The cores are packed by first fit decreasing, among the ones that search
	// the same number of levels: a pattern without error does not pay for the
	// levels of its neighbours.
	byLevels := map[int][]member{}
	for _, e := range entries {
		if e.c.coreLen == 0 {
			s.solo = append(s.solo, e)
		} else {
			byLevels[e.c.levels] = append(byLevels[e.c.levels], e)
		}
	}
	levels := make([]int, 0, len(byLevels))
	for l := range byLevels {
		levels = append(levels, l)
	}
	slices.Sort(levels)
	for _, l := range levels {
		group := byLevels[l]
		slices.SortStableFunc(group, func(a, b member) int { return b.c.coreLen - a.c.coreLen })
		var used []int
		first := len(s.words)
		for _, e := range group {
			k := first
			for ; k < len(s.words) && used[k-first]+e.c.coreLen > wordBits; k++ {
			}
			if k == len(s.words) {
				s.words = append(s.words, word{levels: l})
				used = append(used, 0)
			}
			s.words[k].add(e, uint(used[k-first]))
			used[k-first] += e.c.coreLen
		}
	}

	// A word with a single strict core is no gain over its own search, and BNDM
	// skips bytes where the word reads them all (BenchmarkSetLen: 148-288 Mo/s
	// against 80 for cores of more than 32 positions).
	kept := s.words[:0]
	for _, w := range s.words {
		if len(w.members) == 1 && w.levels == 0 && w.members[0].c.bndm {
			s.solo = append(s.solo, w.members[0])
		} else {
			kept = append(kept, w)
		}
	}
	s.words = kept
	if vector {
		s.vectorize()
	}
	s.makeBatches()
	return s, nil
}

// add packs the core of e from the bit shift.
func (w *word) add(e member, shift uint) {
	c := e.c
	e.shift, e.endBit = shift, c.end<<shift
	for b := range w.mask {
		w.mask[b] |= c.mask[b] << shift
	}
	w.start |= 1 << shift
	w.end |= e.endBit
	w.allow |= c.allow << shift
	w.owner[shift+uint(c.coreLen)-1] = int8(len(w.members))
	w.members = append(w.members, e)
}

// Len returns the number of patterns of the set (not counting the strands).
func (s *PatternSet) Len() int { return s.n }

// Words returns the number of 64 bit words the search runs on (the patterns
// searched on their own are not counted).
func (s *PatternSet) Words() int { return len(s.words) }

// forcedBatch, when not 0, is the number of words advanced in one loop whatever
// their levels (1 to maxBatch of gen_kernels.go). A variable for the tests and benchmarks.
var forcedBatch int

// batchFor is the number of words advanced in one loop for words of that many levels.
func batchFor(levels int) int {
	if forcedBatch > 0 {
		return forcedBatch
	}
	// BenchmarkBatch (M3 Max): 3 words are best up to 3 levels (x1.4 to x2.3 on
	// one word), 2 words at 4 and 5 levels (x1.3; 5 is interpolated, not measured),
	// and from 6 levels on the gain is within the noise or negative, the states
	// no longer fitting the registers.
	switch {
	case levels <= 3:
		return 3
	case levels <= 5:
		return 2
	}
	return 1
}

// makeBatches gathers the words that do not run on vectors by batches of the
// same number of levels. The words come class of levels after class of levels.
func (s *PatternSet) makeBatches() {
	for i := 0; i < len(s.words); {
		if s.words[i].vector {
			i++
			continue
		}
		b := []int{i}
		for i++; i < len(s.words) && len(b) < batchFor(s.words[i].levels) && !s.words[i].vector && s.words[i].levels == s.words[b[0]].levels; i++ {
			b = append(b, i)
		}
		s.batches = append(s.batches, b)
	}
}
