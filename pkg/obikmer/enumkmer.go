package obikmer

import "iter"

// KmerConstraints describes lexicographic properties a k-mer must satisfy.
//
// MaxCount is indexed by the 2-bit base code (A=0, C=1, G=2, T=3):
// a negative value means unlimited, 0 forbids the base altogether
// (so MaxCount[2] = 0 restricts the alphabet to A, C, T).
//
// MaxRun is the maximum number of consecutive identical bases;
// a value <= 0 means unlimited.
//
// The zero value forbids every base: use NoKmerConstraints as a starting point.
type KmerConstraints struct {
	MaxCount [4]int
	MaxRun   int
}

// NoKmerConstraints returns constraints accepting every k-mer.
func NoKmerConstraints() KmerConstraints {
	return KmerConstraints{MaxCount: [4]int{-1, -1, -1, -1}}
}

// IterConstrainedKmers returns an iterator over all k-mers of size k
// satisfying the constraints c, encoded as in EncodeKmer, in increasing
// numerical order.
//
// K-mers are built base by base and a branch is abandoned as soon as a
// constraint is violated (both MaxCount and MaxRun are monotone: a violating
// prefix cannot be repaired by extension), so the cost is proportional to the
// number of explored valid prefixes, not to 4^k.
//
// k must be between 1 and 31, otherwise the iterator is empty.
//
// Example:
//
//	c := NoKmerConstraints()
//	c.MaxCount[2] = 0 // no G
//	c.MaxRun = 3
//	for kmer := range IterConstrainedKmers(8, c) { ... }
func IterConstrainedKmers(k int, c KmerConstraints) iter.Seq[uint64] {
	return func(yield func(uint64) bool) {
		if k < 1 || k > 31 {
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

		var walk func(depth int, kmer uint64, last, run int) bool
		walk = func(depth int, kmer uint64, last, run int) bool {
			if depth == k {
				return yield(kmer)
			}
			for b := range 4 {
				if c.MaxCount[b] >= 0 && counts[b] >= c.MaxCount[b] {
					continue
				}
				r := 1
				if b == last {
					r = run + 1
				}
				if c.MaxRun > 0 && r > c.MaxRun {
					continue
				}
				counts[b]++
				ok := walk(depth+1, kmer<<2|uint64(b), b, r)
				counts[b]--
				if !ok {
					return false
				}
			}
			return true
		}

		walk(0, 0, -1, 0)
	}
}
