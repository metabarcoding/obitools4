package obiprimer

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
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

	ExampleTaxa    int     // taxa amplified among the examples
	CounterTaxa    int     // taxa amplified among the counterexamples
	Bc             float32 // coverage: ExampleTaxa over all the example taxa
	Identified     int     // example taxa with an amplicon no other taxon shares
	Bs             float32 // specificity: Identified over ExampleTaxa
	MinLen, MaxLen int
	MeanLen        float64
	Tm1, Tm2       float64 // melting temperature of the primers against their perfect target
	MinTm1, MinTm2 float64 // lowest melting temperature over the example amplicons
	GC1, GC2       int
}

// Score is the ranking criterion of ecoPrimers.
func (p PairStats) Score() float32 { return p.Yule * p.Bs }

// ampRecord is an amplicon of a retained pair, in one sequence.
type ampRecord struct {
	pair     int32
	seq      int32
	hash     uint64
	tm1, tm2 float32 // NaN when unavailable
}

// ampHash hashes the amplicon between the two primers read from the side of
// the first primer of the pair (reverse complemented when the upstream match is
// the second primer), so that identical amplicons give the same value
// whatever their strand. Non ACGT bases hash as a fifth symbol.
func ampHash(raw []byte, from, to int, reverse bool) uint64 {
	const prime = 1099511628211
	h := uint64(14695981039346656037) ^ uint64(to-from)
	h *= prime
	if !reverse {
		for _, b := range raw[from:to] {
			h = (h ^ uint64(baseCode[b])) * prime
		}
		return h
	}
	for i := to - 1; i >= from; i-- {
		c := baseCode[raw[i]]
		if c < 4 {
			c = 3 - c
		}
		h = (h ^ uint64(c)) * prime
	}
	return h
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
func FindPairs(c *Corpus, taxa *Taxa, wm []WordMatches, p PairParams, model *obithermo.Model) ([]PairStats, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	if taxa == nil || len(taxa.Of) != len(c.Seqs) {
		return nil, fmt.Errorf("obiprimer: taxa do not match the corpus")
	}
	ps := selectPrimers(c, wm, p)
	if len(ps.primers) == 0 {
		return nil, nil
	}

	// Phase 1: which pairs amplify enough examples and few enough counterexamples.
	counts := c.Counts()
	acc := ps.countPairs(c, p)
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
	slices.SortFunc(kept, func(a, b pairKey) int {
		return cmp.Or(cmp.Compare(a.p1, b.p1), cmp.Compare(a.p2, b.p2),
			boolCmp(a.a1, b.a1), boolCmp(a.a2, b.a2))
	})
	index := make(map[pairKey]int32, len(kept))
	for i, k := range kept {
		index[k] = int32(i)
	}

	// Phase 2: the amplicons of the retained pairs only.
	records := ps.collectAmplicons(c, p, index, model)
	byPair := make([][]ampRecord, len(kept))
	for _, r := range records {
		byPair[r.pair] = append(byPair[r.pair], r)
	}

	var out []PairStats
	for i, k := range kept {
		st, ok := ps.evaluate(c, taxa, p, k, acc[k], byPair[i], model)
		if ok {
			out = append(out, st)
		}
	}
	slices.SortFunc(out, func(a, b PairStats) int {
		return cmp.Or(-cmp.Compare(a.Score(), b.Score()), cmp.Compare(a.Primer1, b.Primer1), cmp.Compare(a.Primer2, b.Primer2))
	})
	return out, nil
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

// collectAmplicons gathers, in parallel, one record per amplicon of a retained pair.
func (ps *primerSet) collectAmplicons(c *Corpus, p PairParams, index map[pairKey]int32, model *obithermo.Model) []ampRecord {
	lp := p.PrimerLength
	workers := obidefault.ParallelWorkers()
	parts := make([][]ampRecord, workers)

	// oriented[primer][0] is the reverse complement of the word, [1] the word itself.
	oriented := make([][2][]byte, len(ps.primers))
	for i, pr := range ps.primers {
		oriented[i][0] = []byte(orientedWord(pr.Word, lp, false))
		oriented[i][1] = []byte(orientedWord(pr.Word, lp, true))
	}
	nan := float32(math.NaN())

	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var site1, site2 [32]byte
			for s := range jobs {
				raw := c.Seqs[s].Sequence()
				n := len(raw)
				isExample := c.Roles[s] == Example
				ps.forEachOccurrence(c, s, p, func(o occurrence) {
					pi, ok := index[o.key]
					if !ok {
						return
					}
					rec := ampRecord{pair: pi, seq: int32(s), tm1: nan, tm2: nan}
					if isExample {
						rec.hash = ampHash(raw, int(o.up.pos)+lp, int(o.down.pos), o.swapped)
						if model != nil {
							fp := oriented[o.up.primer][b2i(!o.up.reverse)]
							rp := oriented[o.down.primer][b2i(o.down.reverse)]
							fsite, rsite := site1[:lp], site2[:lp]
							okF := windowAsIs(fsite, raw, int(o.up.pos), n, p.Circular)
							okR := windowRevComp(rsite, raw, int(o.down.pos), n, p.Circular)
							tmF, tmR := nan, nan
							if okF {
								if t, err := model.Tm(fp, fsite); err == nil {
									tmF = float32(t)
								}
							}
							if okR {
								if t, err := model.Tm(rp, rsite); err == nil {
									tmR = float32(t)
								}
							}
							if o.swapped {
								rec.tm1, rec.tm2 = tmR, tmF
							} else {
								rec.tm1, rec.tm2 = tmF, tmR
							}
						}
					}
					parts[w] = append(parts[w], rec)
				})
			}
		}()
	}
	for s := range c.Seqs {
		jobs <- s
	}
	close(jobs)
	wg.Wait()

	var all []ampRecord
	for _, part := range parts {
		all = append(all, part...)
	}
	return all
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

const upper = "ACGT"

// windowAsIs copies the len(dst) bases at pos as upper-case ACGT; it reports
// false if one of them is not A, C, G or T.
func windowAsIs(dst, raw []byte, pos, n int, circular bool) bool {
	for t := range dst {
		i := pos + t
		if circular {
			i %= n
		}
		c := baseCode[raw[i]]
		if c > 3 {
			return false
		}
		dst[t] = upper[c]
	}
	return true
}

// windowRevComp writes the reverse complement of the len(dst) bases at pos.
func windowRevComp(dst, raw []byte, pos, n int, circular bool) bool {
	l := len(dst)
	for t := range dst {
		i := pos + l - 1 - t
		if circular {
			i %= n
		}
		c := baseCode[raw[i]]
		if c > 3 {
			return false
		}
		dst[t] = upper[3-c]
	}
	return true
}

// evaluate computes the indices of one retained pair. It reports false when the
// pair must be dropped (no taxon amplified, or specificity under the threshold).
func (ps *primerSet) evaluate(c *Corpus, taxa *Taxa, p PairParams, k pairKey, a *pairAcc,
	recs []ampRecord, model *obithermo.Model) (PairStats, bool) {

	exTaxa := map[int32]struct{}{}
	coTaxa := map[int32]struct{}{}
	ampTaxon := map[uint64]int32{} // amplicon -> its taxon, or -2 when shared
	for _, r := range recs {
		t := taxa.Of[r.seq]
		if t < 0 {
			continue
		}
		if c.Roles[r.seq] != Example {
			coTaxa[t] = struct{}{}
			continue
		}
		exTaxa[t] = struct{}{}
		if prev, seen := ampTaxon[r.hash]; !seen {
			ampTaxon[r.hash] = t
		} else if prev != t {
			ampTaxon[r.hash] = -2
		}
	}
	if len(exTaxa) == 0 {
		return PairStats{}, false
	}
	identified := map[int32]struct{}{}
	for _, t := range ampTaxon {
		if t >= 0 {
			identified[t] = struct{}{}
		}
	}

	counts := c.Counts()
	st := PairStats{
		Examples:        int(a.examples),
		CounterExamples: int(a.counters),
		ExampleTaxa:     len(exTaxa),
		CounterTaxa:     len(coTaxa),
		Identified:      len(identified),
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
	st.Bc = float32(st.ExampleTaxa) / float32(taxa.ExampleTaxa)
	st.Bs = float32(st.Identified) / float32(st.ExampleTaxa)
	if st.Bs < float32(p.MinSpecificity) {
		return PairStats{}, false
	}

	lp := p.PrimerLength
	st.Primer1 = orientedWord(ps.primers[k.p1].Word, lp, k.a1)
	st.Primer2 = orientedWord(ps.primers[k.p2].Word, lp, k.a2)
	st.Good1, st.Good2 = ps.primers[k.p1].Good, ps.primers[k.p2].Good
	st.GC1, st.GC2 = obithermo.GCCount([]byte(st.Primer1)), obithermo.GCCount([]byte(st.Primer2))

	st.Tm1, st.Tm2, st.MinTm1, st.MinTm2 = math.NaN(), math.NaN(), math.NaN(), math.NaN()
	if model != nil {
		st.Tm1, _ = model.SelfTm([]byte(st.Primer1))
		st.Tm2, _ = model.SelfTm([]byte(st.Primer2))
		m1, m2 := float32(math.Inf(1)), float32(math.Inf(1))
		for _, r := range recs {
			if c.Roles[r.seq] != Example {
				continue
			}
			if r.tm1 == r.tm1 { // not NaN
				m1 = min(m1, r.tm1)
			}
			if r.tm2 == r.tm2 {
				m2 = min(m2, r.tm2)
			}
		}
		if !math.IsInf(float64(m1), 1) {
			st.MinTm1 = float64(m1)
		}
		if !math.IsInf(float64(m2), 1) {
			st.MinTm2 = float64(m2)
		}
	}

	if st.Primer2 < st.Primer1 {
		st.Primer1, st.Primer2 = st.Primer2, st.Primer1
		st.Good1, st.Good2 = st.Good2, st.Good1
		st.GC1, st.GC2 = st.GC2, st.GC1
		st.Tm1, st.Tm2 = st.Tm2, st.Tm1
		st.MinTm1, st.MinTm2 = st.MinTm2, st.MinTm1
	}
	return st, true
}
