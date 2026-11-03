package obiprimer

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
)

// WordSearch holds the parameters of the strict word search.
type WordSearch struct {
	PrimerLength int     // word size (ecoPrimers -O, 18 by default; 2..31)
	Quorum       float64 // fraction of example sequences that must contain the word (ecoPrimers -q, 0.7)
	Circular     bool    // sequences are circular (ecoPrimers -c)
	Partitions   int     // partition files of the counter (default 256)
	TmpDir       string  // parent of the temporary directory ("" = system default)
}

// Counts gives the number of sequences by role.
type Counts struct {
	Examples        int
	CounterExamples int
	Ignored         int
}

// StrictWords is the result of the strict word search: the canonical words of
// PrimerLength bases present, without any error, in at least MinSequences
// example sequences.
type StrictWords struct {
	PrimerLength int
	MinSequences int // quorum in number of example sequences
	Words        []obikmer.SeqKmerCount
	Counts       Counts
}

// MinSequences returns the quorum in sequences: floor(examples * quorum).
// The tiny epsilon makes exact products (10 * 0.7 = 7) robust to the binary
// representation of the fraction.
func MinSequences(examples int, quorum float64) int {
	return int(math.Floor(float64(examples)*quorum + 1e-9))
}

// FindStrictWords reads every sequence of seqs once, indexes the example
// sequences, and returns the words reaching the quorum. Sequences no longer
// than the word size are counted as examples but cannot contain any word, as
// in ecoPrimers. Words are canonical (the smaller of the word and its reverse
// complement), so both strands are searched.
func FindStrictWords(seqs obiiter.IBioSequence, sel Selection, cfg WordSearch) (*StrictWords, error) {
	if cfg.PrimerLength < 2 || cfg.PrimerLength > 31 {
		return nil, fmt.Errorf("obiprimer: primer length must be between 2 and 31, got %d", cfg.PrimerLength)
	}
	if cfg.Quorum <= 0 || cfg.Quorum > 1 {
		return nil, fmt.Errorf("obiprimer: quorum must be in ]0,1], got %g", cfg.Quorum)
	}
	if cfg.Partitions <= 0 {
		cfg.Partitions = 256
	}

	counter, err := obikmer.NewSeqKmerCounter(cfg.TmpDir, cfg.PrimerLength, 0, cfg.Partitions)
	if err != nil {
		return nil, err
	}
	defer counter.Close()

	var nextID atomic.Uint32
	var examples, counters, ignored atomic.Int64

	consumer := func(it obiiter.IBioSequence, wg *sync.WaitGroup) {
		defer wg.Done()
		for it.Next() {
			for _, seq := range it.Get().Slice() {
				switch sel.Role(seq) {
				case Example:
					examples.Add(1)
					if seq.Len() > cfg.PrimerLength {
						counter.AddSequence(nextID.Add(1)-1, seq.Sequence(), cfg.Circular)
					}
				case CounterExample:
					counters.Add(1)
				default:
					ignored.Add(1)
				}
			}
		}
	}

	var wg sync.WaitGroup
	for i := 1; i < obidefault.ParallelWorkers(); i++ {
		wg.Add(1)
		go consumer(seqs.Split(), &wg)
	}
	wg.Add(1)
	go consumer(seqs, &wg)
	wg.Wait()

	res := &StrictWords{
		PrimerLength: cfg.PrimerLength,
		Counts: Counts{
			Examples:        int(examples.Load()),
			CounterExamples: int(counters.Load()),
			Ignored:         int(ignored.Load()),
		},
	}
	res.MinSequences = MinSequences(res.Counts.Examples, cfg.Quorum)
	if res.Counts.Examples == 0 {
		return nil, fmt.Errorf("obiprimer: no example sequence selected")
	}

	res.Words, err = counter.CountAtLeast(max(res.MinSequences, 1))
	if err != nil {
		return nil, err
	}
	return res, nil
}
