package obiprimer

import (
	"math"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obithermo"
)

// refHit is an amplicon of a retained pair on the reference sequence.
type refHit struct {
	up, down seqMatch
	swapped  bool
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

// eachAmplicon runs forEachOccurrence over the sequences of the corpus (the
// examples only, or all of them) in parallel, and calls visit for every
// amplicon of a pair of index. Every worker gets its own visit function, from
// newVisit, and done is called once the worker has no more sequence.
//
// Without a plan the sequences are handed to the workers as they get free, and
// the assignment is recorded in record. With one, every worker scans the
// sequences that the plan gave it.
func (ps *primerSet) eachAmplicon(c *Corpus, p PairParams, index map[pairKey]int32, examplesOnly bool, plan *ampPlan, record []uint16,
	newVisit func(worker int) (visit func(s int, raw []byte, o occurrence, pair int32), done func())) {

	workers := obidefault.ParallelWorkers()

	// only the matches of the primers of the pairs of index can make an amplicon
	// of these pairs, and dropping the others spares most of the comparisons
	wanted := make([]bool, len(ps.primers))
	for k := range index {
		wanted[k.p1], wanted[k.p2] = true, true
	}
	type scratch struct {
		ms    []seqMatch
		bases []byte
	}
	scan := func(sc *scratch, visit func(int, []byte, occurrence, int32), s int) {
		if examplesOnly && c.Roles[s] != Example {
			return
		}
		ms := sc.ms[:0]
		for _, m := range ps.perSeq[s] {
			if wanted[m.primer] {
				ms = append(ms, m)
			}
		}
		sc.ms = ms
		if len(ms) < 2 {
			return // an amplicon needs two primer sites
		}
		var raw []byte
		raw, sc.bases = c.bases(s, sc.bases)
		ps.forEachOccurrenceIn(raw, ms, p, func(o occurrence) {
			if pi, ok := index[o.key]; ok {
				visit(s, raw, o, pi)
			}
		})
	}

	var wg sync.WaitGroup
	if plan != nil {
		mine := make([][]int, workers)
		for s, w := range plan.owner {
			mine[w] = append(mine[w], s)
		}
		for w := range workers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				visit, done := newVisit(w)
				var sc scratch
				for _, s := range mine[w] {
					scan(&sc, visit, s)
				}
				if done != nil {
					done()
				}
			}()
		}
		wg.Wait()
		return
	}

	jobs := make(chan int)
	for w := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			visit, done := newVisit(w)
			var sc scratch
			for s := range jobs {
				record[s] = uint16(w)
				scan(&sc, visit, s)
			}
			if done != nil {
				done()
			}
		}()
	}
	for s := range c.Roles {
		jobs <- s
	}
	close(jobs)
	wg.Wait()
}

// planAmplicons counts the amplicons of every pair of index in the sequences
// of the corpus (all of them, or the examples only), by worker.
func (ps *primerSet) planAmplicons(c *Corpus, p PairParams, index map[pairKey]int32, examplesOnly bool) *ampPlan {
	workers := obidefault.ParallelWorkers()
	plan := &ampPlan{owner: make([]uint16, c.Len()), count: make([][]uint32, workers)}
	for w := range plan.count {
		plan.count[w] = make([]uint32, len(index))
	}
	ps.eachAmplicon(c, p, index, examplesOnly, nil, plan.owner, func(w int) (func(int, []byte, occurrence, int32), func()) {
		row := plan.count[w]
		return func(_ int, _ []byte, _ occurrence, pi int32) { row[pi]++ }, nil
	})
	return plan
}

// ampSites is what the gathering of the amplicons of the targets learns besides
// the records.
type ampSites struct {
	minTm1 []float32 // by pair: lowest Tm of the first primer over the example sites
	minTm2 []float32
	refs   map[int32]refHit // by pair: the amplicon on the reference sequence
}

// collectOptions selects what collectAmplicons gathers besides the records.
type collectOptions struct {
	examplesOnly bool             // skip the counterexamples and the other sequences
	model        *obithermo.Model // melting temperatures of the example sites (nil: none)
	reference    bool             // record the amplicons on the reference sequence
}

// collectAmplicons writes into store, in parallel, one record per amplicon of
// the pairs of index (whose values are the pair numbers in the store) and
// returns the lowest melting temperatures over the example sites. The store
// was sized from the plan, which says which worker scans which sequence; the
// batch only holds the pairs of index, the plan more.
func (ps *primerSet) collectAmplicons(c *Corpus, taxa *Taxa, p PairParams, index map[pairKey]int32,
	plan *ampPlan, store *ampStore, opt collectOptions) *ampSites {

	model := opt.model
	lp := p.PrimerLength
	npairs := len(index)

	res := &ampSites{
		minTm1: make([]float32, npairs),
		minTm2: make([]float32, npairs),
		refs:   map[int32]refHit{}, // written by the single worker that reads the reference
	}
	inf := float32(math.Inf(1))
	for i := range npairs {
		res.minTm1[i], res.minTm2[i] = inf, inf
	}
	var tmMu sync.Mutex

	// oriented[primer][0] is the reverse complement of the word, [1] the word itself.
	oriented := make([][2][]byte, len(ps.primers))
	for i, pr := range ps.primers {
		oriented[i][0] = []byte(orientedWord(pr.Word, lp, false))
		oriented[i][1] = []byte(orientedWord(pr.Word, lp, true))
	}

	ps.eachAmplicon(c, p, index, opt.examplesOnly, plan, nil, func(w int) (func(int, []byte, occurrence, int32), func()) {
		site1, site2 := make([]byte, lp), make([]byte, lp)
		lowTm1, lowTm2 := map[int32]float32{}, map[int32]float32{}

		visit := func(s int, raw []byte, o occurrence, pi int32) {
			n := len(raw)
			role := c.Roles[s]
			if opt.reference && p.HasReference && s == p.Reference {
				if old, seen := res.refs[pi]; !seen || o.up.pos() < old.up.pos() {
					res.refs[pi] = refHit{o.up, o.down, o.swapped}
				}
			}
			store.add(w, pi, ampHash(raw, int(o.up.pos())+lp, int(o.down.pos()), o.swapped), packTaxon(taxa.Of[s], role))

			if role != Example || model == nil {
				return
			}
			fp := oriented[o.up.primer][b2i(!o.up.reverse())]
			rp := oriented[o.down.primer][b2i(o.down.reverse())]
			tmF, tmR := inf, inf
			if windowAsIs(site1, raw, int(o.up.pos()), n, p.Circular) {
				if t, err := model.Tm(fp, site1); err == nil && !math.IsNaN(t) {
					tmF = float32(t)
				}
			}
			if windowRevComp(site2, raw, int(o.down.pos()), n, p.Circular) {
				if t, err := model.Tm(rp, site2); err == nil && !math.IsNaN(t) {
					tmR = float32(t)
				}
			}
			if o.swapped {
				tmF, tmR = tmR, tmF // slot 1 is the downstream primer
			}
			lowTm1[pi] = min(lowOr(lowTm1, pi, inf), tmF)
			lowTm2[pi] = min(lowOr(lowTm2, pi, inf), tmR)
		}
		done := func() {
			tmMu.Lock()
			for pi, t := range lowTm1 {
				res.minTm1[pi] = min(res.minTm1[pi], t)
			}
			for pi, t := range lowTm2 {
				res.minTm2[pi] = min(res.minTm2[pi], t)
			}
			tmMu.Unlock()
		}
		return visit, done
	})
	return res
}

func lowOr(m map[int32]float32, k int32, def float32) float32 {
	if v, ok := m[k]; ok {
		return v
	}
	return def
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
