package obikmer

// SuperKmer represents a maximal subsequence where all consecutive k-mers
// share the same minimizer.
type SuperKmer struct {
	Minimizer uint64 // The canonical minimizer value (normalized m-mer)
	Start     int    // Starting position in the original sequence (0-indexed)
	End       int    // Ending position (exclusive, like Go slice notation)
	Sequence  []byte // The actual DNA subsequence [Start:End]
}

// dequeItem represents an element in the monotone deque used for
// tracking minimizers in a sliding window.
type dequeItem struct {
	position  int    // Position of the m-mer in the sequence
	canonical uint64 // Canonical (normalized) m-mer value
}

// ExtractSuperKmers extracts super k-mers from a DNA sequence.
// A super k-mer is a maximal subsequence where all consecutive k-mers
// share the same minimizer. The minimizer of a k-mer is the smallest
// canonical m-mer among its (k-m+1) constituent m-mers.
//
// This function uses IterSuperKmers internally and collects results into a slice.
//
// Parameters:
//   - seq: DNA sequence as a byte slice (case insensitive, supports A, C, G, T, U)
//   - k: k-mer size (must be between m+1 and 31)
//   - m: minimizer size (must be between 1 and k-1)
//   - buffer: optional pre-allocated buffer for results. If nil, a new slice is created.
//
// Returns:
//   - slice of SuperKmer structs representing maximal subsequences
//   - nil if parameters are invalid or sequence is too short
//
// Time complexity: O(n) where n is the sequence length
// Space complexity: O(k-m+1) for the deque + O(number of super k-mers) for results
func ExtractSuperKmers(seq []byte, k int, m int, buffer *[]SuperKmer) []SuperKmer {
	if m < 1 || m >= k || k < 2 || k > 31 || len(seq) < k {
		return nil
	}

	var result []SuperKmer
	if buffer == nil {
		estimatedSize := len(seq) / k
		if estimatedSize < 1 {
			estimatedSize = 1
		}
		result = make([]SuperKmer, 0, estimatedSize)
	} else {
		result = (*buffer)[:0]
	}

	for sk := range IterSuperKmers(seq, k, m) {
		result = append(result, sk)
	}

	return result
}
