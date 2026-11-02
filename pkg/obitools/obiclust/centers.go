package obiclust

import (
	"math"
	"sort"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
)

// center is a copy of the entry of a center, with its own table: scanning
// the centers of a list then reads contiguous memory.
type center struct {
	entry
	idx  int // index of the entry in the dataset
	rank int // rank of the center in the order of creation
}

const tabChunk = 1024

// A bucket holds the centers of one length, in the order of creation, and
// their tables by blocks (see tableBlock).
type bucket struct {
	cents  []center
	blocks []tableBlock
}

// centerSet stores the centers by length. A sequence can only be attached to a
// center whose length is compatible with its own (see lengthCompatible), so
// it only visits the buckets of these lengths, instead of all the centers.
type centerSet struct {
	width int // centers per block of transposed tables, 0 if unused
	byLen map[int]*bucket
	lens  []int      // lengths that have a bucket, increasing
	tabs  [][]table4 // chunks holding the tables of the centers
	n     int
}

func newCenterSet(width int) *centerSet {
	return &centerSet{width: width, byLen: make(map[int]*bucket)}
}

// add stores a copy of the entry e, which is the entry idx of the dataset, as
// the next center, and returns it.
func (s *centerSet) add(e entry, idx int) center {
	if s.n%tabChunk == 0 {
		s.tabs = append(s.tabs, make([]table4, tabChunk))
	}
	tab := &s.tabs[len(s.tabs)-1][s.n%tabChunk]
	if e.table != nil {
		*tab = *e.table
	}
	e.table = tab

	b := s.byLen[e.length]
	if b == nil {
		b = new(bucket)
		s.byLen[e.length] = b
		s.lens = append(s.lens, e.length)
		sort.Ints(s.lens)
	}
	if w := s.width; w > 0 && e.table != nil {
		if len(b.cents)%w == 0 {
			b.blocks = append(b.blocks, newTableBlock(w))
		}
		b.blocks[len(b.blocks)-1].set(w, len(b.cents)%w, tab)
	}

	ctr := center{entry: e, idx: idx, rank: s.n}
	b.cents = append(b.cents, ctr)
	s.n++
	return ctr
}

// scan compares entry i with all the centers of the set, and returns the
// first one it matches (in fast mode) or the best one (in exact mode, the
// first of them in case of tie), exactly as compare would on the list of the
// centers in the order of creation.
//
// Only the buckets of compatible lengths are visited: the other centers would
// give the worst score. In each bucket the first match is the one of lowest
// rank, so the first match of the whole set is the one of lowest rank among
// the buckets. In exact mode, the best of each bucket is compared, ties going
// to the lowest rank.
func (c *clustering) scan(i int, buf *obialign.BandedLCSBuffer) attachment {
	q := &c.entries[i]
	set := c.set
	width := set.width
	useBlocks := width > 0 && q.sparse != nil

	best := attachment{-1, c.sc.worst}
	bestRank := math.MaxInt
	var common [64]uint8

	for _, length := range set.lens {
		if !c.sc.lengthCompatible(q.length, length) {
			continue
		}
		b := set.byLen[length]
		local := attachment{-1, c.sc.worst}
		localRank := 0

		for k := range b.cents {
			ctr := &b.cents[k]
			if c.fast && ctr.rank > bestRank {
				break // a center of lower rank has been found in another bucket
			}
			cw := -1
			if useBlocks {
				if k%width == 0 {
					b.blocks[k/width].common(width, q.sparse, common[:])
				}
				cw = int(common[k%width])
			}
			s := c.sc.pairScore(q, &ctr.entry, cw, buf)
			if c.sc.selects(s, local.score) {
				local, localRank = attachment{ctr.idx, s}, ctr.rank
				if c.fast {
					break
				}
			}
		}

		if local.center < 0 {
			continue
		}
		if c.fast {
			if localRank < bestRank {
				best, bestRank = local, localRank
			}
		} else if best.center < 0 || c.sc.beats(local.score, best.score) ||
			(local.score == best.score && localRank < bestRank) {
			best, bestRank = local, localRank
		}
	}
	return best
}
