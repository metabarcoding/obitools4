package obikmer

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// ParseIUPACMasks converts an IUPAC nucleotide pattern into one 4 bits mask per
// position (bit 0 is A, bit 1 is C, bit 2 is G, bit 3 is T), the format used by
// KmerConstraints.Accepted and Rejected.
func ParseIUPACMasks(pattern string) ([]uint8, error) {
	masks := make([]uint8, len(pattern))
	for i := 0; i < len(pattern); i++ {
		m := obiseq.IUPACNucMask(pattern[i])
		if m == 0 {
			return nil, fmt.Errorf("invalid IUPAC nucleotide code %q at position %d of pattern %q",
				pattern[i], i+1, pattern)
		}
		masks[i] = m
	}
	return masks, nil
}

// RestrictAlphabet forbids every base that is not denoted by one of the IUPAC
// codes of alphabet. For instance "ACT" (or "H") forbids G.
func (c *KmerConstraints) RestrictAlphabet(alphabet string) error {
	masks, err := ParseIUPACMasks(alphabet)
	if err != nil {
		return err
	}
	var allowed uint8
	for _, m := range masks {
		allowed |= m
	}
	if allowed == 0 {
		return fmt.Errorf("empty alphabet")
	}
	for b := range 4 {
		if allowed>>uint(b)&1 == 0 {
			c.MaxCount[b] = 0
		}
	}
	return nil
}

// validPositions tells whether Accepted and Rejected, when set, have length k.
func (c *KmerConstraints) validPositions(k int) bool {
	return (c.Accepted == nil || len(c.Accepted) == k) &&
		(c.Rejected == nil || len(c.Rejected) == k)
}

// matchesAllMasks tells whether every base of the k-mer belongs to the set
// given for its position.
func matchesAllMasks(kmer uint64, k int, masks []uint8) bool {
	for i := k - 1; i >= 0; i-- {
		if masks[i]>>(kmer&3)&1 == 0 {
			return false
		}
		kmer >>= 2
	}
	return true
}

// Accepts tells whether a k-mer of size k satisfies every constraint.
// It is the predicate whose solutions IterConstrainedKmers enumerates, to be
// used to filter k-mers coming from elsewhere.
func (c *KmerConstraints) Accepts(kmer uint64, k int) bool {
	if k < 1 || k > 31 || !c.validPositions(k) {
		return false
	}

	var counts [4]int
	gc, run, longest, last := 0, 0, 0, -1
	for i := k - 1; i >= 0; i-- {
		b := int(kmer>>(2*uint(i))) & 3
		if c.Accepted != nil && c.Accepted[k-1-i]>>uint(b)&1 == 0 {
			return false
		}
		counts[b]++
		if c.MaxCount[b] >= 0 && counts[b] > c.MaxCount[b] {
			return false
		}
		if b == 1 || b == 2 {
			gc++
			if c.MaxGC >= 0 && gc > c.MaxGC {
				return false
			}
		}
		if b == last {
			run++
		} else {
			run = 1
		}
		last = b
		longest = max(longest, run)
		if c.MaxRun > 0 && run > c.MaxRun {
			return false
		}
	}

	if c.MinRun > 1 && longest < c.MinRun {
		return false
	}
	if c.Rejected != nil && matchesAllMasks(kmer, k, c.Rejected) {
		return false
	}
	return true
}

// HammingBound returns the sphere-packing (Hamming) upper bound on the number
// of words of length k over an alphabet of alphabetSize letters that can be
// pairwise at Hamming distance >= d. No larger family exists; a family of
// that size may not exist either.
func HammingBound(alphabetSize, k, d int) float64 {
	if d <= 1 {
		total := 1.0
		for i := 0; i < k; i++ {
			total *= float64(alphabetSize)
		}
		return total
	}
	if d%2 == 0 { // A(k, d) <= A(k-1, d-1)
		k--
		d--
	}
	if k < 1 {
		return 1
	}
	t := (d - 1) / 2
	total, ball, binom, pow := 1.0, 1.0, 1.0, 1.0
	for i := 0; i < k; i++ {
		total *= float64(alphabetSize)
	}
	for i := 1; i <= t && i <= k; i++ {
		binom = binom * float64(k-i+1) / float64(i)
		pow *= float64(alphabetSize - 1)
		ball += binom * pow
	}
	return total / ball
}
