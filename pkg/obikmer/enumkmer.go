package obikmer

import "iter"

// KmerConstraints describes lexicographic properties a k-mer must satisfy.
//
// MaxCount is indexed by the 2-bit base code (A=0, C=1, G=2, T=3):
// a negative value means unlimited, 0 forbids the base altogether
// (so MaxCount[2] = 0 restricts the alphabet to A, C, T).
//
// MaxGC is the maximum number of C or G bases; a negative value means
// unlimited.
//
// MaxRun is the maximum number of consecutive identical bases;
// a value <= 0 means unlimited. MinRun requires at least one run of that
// length; a value <= 1 means no requirement.
//
// Accepted and Rejected are optional per-position sets of bases (4 bits masks,
// bit b set when the base of code b is in the set, see ParseIUPACMasks).
// When set, their length must be k. A k-mer is accepted only if each of its
// bases belongs to the Accepted set of its position, and is refused if every
// one of its bases belongs to the Rejected set of its position.
//
// The zero value forbids every base: use NoKmerConstraints as a starting point.
type KmerConstraints struct {
	MaxCount [4]int
	MaxGC    int
	MaxRun   int
	MinRun   int
	Accepted []uint8
	Rejected []uint8
}

// NoKmerConstraints returns constraints accepting every k-mer.
func NoKmerConstraints() KmerConstraints {
	return KmerConstraints{MaxCount: [4]int{-1, -1, -1, -1}, MaxGC: -1}
}

// IterConstrainedKmers returns an iterator over all k-mers of size k
// satisfying the constraints c, encoded as in EncodeKmer, in increasing
// numerical order.
//
// K-mers are built base by base and a branch is abandoned as soon as a
// constraint is violated or cannot be met any more (MaxCount, MaxGC, MaxRun
// and Accepted are monotone: a violating prefix cannot be repaired by
// extension; MinRun is checked against the positions left), so the cost is
// proportional to the number of explored prefixes, not to 4^k. Rejected is
// checked on complete k-mers.
//
// k must be between 1 and 31, otherwise the iterator is empty. It is also
// empty if Accepted or Rejected is set with a length different from k.
//
// Example:
//
//	c := NoKmerConstraints()
//	c.MaxCount[2] = 0 // no G
//	c.MaxRun = 3
//	for kmer := range IterConstrainedKmers(8, c) { ... }
func IterConstrainedKmers(k int, c KmerConstraints) iter.Seq[uint64] {
	return func(yield func(uint64) bool) {
		if k < 1 || k > 31 || !c.validPositions(k) {
			return
		}

		// Nothing to enumerate if the per-base caps cannot even sum to k.
		capacity := 0
		for _, m := range c.MaxCount {
			if m < 0 || m >= k {
				capacity += k
			} else {
				capacity += m
			}
		}
		if capacity < k {
			return
		}

		var counts [4]int
		gc := 0

		var walk func(depth int, kmer uint64, last, run, longest int) bool
		walk = func(depth int, kmer uint64, last, run, longest int) bool {
			if depth == k {
				if c.Rejected != nil && matchesAllMasks(kmer, k, c.Rejected) {
					return true
				}
				return yield(kmer)
			}
			for b := range 4 {
				if c.MaxCount[b] >= 0 && counts[b] >= c.MaxCount[b] {
					continue
				}
				if c.Accepted != nil && c.Accepted[depth]>>uint(b)&1 == 0 {
					continue
				}
				isGC := b == 1 || b == 2
				if isGC && c.MaxGC >= 0 && gc >= c.MaxGC {
					continue
				}
				r := 1
				if b == last {
					r = run + 1
				}
				if c.MaxRun > 0 && r > c.MaxRun {
					continue
				}
				l := max(longest, r)
				if c.MinRun > 1 && l < c.MinRun && r+(k-depth-1) < c.MinRun {
					continue
				}
				counts[b]++
				if isGC {
					gc++
				}
				ok := walk(depth+1, kmer<<2|uint64(b), b, r, l)
				counts[b]--
				if isGC {
					gc--
				}
				if !ok {
					return false
				}
			}
			return true
		}

		walk(0, 0, -1, 0, 0)
	}
}
