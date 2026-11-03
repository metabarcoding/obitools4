package obiprimer

import (
	"cmp"
	"math"
	"slices"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obithermo"
)

// ampRecord is an amplicon of a retained pair, in one sequence.
type ampRecord struct {
	hash  uint64 // the barcode: see ampHash
	pair  int32
	taxon int32 // taxon of the sequence at the chosen rank, -1 when it has none
	role  Role
}

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

// amplicons holds what the second phase gathers on the retained pairs.
type amplicons struct {
	parts  [][]ampRecord // one slice per worker, each sorted by pair
	minTm1 []float32     // by pair: lowest Tm of the first primer over the example sites
	minTm2 []float32
	refs   map[int32]refHit
}

// pair returns the amplicons of one retained pair, as one slice per worker.
func (a *amplicons) pair(i int32) [][]ampRecord {
	out := make([][]ampRecord, 0, len(a.parts))
	byPair := func(r ampRecord, p int32) int { return cmp.Compare(r.pair, p) }
	for _, part := range a.parts {
		lo, _ := slices.BinarySearchFunc(part, i, byPair)
		hi, _ := slices.BinarySearchFunc(part, i+1, byPair)
		if hi > lo {
			out = append(out, part[lo:hi])
		}
	}
	return out
}

// collectOptions selects what collectAmplicons gathers.
type collectOptions struct {
	examplesOnly bool             // skip the counterexamples and the other sequences
	model        *obithermo.Model // melting temperatures of the example sites (nil: none)
	reference    bool             // record the amplicons on the reference sequence
}

// collectAmplicons gathers, in parallel, one record per amplicon of a retained
// pair, in the sequences of the corpus (all of them, or the examples only), and
// the lowest melting temperatures over the example sites. The records are not
// copied nor regrouped: each worker keeps its own, sorted by pair.
func (ps *primerSet) collectAmplicons(c *Corpus, taxa *Taxa, p PairParams, index map[pairKey]int32, opt collectOptions) *amplicons {
	model := opt.model
	lp := p.PrimerLength
	workers := obidefault.ParallelWorkers()
	npairs := len(index)

	res := &amplicons{
		parts:  make([][]ampRecord, workers),
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

	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			site1, site2 := make([]byte, lp), make([]byte, lp)
			lowTm1, lowTm2 := map[int32]float32{}, map[int32]float32{}
			for s := range jobs {
				raw := c.Seqs[s].Sequence()
				n := len(raw)
				role := c.Roles[s]
				if opt.examplesOnly && role != Example {
					continue
				}
				ps.forEachOccurrence(c, s, p, func(o occurrence) {
					pi, ok := index[o.key]
					if !ok {
						return
					}
					if opt.reference && p.HasReference && s == p.Reference {
						if old, seen := res.refs[pi]; !seen || o.up.pos < old.up.pos {
							res.refs[pi] = refHit{o.up, o.down, o.swapped}
						}
					}
					res.parts[w] = append(res.parts[w], ampRecord{
						hash:  ampHash(raw, int(o.up.pos)+lp, int(o.down.pos), o.swapped),
						pair:  pi,
						taxon: taxa.Of[s],
						role:  role,
					})

					if role != Example || model == nil {
						return
					}
					fp := oriented[o.up.primer][b2i(!o.up.reverse)]
					rp := oriented[o.down.primer][b2i(o.down.reverse)]
					tmF, tmR := inf, inf
					if windowAsIs(site1, raw, int(o.up.pos), n, p.Circular) {
						if t, err := model.Tm(fp, site1); err == nil && !math.IsNaN(t) {
							tmF = float32(t)
						}
					}
					if windowRevComp(site2, raw, int(o.down.pos), n, p.Circular) {
						if t, err := model.Tm(rp, site2); err == nil && !math.IsNaN(t) {
							tmR = float32(t)
						}
					}
					if o.swapped {
						tmF, tmR = tmR, tmF // slot 1 is the downstream primer
					}
					lowTm1[pi] = min(lowOr(lowTm1, pi, inf), tmF)
					lowTm2[pi] = min(lowOr(lowTm2, pi, inf), tmR)
				})
			}

			slices.SortFunc(res.parts[w], func(a, b ampRecord) int { return cmp.Compare(a.pair, b.pair) })
			tmMu.Lock()
			for pi, t := range lowTm1 {
				res.minTm1[pi] = min(res.minTm1[pi], t)
			}
			for pi, t := range lowTm2 {
				res.minTm2[pi] = min(res.minTm2[pi], t)
			}
			tmMu.Unlock()
		}()
	}
	for s := range c.Seqs {
		jobs <- s
	}
	close(jobs)
	wg.Wait()
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
