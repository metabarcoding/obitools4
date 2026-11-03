package obiprimer

import (
	"cmp"
	"fmt"
	"slices"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
)

// Match is one approximate occurrence of a word in a corpus sequence.
//
// Pos is the 0-based start of the window on the sequence as given (forward
// coordinates). When Reverse is false the window reads like the word; when it
// is true the window reads like the reverse complement of the word, i.e. the
// word occurs on the other strand. Errors is the number of substitutions.
type Match struct {
	Seq     uint32 // index in the Corpus
	Pos     uint32
	Errors  uint8
	Reverse bool
}

// WordMatches gathers every occurrence of one canonical word, sorted by
// (Seq, Pos, Reverse).
type WordMatches struct {
	Word    uint64
	Matches []Match
}

// MatchParams parameterizes the approximate search.
//
// Rules, those of ecoPrimers: only substitutions are allowed; a base other
// than A, C, G, T (or U, read as T) in the sequence never matches, it counts as
// one error; every window with at most MaxErrors errors is reported, windows
// overlapping each other included; on circular sequences windows may span the
// junction.
type MatchParams struct {
	K         int
	MaxErrors int
	Circular  bool
}

func (p MatchParams) validate() error {
	if p.K < 2 || p.K > 31 {
		return fmt.Errorf("obiprimer: word size must be between 2 and 31, got %d", p.K)
	}
	if p.MaxErrors < 0 || p.MaxErrors >= p.K {
		return fmt.Errorf("obiprimer: error count must be in [0,%d[, got %d", p.K, p.MaxErrors)
	}
	return nil
}

// hit is a Match before its sequence is known: the sequence it belongs to is
// the one being scanned.
type hit struct {
	word    uint32
	pos     uint32
	errors  uint8
	reverse bool
}

// scanFunc finds every occurrence of every word, on both strands, in one
// sequence (its bases) and reports them through emit.
type scanFunc func(raw []byte, emit func(word, pos, errors int, reverse bool))

// scannerMaker returns an independent scanFunc; each worker gets its own so
// that scanners can keep private buffers.
type scannerMaker func() scanFunc

// wordStrings returns, for each canonical word, the word and its reverse
// complement as lowercase strings.
func wordStrings(words []uint64, k int) (fwd, rev []string) {
	fwd = make([]string, len(words))
	rev = make([]string, len(words))
	for i, w := range words {
		fwd[i] = string(obikmer.DecodeKmer(w, k, nil))
		rev[i] = string(obikmer.DecodeKmer(obikmer.ReverseComplement(w, k), k, nil))
	}
	return
}

// matchCorpus runs the scanners over every sequence of the corpus in parallel
// and assembles the per-word results. The output does not depend on the number
// of workers.
//
// The occurrences of a sequence are kept in a slice of their exact size, then
// counted per word, so that the occurrences of all the words are laid out in
// one allocation of the exact size too: nothing grows by doubling, nor is held
// twice for long.
func matchCorpus(c *Corpus, words []uint64, mk scannerMaker) []WordMatches {
	perSeq := make([][]hit, c.Len())

	jobs := make(chan int)
	var wg sync.WaitGroup
	for range obidefault.ParallelWorkers() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scan := mk()
			var buf []hit
			var raw, bases []byte
			for i := range jobs {
				buf = buf[:0]
				raw, bases = c.bases(i, bases)
				scan(raw, func(word, pos, errors int, reverse bool) {
					buf = append(buf, hit{uint32(word), uint32(pos), uint8(errors), reverse})
				})
				if len(buf) > 0 {
					perSeq[i] = slices.Clone(buf)
				}
			}
		}()
	}
	for i := range c.Roles {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	res := make([]WordMatches, len(words))
	count := make([]int, len(words))
	for _, hits := range perSeq {
		for _, h := range hits {
			count[h.word]++
		}
	}
	for i, w := range words {
		res[i].Word = w
		if count[i] > 0 {
			res[i].Matches = make([]Match, 0, count[i])
		}
	}
	for s, hits := range perSeq {
		for _, h := range hits {
			res[h.word].Matches = append(res[h.word].Matches, Match{uint32(s), h.pos, h.errors, h.reverse})
		}
		perSeq[s] = nil // the hits of the sequence are no longer needed
	}
	parallelFor(len(res), func(_, i int) {
		slices.SortFunc(res[i].Matches, func(a, b Match) int {
			if c := cmp.Compare(a.Seq, b.Seq); c != 0 {
				return c
			}
			if c := cmp.Compare(a.Pos, b.Pos); c != 0 {
				return c
			}
			if a.Reverse != b.Reverse {
				if !a.Reverse {
					return -1
				}
				return 1
			}
			return 0
		})
	})
	return res
}

// SequencesMatched returns, for the matches of one word, the number of
// distinct example and counterexample sequences it occurs in. The other
// sequences are not counted.
func (w WordMatches) SequencesMatched(c *Corpus) (examples, counterExamples int) {
	last := ^uint32(0)
	for _, m := range w.Matches {
		if m.Seq == last {
			continue
		}
		last = m.Seq
		switch c.Roles[m.Seq] {
		case Example:
			examples++
		case CounterExample:
			counterExamples++
		}
	}
	return
}
