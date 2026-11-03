package obiprimer

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obithermo"
)

// PairStats describes one primer pair and its quality indices. Primers are
// written 5'->3' and ordered so that Primer1 < Primer2 (as in ecoPrimers).
type PairStats struct {
	Primer1, Primer2 string
	Good1, Good2     bool // primer conserved in the examples and rare in the counterexamples

	Examples        int // example sequences amplified
	CounterExamples int
	Yule            float32 // Examples/|E| - CounterExamples/|C|

	ExampleTaxa    int     // target taxa amplified
	CounterTaxa    int     // counterexample taxa amplified
	Bc             float32 // coverage: ExampleTaxa over all the target taxa
	Identified     int     // target taxa all of whose barcodes belong to them alone (see resolution.go)
	Bs             float32 // specificity: Identified over ExampleTaxa
	Bd             float32 // barcode discrimination: share of the target barcodes leading to a single taxon
	MinLen, MaxLen int
	MeanLen        float64
	Tm1, Tm2       float64 // melting temperature of the primers against their perfect target
	MinTm1, MinTm2 float64 // lowest melting temperature over the example amplicons
	GC1, GC2       int

	// With a reference sequence (-R): where the pair amplifies it. RefLocation
	// is the GenBank location of the two primer sites (1-based, left to right),
	// wrapped in complement() when Primer1 binds the other strand; RefAmplicon
	// is the amplified region, primers excluded, read from Primer1 to Primer2.
	// Both are empty when the pair does not amplify the reference.
	RefLocation string
	RefAmplicon string
}

// Score is the ranking criterion of ecoPrimers.
func (p PairStats) Score() float32 { return p.Yule * p.Bs }

// PairInfo summarizes the pair search, for reporting.
type PairInfo struct {
	Primers        int // conserved words kept as candidate primers
	GoodPrimers    int // of which good ones
	CandidatePairs int // pairs delimiting an amplicon in at least one sequence
	RetainedPairs  int // pairs passing the sensitivity and false positive quorums, before the Bs filter
	Candidates     int // of which pairs whose Bs upper bound reaches the threshold (their amplicons are gathered in full)
}

// FindPairs searches the primer pairs from the approximate occurrences of the
// strict words (see MatchApat) and evaluates them.
//
// A pair is a pair of conserved primers, at least one of them good, that
// delimit an amplicon of the allowed length. It is retained when it amplifies
// more than Sensitivity of the examples and less than FalsePositive of the
// counterexamples, and its specificity Bs reaches MinSpecificity. The result
// is sorted by decreasing Yule*Bs.
//
// model computes the melting temperatures; nil leaves them NaN.
func FindPairs(c *Corpus, taxa *Taxa, wm []WordMatches, p PairParams, model *obithermo.Model) ([]PairStats, PairInfo, error) {
	var info PairInfo
	if err := p.validate(); err != nil {
		return nil, info, err
	}
	if taxa == nil || len(taxa.Of) != c.Len() {
		return nil, info, fmt.Errorf("obiprimer: taxa do not match the corpus")
	}
	ps := selectPrimers(c, wm, p)
	wm = nil // the occurrences are in ps now
	info.Primers = len(ps.primers)
	for _, pr := range ps.primers {
		if pr.Good {
			info.GoodPrimers++
		}
	}
	if len(ps.primers) == 0 {
		return nil, info, nil
	}

	// Phase 1: which pairs amplify enough examples and few enough counterexamples.
	counts := c.Counts()
	acc := ps.countPairs(c, p)
	info.CandidatePairs = len(acc)
	sens, fpq := float32(p.Sensitivity), float32(p.FalsePositive)
	var kept []pairKey
	for k, a := range acc {
		q := float32(a.examples) / float32(counts.Examples)
		qfp := float32(0)
		if counts.CounterExamples > 0 {
			qfp = float32(a.counters) / float32(counts.CounterExamples)
		}
		if q > sens && qfp < fpq {
			kept = append(kept, k)
		}
	}
	info.RetainedPairs = len(kept)
	slices.SortFunc(kept, func(a, b pairKey) int {
		return cmp.Or(cmp.Compare(a.p1, b.p1), cmp.Compare(a.p2, b.p2),
			boolCmp(a.a1, b.a1), boolCmp(a.a2, b.a2))
	})

	// Phase 2a: the amplicons of the target sequences, which give the melting
	// temperatures, the reference sites, and an upper bound of Bs. It is cheap
	// and rejects most pairs on the specificity threshold.
	first := ps.boundPairs(c, taxa, p, kept, model)
	var survivors []int // indices in kept
	for i := range kept {
		if first.targets[i] > 0 && float32(first.identified[i])/float32(first.targets[i]) >= float32(p.MinSpecificity) {
			survivors = append(survivors, i)
		}
	}
	info.Candidates = len(survivors)

	// Phase 2b: for the survivors only, the amplicons of every sequence, whatever
	// its role: they decide which target taxa are identified.
	results := ps.evaluatePairs(c, taxa, p, kept, survivors, acc, first, model)

	var out []PairStats
	for _, st := range results {
		if st != nil {
			out = append(out, *st)
		}
	}
	slices.SortFunc(out, func(a, b PairStats) int {
		return cmp.Or(-cmp.Compare(a.Score(), b.Score()), cmp.Compare(a.Primer1, b.Primer1), cmp.Compare(a.Primer2, b.Primer2))
	})
	return out, info, nil
}

func boolCmp(a, b bool) int {
	switch {
	case a == b:
		return 0
	case !a:
		return -1
	}
	return 1
}

// evaluate puts the indices of one retained pair in a PairStats. It reports
// false when the pair must be dropped (no target taxon amplified, or
// specificity under the threshold).
func (ps *primerSet) evaluate(taxa *Taxa, counts Counts, p PairParams, k pairKey, a *pairAcc, res resolution,
	minTm1, minTm2 float32, model *obithermo.Model) (PairStats, bool) {

	if res.targetTaxa == 0 {
		return PairStats{}, false
	}
	st := PairStats{
		Examples:        int(a.examples),
		CounterExamples: int(a.counters),
		ExampleTaxa:     res.targetTaxa,
		CounterTaxa:     res.counterTaxa,
		Identified:      res.identified,
		MinLen:          int(a.minLen),
		MaxLen:          int(a.maxLen),
		MeanLen:         float64(a.sumLen) / float64(a.amplicons),
	}
	q := float32(a.examples) / float32(counts.Examples)
	qfp := float32(0)
	if counts.CounterExamples > 0 {
		qfp = float32(a.counters) / float32(counts.CounterExamples)
	}
	st.Yule = q - qfp
	st.Bc = float32(res.targetTaxa) / float32(taxa.ExampleTaxa)
	st.Bs = float32(res.identified) / float32(res.targetTaxa)
	st.Bd = float32(res.discriminant) / float32(res.barcodes)
	if st.Bs < float32(p.MinSpecificity) {
		return PairStats{}, false
	}

	return ps.finish(st, p, k, a, minTm1, minTm2, model), true
}

// finish adds the primers, their composition and melting temperatures, and puts
// the two primers in string order.
func (ps *primerSet) finish(st PairStats, p PairParams, k pairKey, a *pairAcc,
	minTm1, minTm2 float32, model *obithermo.Model) PairStats {

	lp := p.PrimerLength
	st.Primer1 = orientedWord(ps.primers[k.p1].Word, lp, k.a1)
	st.Primer2 = orientedWord(ps.primers[k.p2].Word, lp, k.a2)
	st.Good1, st.Good2 = ps.primers[k.p1].Good, ps.primers[k.p2].Good
	st.GC1, st.GC2 = obithermo.GCCount([]byte(st.Primer1)), obithermo.GCCount([]byte(st.Primer2))

	st.Tm1, st.Tm2, st.MinTm1, st.MinTm2 = math.NaN(), math.NaN(), math.NaN(), math.NaN()
	if model != nil {
		st.Tm1, _ = model.SelfTm([]byte(st.Primer1))
		st.Tm2, _ = model.SelfTm([]byte(st.Primer2))
		if !math.IsInf(float64(minTm1), 1) {
			st.MinTm1 = float64(minTm1)
		}
		if !math.IsInf(float64(minTm2), 1) {
			st.MinTm2 = float64(minTm2)
		}
	}

	if st.Primer2 < st.Primer1 {
		st.Primer1, st.Primer2 = st.Primer2, st.Primer1
		st.Good1, st.Good2 = st.Good2, st.Good1
		st.GC1, st.GC2 = st.GC2, st.GC1
		st.Tm1, st.Tm2 = st.Tm2, st.Tm1
		st.MinTm1, st.MinTm2 = st.MinTm2, st.MinTm1
	}
	return st
}
