package obiprimer

import (
	"cmp"
	"fmt"
	"slices"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
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

// hit is a Match before its sequence index is known to be final.
type hit struct {
	word int
	m    Match
}

// scanFunc finds every occurrence of every word, on both strands, in one
// sequence and reports them through emit.
type scanFunc func(seq *obiseq.BioSequence, emit func(word, pos, errors int, reverse bool))

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
func matchCorpus(c *Corpus, words []uint64, mk scannerMaker) []WordMatches {
	perSeq := make([][]hit, len(c.Seqs))

	jobs := make(chan int)
	var wg sync.WaitGroup
	for range obidefault.ParallelWorkers() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scan := mk()
			for i := range jobs {
				var hits []hit
				scan(c.Seqs[i], func(word, pos, errors int, reverse bool) {
					hits = append(hits, hit{word, Match{uint32(i), uint32(pos), uint8(errors), reverse}})
				})
				perSeq[i] = hits
			}
		}()
	}
	for i := range c.Seqs {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	res := make([]WordMatches, len(words))
	for i, w := range words {
		res[i].Word = w
	}
	for _, hits := range perSeq {
		for _, h := range hits {
			res[h.word].Matches = append(res[h.word].Matches, h.m)
		}
	}
	for i := range res {
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
	}
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
