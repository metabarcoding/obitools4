package obiprimer

import (
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obithermo"
)

// DefaultAmpBudget is the number of amplicon records (12 bytes each) phase 2
// holds at once: 20 million, 240 MB. The pairs are processed in batches whose
// amplicons fit in it. With the usual specificity threshold a search needs one
// batch; with none (-T 0) no pair is rejected early and there are many more
// amplicons than any memory could hold.
const DefaultAmpBudget = 20_000_000

func (p PairParams) release() {
	if p.Release != nil {
		p.Release()
	}
}

func (p PairParams) ampBudget() int {
	if p.AmpBudget > 0 {
		return p.AmpBudget
	}
	return DefaultAmpBudget
}

// batchIndex maps the pairs kept[lo:hi] to their number in the batch.
func batchIndex(kept []pairKey, lo, hi int) map[pairKey]int32 {
	index := make(map[pairKey]int32, hi-lo)
	for i := lo; i < hi; i++ {
		index[kept[i]] = int32(i - lo)
	}
	return index
}

// parallelFor calls fn(i) for i in [0,n) on obidefault.ParallelWorkers() goroutines.
func parallelFor(n int, fn func(worker, i int)) {
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := range obidefault.ParallelWorkers() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				fn(w, i)
			}
		}()
	}
	for i := 0; i < n; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
}

// bound is what phase 2a learns about each retained pair (indexed like kept).
type bound struct {
	targets    []int // target taxa amplified
	identified []int // upper bound of the target taxa identified (lenientIdentified)
	minTm1     []float32
	minTm2     []float32
	refs       map[int32]refHit
}

// boundPairs is phase 2a: from the amplicons of the target sequences alone, it
// gives each retained pair the upper bound of its Bs, the lowest melting
// temperatures and its site on the reference. The pairs go by batches whose
// amplicons fit in the budget; they are counted first, so that nothing is resized.
func (ps *primerSet) boundPairs(c *Corpus, taxa *Taxa, p PairParams, kept []pairKey,
	model *obithermo.Model) *bound {

	plan := ps.planAmplicons(c, p, batchIndex(kept, 0, len(kept)), true)
	res := &bound{
		targets:    make([]int, len(kept)),
		identified: make([]int, len(kept)),
		minTm1:     make([]float32, len(kept)),
		minTm2:     make([]float32, len(kept)),
		refs:       map[int32]refHit{},
	}
	resolvers := make([]*resolver, obidefault.ParallelWorkers())
	for i := range resolvers {
		resolvers[i] = newResolver()
	}

	for _, b := range ampBatches(plan.total(), p.ampBudget()) {
		lo, hi := b[0], b[1]
		p.release() // the previous batch, or the stage before, is garbage
		store := newAmpStore(plan, lo, hi)
		sites := ps.collectAmplicons(c, taxa, p, batchIndex(kept, lo, hi), plan, store,
			collectOptions{examplesOnly: true, model: model, reference: true})
		store.check()
		parallelFor(hi-lo, func(w, i int) {
			res.targets[lo+i], res.identified[lo+i] = resolvers[w].lenientIdentified(store.pair(i))
		})
		copy(res.minTm1[lo:hi], sites.minTm1)
		copy(res.minTm2[lo:hi], sites.minTm2)
		for i, h := range sites.refs {
			res.refs[int32(lo)+i] = h
		}
	}
	return res
}

// evaluatePairs is phase 2b: for the surviving pairs (indices in kept), it
// gathers the amplicons of every sequence, whatever its role, by batches whose
// amplicons fit in the budget, and evaluates the pairs. The result is indexed
// like survivors; a nil entry is a pair dropped by the specificity threshold.
func (ps *primerSet) evaluatePairs(c *Corpus, taxa *Taxa, p PairParams, kept []pairKey, survivors []int,
	acc map[pairKey]*pairAcc, first *bound, model *obithermo.Model) []*PairStats {

	results := make([]*PairStats, len(survivors))
	if len(survivors) == 0 {
		return results
	}
	keys := make([]pairKey, len(survivors))
	for j, i := range survivors {
		keys[j] = kept[i]
	}
	// the amplicons of every sequence are counted first: the store is allocated
	// at its exact size, and the batches are cut on the exact counts
	plan := ps.planAmplicons(c, p, batchIndex(keys, 0, len(keys)), false)

	counts := c.Counts()
	var refBases []byte
	if p.HasReference {
		refBases, _ = c.bases(p.Reference, nil)
	}
	resolvers := make([]*resolver, obidefault.ParallelWorkers())
	for i := range resolvers {
		resolvers[i] = newResolver()
	}
	for _, b := range ampBatches(plan.total(), p.ampBudget()) {
		lo, hi := b[0], b[1]
		p.release()
		store := newAmpStore(plan, lo, hi)
		ps.collectAmplicons(c, taxa, p, batchIndex(keys, lo, hi), plan, store, collectOptions{})
		store.check()
		parallelFor(hi-lo, func(w, n int) {
			j := lo + n
			i := survivors[j]
			k := keys[j]
			st, ok := ps.evaluate(taxa, counts, p, k, acc[k], resolvers[w].resolve(store.pair(n)),
				first.minTm1[i], first.minTm2[i], model)
			if !ok {
				return
			}
			if h, found := first.refs[int32(i)]; found {
				// evaluate orders the primers by string; the amplicon is read from the first one
				s1 := orientedWord(ps.primers[k.p1].Word, p.PrimerLength, k.a1)
				s2 := orientedWord(ps.primers[k.p2].Word, p.PrimerLength, k.a2)
				st.RefLocation, st.RefAmplicon = referenceSites(refBases, h, p, s2 < s1)
			}
			results[j] = &st
		})
	}
	return results
}
