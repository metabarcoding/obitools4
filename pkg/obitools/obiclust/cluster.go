package obiclust

import (
	"sort"
	"sync"
	"sync/atomic"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
)

// clustering is the state of the star clustering of the distinct sequences,
// taken in their sorted order.
//
// The reference algorithm is sequential: each sequence is compared with the
// centers created so far, in the order of their creation, and is attached to
// the first one it matches (or to the best one in exact mode), or becomes a
// new center. Its result is reproduced exactly while the comparisons are run
// in parallel, by processing the sequences by blocks:
//
//   - phase A (parallel): every sequence of the block is compared with the
//     centers existing when the block starts. This is independent from one
//     sequence to the other, a sequence of the block cannot change the centers;
//   - phase B (sequential): the sequences matching no center are then taken
//     in their order. Each one is compared with the centers created earlier in
//     the block, and becomes a new center if it still matches nothing. The
//     new centers come after the old ones in the order of the centers, so
//     the result is the one of the sequential algorithm;
//   - phase C (parallel, exact mode only): the sequences attached in phase A
//     are compared with the new centers that precede them, the best score of
//     the whole list of centers being searched for.
//
// The result does not depend on the size of the blocks, which only trades the
// sequential work of phase B against the synchronizations.
type clustering struct {
	entries []entry
	sc      *scoring
	fast    bool
	workers int
	// bounds of the size of the blocks, 0 for the defaults (tests only)
	blockMin, blockMax int

	centers []int // indices of the centers, in the order of creation
	set     *centerSet
	fresh   []center // the centers created in the current block

	center []int     // for each entry, the index of its center, -1 if it has none
	score  []float64 // for each entry, the score with its center
	bufs   []obialign.BandedLCSBuffer
}

// result of the comparison of one entry with the list of centers
type attachment struct {
	center int // index of the center, -1 if none
	score  float64
}

func newClustering(entries []entry, sc *scoring, fast bool, workers int) *clustering {
	c := &clustering{
		entries: entries,
		sc:      sc,
		fast:    fast,
		workers: max(workers, 1),
		center:  make([]int, len(entries)),
		score:   make([]float64, len(entries)),
		bufs:    make([]obialign.BandedLCSBuffer, max(workers, 1)),
	}
	for i := range c.center {
		c.center[i] = -1
	}
	c.set = newCenterSet(simdBlockWidth())
	return c
}

func (c *clustering) addCenter(i int) {
	c.fresh = append(c.fresh, c.set.add(c.entries[i], i))
	c.centers = append(c.centers, i)
	c.center[i] = i
	c.score[i] = c.sc.centerScore(c.entries[i].length)
}

// compare looks for the best center of entry i among the given ones. The
// attachment `from` is the best one already known. The search stops at the
// first match in fast mode.
func (c *clustering) compare(i int, cents []center, from attachment, buf *obialign.BandedLCSBuffer) attachment {
	best := from
	for k := range cents {
		s := c.sc.pairScore(&c.entries[i], &cents[k].entry, -1, buf)
		if c.sc.selects(s, best.score) {
			best = attachment{cents[k].idx, s}
			if c.fast {
				break
			}
		}
	}
	return best
}

// parallelFor runs f(i, buf) for i in [from, to) on the workers.
func (c *clustering) parallelFor(from, to int, f func(i int, buf *obialign.BandedLCSBuffer)) {
	n := to - from
	if c.workers == 1 || n == 1 {
		for i := from; i < to; i++ {
			f(i, &c.bufs[0])
		}
		return
	}

	var next atomic.Int64
	next.Store(int64(from))
	var wg sync.WaitGroup
	for w := 0; w < min(c.workers, n); w++ {
		wg.Add(1)
		go func(buf *obialign.BandedLCSBuffer) {
			defer wg.Done()
			for {
				i := int(next.Add(1)) - 1
				if i >= to {
					return
				}
				f(i, buf)
			}
		}(&c.bufs[w])
	}
	wg.Wait()
}

// run clusters all the entries. progress is called with the number of
// entries processed.
func (c *clustering) run(progress func(done int)) {
	n := len(c.entries)
	if n == 0 {
		return
	}

	minBlock, maxBlock := 2*c.workers, 256*c.workers
	if c.blockMax > 0 {
		minBlock, maxBlock = c.blockMin, c.blockMax
	}
	block := minBlock
	att := make([]attachment, maxBlock)
	inA := make([]bool, maxBlock) // attached to a center existing before the block
	var newCenters []int          // indices of the centers created in the block

	for pos := 0; pos < n; {
		end := min(n, pos+block)
		c.fresh = c.fresh[:0]
		att, inA := att[:end-pos], inA[:end-pos]

		// phase A
		c.parallelFor(pos, end, func(i int, buf *obialign.BandedLCSBuffer) {
			att[i-pos] = c.scan(i, buf)
			inA[i-pos] = att[i-pos].center >= 0
		})

		// phase B
		unmatched := 0
		newCenters = newCenters[:0]
		for i := pos; i < end; i++ {
			a := &att[i-pos]
			if inA[i-pos] {
				continue
			}
			unmatched++
			*a = c.compare(i, c.fresh, *a, &c.bufs[0])
			if a.center < 0 && !c.entries[i].ambiguous {
				newCenters = append(newCenters, i)
				c.addCenter(i)
			}
		}

		// phase C
		if !c.fast && len(newCenters) > 0 {
			c.parallelFor(pos, end, func(i int, buf *obialign.BandedLCSBuffer) {
				if inA[i-pos] {
					before := c.fresh[:sort.SearchInts(newCenters, i)]
					att[i-pos] = c.compare(i, before, att[i-pos], buf)
				}
			})
		}

		for i := pos; i < end; i++ {
			if a := att[i-pos]; a.center >= 0 {
				c.center[i] = a.center
				c.score[i] = c.sc.memberScore(a.score)
			}
		}

		// Size of the next block. Phase B is sequential: for a block of b
		// sequences of which a fraction u matches no old center and a fraction
		// n becomes new centers, it makes about b*u*b*n/2 comparisons, against
		// b*K/workers in the time of phase A (K old centers). The block is
		// sized so that phase B stays around a tenth of phase A.
		if fu, fn := float64(unmatched)/float64(end-pos), float64(len(newCenters))/float64(end-pos); fu*fn > 0 {
			target := 0.2 * float64(len(c.centers)) / (float64(c.workers) * fu * fn)
			block = int(min(float64(maxBlock), max(float64(minBlock), target)))
		} else {
			block = maxBlock
		}
		pos = end
		progress(pos)
	}
}
