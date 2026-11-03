package obiprimer

import (
	"fmt"
	"math"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
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

// FindStrictWords indexes the example sequences of the corpus and returns the
// words reaching the quorum. Sequences no longer than the word size are
// counted as examples but cannot contain any word, as in ecoPrimers. Words are
// canonical (the smaller of the word and its reverse complement), so both
// strands are searched.
func FindStrictWords(c *Corpus, cfg WordSearch) (*StrictWords, error) {
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

	jobs := make(chan int)
	var wg sync.WaitGroup
	for range obidefault.ParallelWorkers() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				if seq := c.Seqs[i]; c.Roles[i] == Example && seq.Len() > cfg.PrimerLength {
					counter.AddSequence(uint32(i), seq.Sequence(), cfg.Circular)
				}
			}
		}()
	}
	for i := range c.Seqs {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	res := &StrictWords{PrimerLength: cfg.PrimerLength, Counts: c.Counts()}
	res.MinSequences = MinSequences(res.Counts.Examples, cfg.Quorum)
	res.Words, err = counter.CountAtLeast(max(res.MinSequences, 1))
	if err != nil {
		return nil, err
	}
	return res, nil
}
