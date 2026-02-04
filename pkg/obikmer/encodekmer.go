package obikmer

import "iter"

// Error markers for k-mers of odd length ≤ 31
// For odd k ≤ 31, only k*2 bits are used (max 62 bits), leaving 2 high bits
// available for error coding in the top 2 bits (bits 62-63).
//
// Error codes are simple integers:
//   0 = no error
//   1 = error type 1
//   2 = error type 2
//   3 = error type 3
//
// Use SetKmerError(kmer, code) and GetKmerError(kmer) to manipulate error bits.
const (
	KmerErrorMask    uint64 = 0b11 << 62  // Mask to extract error bits (bits 62-63)
	KmerSequenceMask uint64 = ^KmerErrorMask // Mask to extract sequence bits (bits 0-61)
)

// SetKmerError sets the error marker bits on a k-mer encoded value.
// Only valid for odd k-mer sizes ≤ 31 where 2 bits remain unused.
//
// Parameters:
//   - kmer: the encoded k-mer value
//   - errorCode: error code (0-3), where 0=no error, 1-3=error types
//
// Returns:
//   - k-mer with error bits set
func SetKmerError(kmer uint64, errorCode uint64) uint64 {
	return (kmer & KmerSequenceMask) | ((errorCode & 0b11) << 62)
}

// GetKmerError extracts the error marker bits from a k-mer encoded value.
//
// Returns:
//   - error code (0-3) as raw value (not shifted)
func GetKmerError(kmer uint64) uint64 {
	return (kmer & KmerErrorMask) >> 62
}

// ClearKmerError removes the error marker bits from a k-mer, returning
// just the sequence encoding.
//
// Returns:
//   - k-mer with error bits cleared (set to 00)
func ClearKmerError(kmer uint64) uint64 {
	return kmer & KmerSequenceMask
}

// EncodeKmers converts a DNA sequence to a slice of encoded k-mers.
// Each nucleotide is encoded on 2 bits according to __single_base_code__:
//   - A = 0 (00)
//   - C = 1 (01)
//   - G = 2 (10)
//   - T/U = 3 (11)
//
// The function returns overlapping k-mers of size k encoded as uint64.
// For a sequence of length n, it returns n-k+1 k-mers.
//
// The maximum k-mer size is 31 (using 62 bits), leaving the top 2 bits
// available for error markers (see SetKmerError).
//
// Parameters:
//   - seq: DNA sequence as a byte slice (case insensitive, supports A, C, G, T, U)
//   - k: k-mer size (must be between 1 and 31)
//   - buffer: optional pre-allocated buffer for results. If nil, a new slice is created.
//
// Returns:
//   - slice of uint64 encoded k-mers
//   - nil if sequence is shorter than k or k is invalid
func EncodeKmers(seq []byte, k int, buffer *[]uint64) []uint64 {
	if k < 1 || k > 31 || len(seq) < k {
		return nil
	}

	n := len(seq) - k + 1

	var result []uint64
	if buffer == nil {
		result = make([]uint64, 0, n)
	} else {
		result = (*buffer)[:0]
	}

	// Mask to keep only k*2 bits
	mask := uint64(1)<<(k*2) - 1

	// Build the first k-mer
	var kmer uint64
	for i := 0; i < k; i++ {
		kmer <<= 2
		kmer |= uint64(__single_base_code__[seq[i]&31])
	}
	result = append(result, kmer)

	// Slide through the rest of the sequence
	for i := k; i < len(seq); i++ {
		kmer <<= 2
		kmer |= uint64(__single_base_code__[seq[i]&31])
		kmer &= mask
		result = append(result, kmer)
	}

	return result
}

// IterKmers returns an iterator over all k-mers in the sequence.
// No intermediate slice is allocated, making it memory-efficient for
// processing k-mers one by one (e.g., adding to a Roaring Bitmap).
//
// Parameters:
//   - seq: DNA sequence as a byte slice (case insensitive, supports A, C, G, T, U)
//   - k: k-mer size (must be between 1 and 31)
//
// Returns:
//   - iterator yielding uint64 encoded k-mers
//
// Example:
//   for kmer := range IterKmers(seq, 21) {
//       bitmap.Add(kmer)
//   }
func IterKmers(seq []byte, k int) iter.Seq[uint64] {
	return func(yield func(uint64) bool) {
		if k < 1 || k > 31 || len(seq) < k {
			return
		}

		// Mask to keep only k*2 bits
		mask := uint64(1)<<(k*2) - 1

		// Build the first k-mer
		var kmer uint64
		for i := 0; i < k; i++ {
			kmer <<= 2
			kmer |= uint64(__single_base_code__[seq[i]&31])
		}

		if !yield(kmer) {
			return
		}

		// Slide through the rest of the sequence
		for i := k; i < len(seq); i++ {
			kmer <<= 2
			kmer |= uint64(__single_base_code__[seq[i]&31])
			kmer &= mask

			if !yield(kmer) {
				return
			}
		}
	}
}

// IterNormalizedKmers returns an iterator over all normalized (canonical) k-mers.
// No intermediate slice is allocated, making it memory-efficient.
//
// Parameters:
//   - seq: DNA sequence as a byte slice (case insensitive, supports A, C, G, T, U)
//   - k: k-mer size (must be between 1 and 31)
//
// Returns:
//   - iterator yielding uint64 normalized k-mers
//
// Example:
//   for canonical := range IterNormalizedKmers(seq, 21) {
//       bitmap.Add(canonical)
//   }
func IterNormalizedKmers(seq []byte, k int) iter.Seq[uint64] {
	return func(yield func(uint64) bool) {
		if k < 1 || k > 31 || len(seq) < k {
			return
		}

		// Mask to keep only k*2 bits
		mask := uint64(1)<<(k*2) - 1

		// Shift amount for adding to reverse complement (high position)
		rcShift := uint((k - 1) * 2)

		// Build the first k-mer (forward and reverse complement)
		var fwd, rvc uint64
		for i := 0; i < k; i++ {
			code := uint64(__single_base_code__[seq[i]&31])
			// Forward: shift left and add new code at low end
			fwd <<= 2
			fwd |= code
			// Reverse complement: shift right and add complement at high end
			rvc >>= 2
			rvc |= (code ^ 3) << rcShift
		}

		// Yield normalized k-mer
		var canonical uint64
		if fwd <= rvc {
			canonical = fwd
		} else {
			canonical = rvc
		}

		if !yield(canonical) {
			return
		}

		// Slide through the rest of the sequence
		for i := k; i < len(seq); i++ {
			code := uint64(__single_base_code__[seq[i]&31])

			// Update forward k-mer: shift left, add new code, mask
			fwd <<= 2
			fwd |= code
			fwd &= mask

			// Update reverse complement: shift right, add complement at high end
			rvc >>= 2
			rvc |= (code ^ 3) << rcShift

			// Yield normalized k-mer
			if fwd <= rvc {
				canonical = fwd
			} else {
				canonical = rvc
			}

			if !yield(canonical) {
				return
			}
		}
	}
}

// SuperKmer represents a maximal subsequence where all consecutive k-mers
// share the same minimizer. A minimizer is the smallest canonical m-mer
// among the (k-m+1) m-mers contained in a k-mer.
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
// The algorithm uses:
// - Simultaneous forward/reverse m-mer encoding for O(1) canonical m-mer computation
// - Monotone deque for O(1) amortized minimizer tracking per position
//
// The maximum k-mer size is 31 (using 62 bits), leaving the top 2 bits
// available for error markers if needed.
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
	// Validate parameters
	if m < 1 || m >= k || k < 2 || k > 31 || len(seq) < k {
		return nil
	}

	// Initialize result buffer
	var result []SuperKmer
	if buffer == nil {
		// Estimate: worst case is one super k-mer per k nucleotides
		estimatedSize := len(seq) / k
		if estimatedSize < 1 {
			estimatedSize = 1
		}
		result = make([]SuperKmer, 0, estimatedSize)
	} else {
		result = (*buffer)[:0]
	}

	// Initialize monotone deque for tracking minimizers
	deque := make([]dequeItem, 0, k-m+1)

	// Masks for m-mer encoding
	mMask := uint64(1)<<(m*2) - 1
	rcShift := uint((m - 1) * 2)

	// Build first m-1 nucleotides (can't form complete m-mer yet)
	var fwdMmer, rvcMmer uint64
	for i := 0; i < m-1 && i < len(seq); i++ {
		code := uint64(__single_base_code__[seq[i]&31])
		fwdMmer = (fwdMmer << 2) | code
		rvcMmer = (rvcMmer >> 2) | ((code ^ 3) << rcShift)
	}

	// Track super k-mer boundaries
	superKmerStart := 0
	var currentMinimizer uint64
	firstKmer := true

	// Slide through sequence, processing each position that completes an m-mer
	for pos := m - 1; pos < len(seq); pos++ {
		// Add new nucleotide to m-mer
		code := uint64(__single_base_code__[seq[pos]&31])
		fwdMmer = ((fwdMmer << 2) | code) & mMask
		rvcMmer = (rvcMmer >> 2) | ((code ^ 3) << rcShift)

		// Get canonical m-mer (minimum of forward and reverse complement)
		canonical := fwdMmer
		if rvcMmer < fwdMmer {
			canonical = rvcMmer
		}

		mmerPos := pos - m + 1

		// Remove m-mers outside the current k-mer window from front of deque
		// The k-mer at position pos spans from (pos-k+1) to pos
		// It contains m-mers from position (pos-k+1) to (pos-m+1)
		if pos >= k-1 {
			windowStart := pos - k + 1
			for len(deque) > 0 && deque[0].position < windowStart {
				deque = deque[1:]
			}
		}

		// Maintain monotone property: remove larger values from back
		for len(deque) > 0 && deque[len(deque)-1].canonical >= canonical {
			deque = deque[:len(deque)-1]
		}

		// Add new m-mer to deque
		deque = append(deque, dequeItem{position: mmerPos, canonical: canonical})

		// Once we have processed the first k nucleotides, we have our first k-mer
		if pos >= k-1 {
			// The minimizer is at the front of the deque
			newMinimizer := deque[0].canonical
			kmerStart := pos - k + 1 // Start position of current k-mer (ending at pos)

			if firstKmer {
				// Initialize first super k-mer
				currentMinimizer = newMinimizer
				firstKmer = false
			} else if newMinimizer != currentMinimizer {
				// Minimizer changed at this k-mer position
				// Previous k-mer started at position kmerStart-1
				// That k-mer is seq[kmerStart-1 : kmerStart-1+k] (Go slice notation)
				// The last base of that k-mer is at kmerStart-1+k-1 = kmerStart+k-2
				// In Go slice notation (exclusive end): kmerStart+k-1
				endPos := kmerStart + k - 1
				superKmer := SuperKmer{
					Minimizer: currentMinimizer,
					Start:     superKmerStart,
					End:       endPos,
					Sequence:  seq[superKmerStart:endPos],
				}
				result = append(result, superKmer)

				// New super k-mer starts at current k-mer position
				superKmerStart = kmerStart
				currentMinimizer = newMinimizer
			}
		}
	}

	// Emit final super k-mer
	if !firstKmer {
		superKmer := SuperKmer{
			Minimizer: currentMinimizer,
			Start:     superKmerStart,
			End:       len(seq),
			Sequence:  seq[superKmerStart:],
		}
		result = append(result, superKmer)
	}

	return result
}

// ReverseComplement computes the reverse complement of an encoded k-mer.
// The k-mer is encoded with 2 bits per nucleotide (A=00, C=01, G=10, T=11).
// The complement is: A↔T (00↔11), C↔G (01↔10), which is simply XOR with 11.
// The reverse swaps the order of 2-bit pairs.
//
// For k-mers with error markers (top 2 bits), the error bits are preserved
// and transferred to the reverse complement.
//
// Parameters:
//   - kmer: the encoded k-mer (possibly with error bits in positions 62-63)
//   - k: the k-mer size (number of nucleotides)
//
// Returns:
//   - the reverse complement of the k-mer with error bits preserved
func ReverseComplement(kmer uint64, k int) uint64 {
	// Step 0: Extract and preserve error bits
	errorBits := kmer & KmerErrorMask

	// Step 1: Complement - XOR with all 1s to flip A↔T and C↔G
	// For a k-mer of size k, we only want to flip the lower k*2 bits
	mask := uint64(1)<<(k*2) - 1
	rc := (^kmer) & mask

	// Step 2: Reverse the order of 2-bit pairs
	// We use a series of swaps at increasing granularity
	rc = ((rc & 0x3333333333333333) << 2) | ((rc & 0xCCCCCCCCCCCCCCCC) >> 2)   // Swap adjacent pairs
	rc = ((rc & 0x0F0F0F0F0F0F0F0F) << 4) | ((rc & 0xF0F0F0F0F0F0F0F0) >> 4)   // Swap nibbles
	rc = ((rc & 0x00FF00FF00FF00FF) << 8) | ((rc & 0xFF00FF00FF00FF00) >> 8)   // Swap bytes
	rc = ((rc & 0x0000FFFF0000FFFF) << 16) | ((rc & 0xFFFF0000FFFF0000) >> 16) // Swap 16-bit words
	rc = (rc << 32) | (rc >> 32)                                               // Swap 32-bit words

	// Step 3: Shift right to align the k-mer (we reversed all 32 pairs, need only k)
	rc >>= (64 - k*2)

	// Step 4: Restore error bits
	rc |= errorBits

	return rc
}

// NormalizeKmer returns the lexicographically smaller of a k-mer and its
// reverse complement. This canonical form ensures that a k-mer and its
// reverse complement map to the same value.
//
// Parameters:
//   - kmer: the encoded k-mer
//   - k: the k-mer size (number of nucleotides)
//
// Returns:
//   - the canonical (normalized) k-mer
func NormalizeKmer(kmer uint64, k int) uint64 {
	rc := ReverseComplement(kmer, k)
	if rc < kmer {
		return rc
	}
	return kmer
}

// EncodeNormalizedKmers converts a DNA sequence to a slice of normalized k-mers.
// Each k-mer is replaced by the lexicographically smaller of itself and its
// reverse complement. This ensures that forward and reverse complement sequences
// produce the same k-mer set.
//
// The maximum k-mer size is 31 (using 62 bits), leaving the top 2 bits
// available for error markers (see SetKmerError).
//
// Parameters:
//   - seq: DNA sequence as a byte slice (case insensitive, supports A, C, G, T, U)
//   - k: k-mer size (must be between 1 and 31)
//   - buffer: optional pre-allocated buffer for results. If nil, a new slice is created.
//
// Returns:
//   - slice of uint64 normalized k-mers
//   - nil if sequence is shorter than k or k is invalid
func EncodeNormalizedKmers(seq []byte, k int, buffer *[]uint64) []uint64 {
	if k < 1 || k > 31 || len(seq) < k {
		return nil
	}

	n := len(seq) - k + 1

	var result []uint64
	if buffer == nil {
		result = make([]uint64, 0, n)
	} else {
		result = (*buffer)[:0]
	}

	// Mask to keep only k*2 bits
	mask := uint64(1)<<(k*2) - 1

	// Shift amount for adding to reverse complement (high position)
	rcShift := uint((k - 1) * 2)

	// Complement lookup: A(00)->T(11), C(01)->G(10), G(10)->C(01), T(11)->A(00)
	// This is simply XOR with 3

	// Build the first k-mer (forward and reverse complement)
	var fwd, rvc uint64
	for i := 0; i < k; i++ {
		code := uint64(__single_base_code__[seq[i]&31])
		// Forward: shift left and add new code at low end
		fwd <<= 2
		fwd |= code
		// Reverse complement: shift right and add complement at high end
		rvc >>= 2
		rvc |= (code ^ 3) << rcShift
	}

	// Store the normalized (canonical) k-mer
	if fwd <= rvc {
		result = append(result, fwd)
	} else {
		result = append(result, rvc)
	}

	// Slide through the rest of the sequence
	for i := k; i < len(seq); i++ {
		code := uint64(__single_base_code__[seq[i]&31])

		// Update forward k-mer: shift left, add new code, mask
		fwd <<= 2
		fwd |= code
		fwd &= mask

		// Update reverse complement: shift right, add complement at high end
		rvc >>= 2
		rvc |= (code ^ 3) << rcShift

		// Store the normalized k-mer
		if fwd <= rvc {
			result = append(result, fwd)
		} else {
			result = append(result, rvc)
		}
	}

	return result
}
