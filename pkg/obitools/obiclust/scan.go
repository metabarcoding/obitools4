package obiclust

import (
	"math"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	sim "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obisimilarity"
)

// scan compares entry i with all the centers of the index, and returns the
// first one it matches (in fast mode) or the best one (in exact mode, the
// first of them in case of tie), exactly as compare would on the list of the
// centers in the order of creation.
//
// Only the buckets of compatible lengths are visited: the other centers would
// not match. In each bucket the first match is the one of lowest rank, so the
// first match of the whole index is the one of lowest rank among the buckets.
// In exact mode, the best of each bucket is compared, ties going to the
// lowest rank.
func (c *clustering) scan(i int, buf *obialign.BandedLCSBuffer) attachment {
	q := &c.entries[i]

	best := attachment{-1, 0}
	bestRank := math.MaxInt
	var commons sim.Commons

	for _, length := range c.index.Lengths() {
		if !c.sc.LengthCompatible(q.Length, length) {
			continue
		}
		b := c.index.Bucket(length)
		commons.Reset()
		local := attachment{-1, 0}
		localRank := 0

		for k := range b.Members {
			m := &b.Members[k]
			if c.fast && m.Rank > bestRank {
				break // a center of lower rank has been found in another bucket
			}
			s, ok := c.sc.PairScore(q, &m.Entry, b.Common(k, q.Sparse, &commons), buf)
			if c.selects(s, ok, local) {
				local, localRank = attachment{m.ID, s}, m.Rank
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
		} else if best.center < 0 || c.sc.Beats(local.score, best.score) ||
			(local.score == best.score && localRank < bestRank) {
			best, bestRank = local, localRank
		}
	}
	return best
}
