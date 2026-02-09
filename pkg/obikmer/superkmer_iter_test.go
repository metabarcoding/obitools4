package obikmer

import (
	"testing"
)

func TestIterSuperKmers(t *testing.T) {
	seq := []byte("ACGTACGTGGGGAAAA")
	k := 5
	m := 3

	count := 0
	for sk := range IterSuperKmers(seq, k, m) {
		count++
		t.Logf("SuperKmer %d: Minimizer=%d, Start=%d, End=%d, Seq=%s",
			count, sk.Minimizer, sk.Start, sk.End, string(sk.Sequence))

		// Verify sequence boundaries
		if sk.Start < 0 || sk.End > len(seq) {
			t.Errorf("Invalid boundaries: Start=%d, End=%d, seqLen=%d",
				sk.Start, sk.End, len(seq))
		}

		// Verify sequence content
		if string(sk.Sequence) != string(seq[sk.Start:sk.End]) {
			t.Errorf("Sequence mismatch: expected %s, got %s",
				string(seq[sk.Start:sk.End]), string(sk.Sequence))
		}
	}

	if count == 0 {
		t.Error("No super k-mers extracted")
	}

	t.Logf("Total super k-mers extracted: %d", count)
}

func TestIterSuperKmersVsSlice(t *testing.T) {
	seq := []byte("ACGTACGTGGGGAAAAACGTACGT")
	k := 7
	m := 4

	// Extract using slice version
	sliceResult := ExtractSuperKmers(seq, k, m, nil)

	// Extract using iterator version
	var iterResult []SuperKmer
	for sk := range IterSuperKmers(seq, k, m) {
		iterResult = append(iterResult, sk)
	}

	// Compare counts
	if len(sliceResult) != len(iterResult) {
		t.Errorf("Different number of super k-mers: slice=%d, iter=%d",
			len(sliceResult), len(iterResult))
	}

	// Compare each super k-mer
	for i := 0; i < len(sliceResult) && i < len(iterResult); i++ {
		slice := sliceResult[i]
		iter := iterResult[i]

		if slice.Minimizer != iter.Minimizer {
			t.Errorf("SuperKmer %d: different minimizers: slice=%d, iter=%d",
				i, slice.Minimizer, iter.Minimizer)
		}

		if slice.Start != iter.Start || slice.End != iter.End {
			t.Errorf("SuperKmer %d: different boundaries: slice=[%d:%d], iter=[%d:%d]",
				i, slice.Start, slice.End, iter.Start, iter.End)
		}

		if string(slice.Sequence) != string(iter.Sequence) {
			t.Errorf("SuperKmer %d: different sequences: slice=%s, iter=%s",
				i, string(slice.Sequence), string(iter.Sequence))
		}
	}
}

// TestSuperKmerMinimizerBijection validates the intrinsic property that
// a super k-mer sequence has one and only one minimizer (bijection property).
// This test ensures that:
// 1. All k-mers in a super k-mer share the same minimizer
// 2. Two identical super k-mer sequences must have the same minimizer
func TestSuperKmerMinimizerBijection(t *testing.T) {
	testCases := []struct {
		name string
		seq  []byte
		k    int
		m    int
	}{
		{
			name: "simple sequence",
			seq:  []byte("ACGTACGTACGTACGTACGTACGTACGTACGT"),
			k:    21,
			m:    11,
		},
		{
			name: "homopolymer blocks",
			seq:  []byte("AAAACCCCGGGGTTTTAAAACCCCGGGGTTTT"),
			k:    21,
			m:    11,
		},
		{
			name: "complex sequence",
			seq:  []byte("ATCGATCGATCGATCGATCGATCGATCGATCG"),
			k:    15,
			m:    7,
		},
		{
			name: "longer sequence",
			seq:  []byte("ACGTACGTGGGGAAAAACGTACGTTTTTCCCCACGTACGT"),
			k:    13,
			m:    7,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Map to track sequence -> minimizer
			seqToMinimizer := make(map[string]uint64)

			for sk := range IterSuperKmers(tc.seq, tc.k, tc.m) {
				seqStr := string(sk.Sequence)

				// Check if we've seen this sequence before
				if prevMinimizer, exists := seqToMinimizer[seqStr]; exists {
					if prevMinimizer != sk.Minimizer {
						t.Errorf("BIJECTION VIOLATION: sequence %s has two different minimizers:\n"+
							"  First: %d\n"+
							"  Second: %d\n"+
							"  This violates the super k-mer definition!",
							seqStr, prevMinimizer, sk.Minimizer)
					}
				} else {
					seqToMinimizer[seqStr] = sk.Minimizer
				}

				// Verify all k-mers in this super k-mer have the same minimizer
				if len(sk.Sequence) >= tc.k {
					for i := 0; i <= len(sk.Sequence)-tc.k; i++ {
						kmerSeq := sk.Sequence[i : i+tc.k]
						minimizer := findMinimizer(kmerSeq, tc.k, tc.m)
						if minimizer != sk.Minimizer {
							t.Errorf("K-mer at position %d in super k-mer has different minimizer:\n"+
								"  K-mer: %s\n"+
								"  Expected minimizer: %d\n"+
								"  Actual minimizer: %d\n"+
								"  Super k-mer: %s",
								i, string(kmerSeq), sk.Minimizer, minimizer, seqStr)
						}
					}
				}
			}
		})
	}
}

// findMinimizer computes the minimizer of a k-mer for testing purposes
func findMinimizer(kmer []byte, k int, m int) uint64 {
	if len(kmer) != k {
		return 0
	}

	mMask := uint64(1)<<(m*2) - 1
	rcShift := uint((m - 1) * 2)

	minMinimizer := uint64(^uint64(0)) // max uint64

	// Scan all m-mers in the k-mer
	var fwdMmer, rvcMmer uint64
	for i := 0; i < m-1 && i < len(kmer); i++ {
		code := uint64(__single_base_code__[kmer[i]&31])
		fwdMmer = (fwdMmer << 2) | code
		rvcMmer = (rvcMmer >> 2) | ((code ^ 3) << rcShift)
	}

	for i := m - 1; i < len(kmer); i++ {
		code := uint64(__single_base_code__[kmer[i]&31])
		fwdMmer = ((fwdMmer << 2) | code) & mMask
		rvcMmer = (rvcMmer >> 2) | ((code ^ 3) << rcShift)

		canonical := fwdMmer
		if rvcMmer < fwdMmer {
			canonical = rvcMmer
		}

		if canonical < minMinimizer {
			minMinimizer = canonical
		}
	}

	return minMinimizer
}

// Note: Tests for ToBioSequence and SuperKmerWorker are in a separate
// integration test package to avoid circular dependencies between
// obikmer and obiseq packages.
