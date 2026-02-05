package obikmer

import "iter"

var __single_base_code_err__ = []byte{0,
	//     A,     B,     C,     D,
	0, 0xFF, 1, 0xFF,
	//     E,     F,     G,     H,
	0xFF, 0xFF, 2, 0xFF,
	//     I,     J,     K,     L,
	0xFF, 0xFF, 0xFF, 0xFF,
	//     M,     N,     O,     P,
	0xFF, 0xFF, 0xFF, 0xFF,
	//     Q,     R,     S,     T,
	0xFF, 0xFF, 0xFF, 3,
	//     U,     V,     W,     X,
	3, 0xFF, 0xFF, 0xFF,
	//     Y,     Z,     .,     .
	0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF,
}

const ambiguousBaseCode = byte(0xFF)

// Error markers for k-mers of odd length ≤ 31
// For odd k ≤ 31, only k*2 bits are used (max 62 bits), leaving 2 high bits
// available for error coding in the top 2 bits (bits 62-63).
//
// Error codes are simple integers:
//
//	0 = no error
//	1 = error type 1
//	2 = error type 2
//	3 = error type 3
//
// Use SetKmerError(kmer, code) and GetKmerError(kmer) to manipulate error bits.
const (
	KmerErrorMask    uint64 = 0b11 << 62     // Mask to extract error bits (bits 62-63)
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

// EncodeKmer encodes a single k-mer sequence to uint64.
// This is the optimal zero-allocation function for encoding a single k-mer.
//
// Each nucleotide is encoded on 2 bits according to __single_base_code__:
//   - A = 0 (00)
//   - C = 1 (01)
//   - G = 2 (10)
//   - T/U = 3 (11)
//
// The maximum k-mer size is 31 (using 62 bits), leaving the top 2 bits
// available for error markers if needed.
//
// Parameters:
//   - seq: DNA sequence as a byte slice (case insensitive, supports A, C, G, T, U)
//   - k: k-mer size (must be between 1 and 31)
//
// Returns:
//   - encoded k-mer as uint64
//   - panics if len(seq) != k or k is invalid
//
// Example:
//
//	kmer := EncodeKmer([]byte("ACGT"), 4)
func EncodeKmer(seq []byte, k int) uint64 {
	if k < 1 || k > 31 {
		panic("k must be between 1 and 31")
	}
	if len(seq) != k {
		panic("sequence length must equal k")
	}

	var kmer uint64
	for i := 0; i < k; i++ {
		kmer <<= 2
		kmer |= uint64(__single_base_code__[seq[i]&31])
	}
	return kmer
}

// EncodeCanonicalKmer encodes a single k-mer sequence to its canonical form (uint64).
// Returns the lexicographically smaller of the k-mer and its reverse complement.
// This is the optimal zero-allocation function for encoding a single canonical k-mer.
//
// Parameters:
//   - seq: DNA sequence as a byte slice (case insensitive, supports A, C, G, T, U)
//   - k: k-mer size (must be between 1 and 31)
//
// Returns:
//   - canonical k-mer as uint64
//   - panics if len(seq) != k or k is invalid
//
// Example:
//
//	canonical := EncodeCanonicalKmer([]byte("ACGT"), 4)
func EncodeCanonicalKmer(seq []byte, k int) uint64 {
	if k < 1 || k > 31 {
		panic("k must be between 1 and 31")
	}
	if len(seq) != k {
		panic("sequence length must equal k")
	}

	rcShift := uint((k - 1) * 2)

	var fwd, rvc uint64
	for i := 0; i < k; i++ {
		code := uint64(__single_base_code__[seq[i]&31])
		fwd <<= 2
		fwd |= code
		rvc >>= 2
		rvc |= (code ^ 3) << rcShift
	}

	if fwd <= rvc {
		return fwd
	}
	return rvc
}

// DecodeKmer decodes a uint64 k-mer back to a DNA sequence.
// This function reuses a provided buffer to avoid allocation.
//
// Parameters:
//   - kmer: encoded k-mer as uint64
//   - k: k-mer size (number of nucleotides)
//   - buffer: pre-allocated buffer of length >= k (if nil, allocates new slice)
//
// Returns:
//   - decoded DNA sequence as []byte (lowercase acgt)
//
// Example:
//
//	var buf [32]byte
//	seq := DecodeKmer(kmer, 21, buf[:])
func DecodeKmer(kmer uint64, k int, buffer []byte) []byte {
	var result []byte
	if buffer == nil || len(buffer) < k {
		result = make([]byte, k)
	} else {
		result = buffer[:k]
	}

	bases := [4]byte{'a', 'c', 'g', 't'}
	for i := k - 1; i >= 0; i-- {
		result[i] = bases[kmer&3]
		kmer >>= 2
	}
	return result
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

	var result []uint64
	if buffer == nil {
		result = make([]uint64, 0, len(seq)-k+1)
	} else {
		result = (*buffer)[:0]
	}

	for kmer := range IterKmers(seq, k) {
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
//
//	for kmer := range IterKmers(seq, 21) {
//	    bitmap.Add(kmer)
//	}
func IterKmers(seq []byte, k int) iter.Seq[uint64] {
	return func(yield func(uint64) bool) {
		if k < 1 || k > 31 || len(seq) < k {
			return
		}

		mask := uint64(1)<<(k*2) - 1

		var kmer uint64
		for i := 0; i < k; i++ {
			kmer <<= 2
			kmer |= uint64(__single_base_code__[seq[i]&31])
		}

		if !yield(kmer) {
			return
		}

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

// IterCanonicalKmersWithErrors returns an iterator over all canonical k-mers
// with error markers for ambiguous bases. No intermediate slice is allocated.
//
// Ambiguous bases (N, R, Y, W, S, K, M, B, D, H, V) are encoded as 0xFF and detected
// during k-mer construction. The error code in bits 62-63 indicates the number of
// ambiguous bases in each k-mer (0=clean, 1-3=error count).
//
// Only valid for odd k ≤ 31 where 2 bits remain unused for error markers.
//
// Parameters:
//   - seq: DNA sequence as a byte slice (case insensitive, supports A, C, G, T, U, and ambiguous bases)
//   - k: k-mer size (must be odd, between 1 and 31)
//
// Returns:
//   - iterator yielding uint64 canonical k-mers with error markers
//
// Example:
//
//	for kmer := range IterCanonicalKmersWithErrors(seq, 21) {
//	    if GetKmerError(kmer) == 0 {
//	        bitmap.Add(kmer) // Only add clean k-mers
//	    }
//	}
func IterCanonicalKmersWithErrors(seq []byte, k int) iter.Seq[uint64] {
	return func(yield func(uint64) bool) {
		if k < 1 || k > 31 || k%2 == 0 || len(seq) < k {
			return
		}

		mask := uint64(1)<<(k*2) - 1

		rcShift := uint((k - 1) * 2)

		ambiguousCount := 0
		const ambiguousCode = byte(0xFF)

		var fwd, rvc uint64
		hasError := false
		for i := 0; i < k; i++ {
			code := __single_base_code_err__[seq[i]&31]

			if code == ambiguousCode {
				ambiguousCount++
				hasError = true
				code = 0
			}

			codeUint := uint64(code)
			fwd <<= 2
			fwd |= codeUint
			rvc >>= 2
			rvc |= (codeUint ^ 3) << rcShift
		}

		var canonical uint64
		if fwd <= rvc {
			canonical = fwd
		} else {
			canonical = rvc
		}

		if hasError {
			errorCode := uint64(ambiguousCount)
			if errorCode > 3 {
				errorCode = 3
			}
			canonical = SetKmerError(canonical, errorCode)
		}

		if !yield(canonical) {
			return
		}

		for i := k; i < len(seq); i++ {
			outgoingCode := __single_base_code_err__[seq[i-k]&31]
			if outgoingCode == ambiguousCode {
				ambiguousCount--
			}

			code := __single_base_code_err__[seq[i]&31]
			if code == ambiguousCode {
				ambiguousCount++
				code = 0
			}

			codeUint := uint64(code)

			fwd <<= 2
			fwd |= codeUint
			fwd &= mask

			rvc >>= 2
			rvc |= (codeUint ^ 3) << rcShift

			if fwd <= rvc {
				canonical = fwd
			} else {
				canonical = rvc
			}

			if ambiguousCount > 0 {
				errorCode := uint64(ambiguousCount)
				if errorCode > 3 {
					errorCode = 3
				}
				canonical = SetKmerError(canonical, errorCode)
			}

			if !yield(canonical) {
				return
			}
		}
	}
}

// IterCanonicalKmers returns an iterator over all canonical k-mers.
// No intermediate slice is allocated, making it memory-efficient.
//
// Parameters:
//   - seq: DNA sequence as a byte slice (case insensitive, supports A, C, G, T, U)
//   - k: k-mer size (must be between 1 and 31)
//
// Returns:
//   - iterator yielding uint64 canonical k-mers
//
// Example:
//
//	for canonical := range IterCanonicalKmers(seq, 21) {
//	    bitmap.Add(canonical)
//	}
func IterCanonicalKmers(seq []byte, k int) iter.Seq[uint64] {
	return func(yield func(uint64) bool) {
		if k < 1 || k > 31 || len(seq) < k {
			return
		}

		mask := uint64(1)<<(k*2) - 1

		rcShift := uint((k - 1) * 2)

		var fwd, rvc uint64
		for i := 0; i < k; i++ {
			code := uint64(__single_base_code__[seq[i]&31])
			fwd <<= 2
			fwd |= code
			rvc >>= 2
			rvc |= (code ^ 3) << rcShift
		}

		var canonical uint64
		if fwd <= rvc {
			canonical = fwd
		} else {
			canonical = rvc
		}

		if !yield(canonical) {
			return
		}

		for i := k; i < len(seq); i++ {
			code := uint64(__single_base_code__[seq[i]&31])

			fwd <<= 2
			fwd |= code
			fwd &= mask

			rvc >>= 2
			rvc |= (code ^ 3) << rcShift

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

	deque := make([]dequeItem, 0, k-m+1)

	mMask := uint64(1)<<(m*2) - 1
	rcShift := uint((m - 1) * 2)

	var fwdMmer, rvcMmer uint64
	for i := 0; i < m-1 && i < len(seq); i++ {
		code := uint64(__single_base_code__[seq[i]&31])
		fwdMmer = (fwdMmer << 2) | code
		rvcMmer = (rvcMmer >> 2) | ((code ^ 3) << rcShift)
	}

	superKmerStart := 0
	var currentMinimizer uint64
	firstKmer := true

	for pos := m - 1; pos < len(seq); pos++ {
		code := uint64(__single_base_code__[seq[pos]&31])
		fwdMmer = ((fwdMmer << 2) | code) & mMask
		rvcMmer = (rvcMmer >> 2) | ((code ^ 3) << rcShift)

		canonical := fwdMmer
		if rvcMmer < fwdMmer {
			canonical = rvcMmer
		}

		mmerPos := pos - m + 1

		if pos >= k-1 {
			windowStart := pos - k + 1
			for len(deque) > 0 && deque[0].position < windowStart {
				deque = deque[1:]
			}
		}

		for len(deque) > 0 && deque[len(deque)-1].canonical >= canonical {
			deque = deque[:len(deque)-1]
		}

		deque = append(deque, dequeItem{position: mmerPos, canonical: canonical})

		if pos >= k-1 {
			newMinimizer := deque[0].canonical
			kmerStart := pos - k + 1

			if firstKmer {
				currentMinimizer = newMinimizer
				firstKmer = false
			} else if newMinimizer != currentMinimizer {
				endPos := kmerStart + k - 1
				superKmer := SuperKmer{
					Minimizer: currentMinimizer,
					Start:     superKmerStart,
					End:       endPos,
					Sequence:  seq[superKmerStart:endPos],
				}
				result = append(result, superKmer)

				superKmerStart = kmerStart
				currentMinimizer = newMinimizer
			}
		}
	}

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
	errorBits := kmer & KmerErrorMask

	mask := uint64(1)<<(k*2) - 1
	rc := (^kmer) & mask

	rc = ((rc & 0x3333333333333333) << 2) | ((rc & 0xCCCCCCCCCCCCCCCC) >> 2)
	rc = ((rc & 0x0F0F0F0F0F0F0F0F) << 4) | ((rc & 0xF0F0F0F0F0F0F0F0) >> 4)
	rc = ((rc & 0x00FF00FF00FF00FF) << 8) | ((rc & 0xFF00FF00FF00FF00) >> 8)
	rc = ((rc & 0x0000FFFF0000FFFF) << 16) | ((rc & 0xFFFF0000FFFF0000) >> 16)
	rc = (rc << 32) | (rc >> 32)

	rc >>= (64 - k*2)

	rc |= errorBits

	return rc
}

// CanonicalKmer returns the lexicographically smaller of a k-mer and its
// reverse complement. This canonical form ensures that a k-mer and its
// reverse complement map to the same value.
//
// This implements REVERSE COMPLEMENT canonicalization (biological canonical form).
//
// Parameters:
//   - kmer: the encoded k-mer
//   - k: the k-mer size (number of nucleotides)
//
// Returns:
//   - the canonical k-mer
func CanonicalKmer(kmer uint64, k int) uint64 {
	rc := ReverseComplement(kmer, k)
	if rc < kmer {
		return rc
	}
	return kmer
}

// NormalizeCircular returns the lexicographically smallest circular rotation
// of a k-mer. This is used for entropy calculations in low-complexity masking.
//
// This implements CIRCULAR PERMUTATION normalization (rotation-based canonicalization).
// Example: ACGT → min(ACGT, CGTA, GTAC, TACG) by circular rotation
//
// This is DIFFERENT from NormalizeKmer which uses reverse complement.
//
// Parameters:
//   - kmer: the encoded k-mer
//   - k: the k-mer size (number of nucleotides)
//
// Returns:
//   - the lexicographically smallest circular rotation
//
// Time complexity: O(k) - checks all k rotations
func NormalizeCircular(kmer uint64, k int) uint64 {
	if k < 1 || k > 31 {
		return kmer
	}

	mask := uint64(1)<<(k*2) - 1
	canonical := kmer
	current := kmer

	// Try all k rotations
	for i := 0; i < k; i++ {
		// Rotate: take top 2 bits, shift left, add to bottom
		top := (current >> ((k - 1) * 2)) & 3
		current = ((current << 2) | top) & mask

		if current < canonical {
			canonical = current
		}
	}

	return canonical
}

// EncodeCircularCanonicalKmer encodes a k-mer and returns its lexicographically
// smallest circular rotation. This is optimized for single k-mer encoding with
// circular canonicalization.
//
// This implements CIRCULAR PERMUTATION canonicalization, used for entropy-based
// low-complexity masking. This is DIFFERENT from EncodeCanonicalKmer which
// uses reverse complement canonicalization.
//
// Parameters:
//   - seq: DNA sequence as a byte slice (case insensitive, supports A, C, G, T, U)
//   - k: k-mer size (must be between 1 and 31)
//
// Returns:
//   - canonical k-mer as uint64 (smallest circular rotation)
//   - panics if len(seq) != k or k is invalid
//
// Example:
//
//	canonical := EncodeCircularCanonicalKmer([]byte("ACGT"), 4)
func EncodeCircularCanonicalKmer(seq []byte, k int) uint64 {
	kmer := EncodeKmer(seq, k)
	return NormalizeCircular(kmer, k)
}

// CanonicalCircularKmerCount returns the number of unique canonical k-mers
// under circular permutation normalization for DNA sequences (4-letter alphabet).
//
// This counts equivalence classes where k-mers are considered the same if one
// is a circular rotation of another (e.g., "ACGT", "CGTA", "GTAC", "TACG" are
// all equivalent).
//
// Uses Moreau's necklace-counting formula for exact counts:
//
//	N(n, a) = (1/n) * Σ φ(d) * a^(n/d)
//
// where the sum is over all divisors d of n, and φ is Euler's totient function.
//
// Parameters:
//   - k: k-mer size
//
// Returns:
//   - number of unique canonical k-mers under circular rotation
//
// Example:
//
//	count := CanonicalCircularKmerCount(4) // Returns 70 (not 256)
func CanonicalCircularKmerCount(k int) int {
	// Hardcoded exact counts for k=1 to 6 (optimization)
	switch k {
	case 1:
		return 4
	case 2:
		return 10
	case 3:
		return 24
	case 4:
		return 70
	case 5:
		return 208
	case 6:
		return 700
	default:
		// For k>6, use Moreau's necklace-counting formula
		return necklaceCount(k, 4)
	}
}

// eulerTotient computes Euler's totient function φ(n), which counts
// the number of integers from 1 to n that are coprime with n.
func eulerTotient(n int) int {
	if n <= 0 {
		return 0
	}

	result := n

	// Process all prime factors
	for p := 2; p*p <= n; p++ {
		if n%p == 0 {
			// Remove all occurrences of p
			for n%p == 0 {
				n /= p
			}
			// Apply: φ(n) = n * (1 - 1/p) = n * (p-1)/p
			result -= result / p
		}
	}

	// If n is still greater than 1, then it's a prime factor
	if n > 1 {
		result -= result / n
	}

	return result
}

// divisors returns all divisors of n in ascending order.
func divisors(n int) []int {
	if n <= 0 {
		return []int{}
	}

	divs := []int{}
	for i := 1; i*i <= n; i++ {
		if n%i == 0 {
			divs = append(divs, i)
			if i != n/i {
				divs = append(divs, n/i)
			}
		}
	}

	// Bubble sort in ascending order
	for i := 0; i < len(divs)-1; i++ {
		for j := i + 1; j < len(divs); j++ {
			if divs[i] > divs[j] {
				divs[i], divs[j] = divs[j], divs[i]
			}
		}
	}

	return divs
}

// necklaceCount computes the number of distinct necklaces (equivalence classes
// under rotation) for sequences of length n over an alphabet of size a.
// Uses Moreau's necklace-counting formula:
//
//	N(n, a) = (1/n) * Σ φ(d) * a^(n/d)
//
// where the sum is over all divisors d of n, and φ is Euler's totient function.
func necklaceCount(n, alphabetSize int) int {
	if n <= 0 {
		return 0
	}

	divs := divisors(n)
	sum := 0

	for _, d := range divs {
		// Compute a^(n/d)
		power := 1
		exp := n / d
		for i := 0; i < exp; i++ {
			power *= alphabetSize
		}

		sum += eulerTotient(d) * power
	}

	return sum / n
}

// EncodeCanonicalKmersWithErrors converts a DNA sequence to a slice of canonical k-mers
// with error markers for ambiguous bases (N, R, Y, W, S, K, M, B, D, H, V).
//
// Ambiguous bases are encoded as 0xFF by __single_base_code__ and detected during
// k-mer construction. The error code in bits 62-63 indicates the number of ambiguous
// bases in each k-mer:
//   - errorCode 0: no ambiguous bases (clean k-mer)
//   - errorCode 1: 1 ambiguous base
//   - errorCode 2: 2 ambiguous bases
//   - errorCode 3: 3 or more ambiguous bases
//
// Only valid for odd k ≤ 31 where 2 bits remain unused for error markers.
//
// Parameters:
//   - seq: DNA sequence as a byte slice (case insensitive, supports A, C, G, T, U, and ambiguous bases)
//   - k: k-mer size (must be odd, between 1 and 31)
//   - buffer: optional pre-allocated buffer for results. If nil, a new slice is created.
//
// Returns:
//   - slice of uint64 canonical k-mers with error markers
//   - nil if sequence is shorter than k, k is invalid, or k is even
func EncodeCanonicalKmersWithErrors(seq []byte, k int, buffer *[]uint64) []uint64 {
	if k < 1 || k > 31 || k%2 == 0 || len(seq) < k {
		return nil
	}

	var result []uint64
	if buffer == nil {
		result = make([]uint64, 0, len(seq)-k+1)
	} else {
		result = (*buffer)[:0]
	}

	for kmer := range IterCanonicalKmersWithErrors(seq, k) {
		result = append(result, kmer)
	}

	return result
}

// EncodeCanonicalKmers converts a DNA sequence to a slice of canonical k-mers.
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
//   - slice of uint64 canonical k-mers
//   - nil if sequence is shorter than k or k is invalid
func EncodeCanonicalKmers(seq []byte, k int, buffer *[]uint64) []uint64 {
	if k < 1 || k > 31 || len(seq) < k {
		return nil
	}

	var result []uint64
	if buffer == nil {
		result = make([]uint64, 0, len(seq)-k+1)
	} else {
		result = (*buffer)[:0]
	}

	for kmer := range IterCanonicalKmers(seq, k) {
		result = append(result, kmer)
	}

	return result
}
