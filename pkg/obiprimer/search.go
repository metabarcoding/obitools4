package obiprimer

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obithermo"
)

// SearchParams gathers every parameter of the primer pair search, named after
// the ecoPrimers options they come from.
type SearchParams struct {
	PrimerLength   int     // -O: size of the primers (default 18)
	MaxErrors      int     // -e: substitutions allowed between a primer and its sites (default 3)
	MinLength      int     // -l: the amplicon (primers excluded) must be longer than this
	MaxLength      int     // -L: and shorter than this
	ThreePrime     int     // -3: number of 3' bases that must match exactly
	StrictQuorum   float64 // -q: fraction of the examples containing the exact word (0.7)
	Sensitivity    float64 // -s: fraction of the examples a primer must match (0.9)
	FalsePositive  float64 // -x: fraction of the counterexamples it may match (0.1)
	MinSpecificity float64 // -T: pairs with a lower Bs are dropped (0.6)
	Circular       bool    // -c: sequences are circular

	Partitions int // partitions of the k-mer counter (default 256)
	TmpDir     string

	Model *obithermo.Model // melting temperatures; nil leaves them NaN
}

// DefaultSearchParams returns the ecoPrimers defaults.
func DefaultSearchParams() SearchParams {
	return SearchParams{
		PrimerLength: 18, MaxErrors: 3, MinLength: 0, MaxLength: 1000,
		StrictQuorum: 0.7, Sensitivity: 0.9, FalsePositive: 0.1, MinSpecificity: 0.6,
		Partitions: 256, Model: obithermo.DefaultModel(),
	}
}

// SearchInfo reports the size of every stage of a search.
type SearchInfo struct {
	Counts      Counts
	StrictWords int // words found without error in at least StrictQuorum of the examples
	MinStrict   int // that quorum, in sequences
	Matches     int // approximate occurrences of those words
	Pair        PairInfo
}

// Search runs the whole ecoPrimers pipeline on a corpus: strict words,
// approximate occurrences, pairs and their indices. The pairs are sorted by
// decreasing Yule*Bs.
func Search(c *Corpus, taxa *Taxa, p SearchParams) ([]PairStats, *SearchInfo, error) {
	if p.MaxErrors < 0 {
		return nil, nil, fmt.Errorf("obiprimer: the allowed mismatch count cannot be negative")
	}

	sw, err := FindStrictWords(c, WordSearch{
		PrimerLength: p.PrimerLength,
		Quorum:       p.StrictQuorum,
		Circular:     p.Circular,
		Partitions:   p.Partitions,
		TmpDir:       p.TmpDir,
	})
	if err != nil {
		return nil, nil, err
	}
	info := &SearchInfo{Counts: sw.Counts, StrictWords: len(sw.Words), MinStrict: sw.MinSequences}

	words := make([]uint64, len(sw.Words))
	for i, w := range sw.Words {
		words[i] = w.Kmer
	}
	wm, err := MatchWords(c, words, MatchParams{K: p.PrimerLength, MaxErrors: p.MaxErrors, Circular: p.Circular})
	if err != nil {
		return nil, info, err
	}
	for _, w := range wm {
		info.Matches += len(w.Matches)
	}

	pairs, pinfo, err := FindPairs(c, taxa, wm, PairParams{
		PrimerLength:   p.PrimerLength,
		MinLength:      p.MinLength,
		MaxLength:      p.MaxLength,
		ThreePrime:     p.ThreePrime,
		Sensitivity:    p.Sensitivity,
		FalsePositive:  p.FalsePositive,
		MinSpecificity: p.MinSpecificity,
		Circular:       p.Circular,
	}, p.Model)
	info.Pair = pinfo
	return pairs, info, err
}
