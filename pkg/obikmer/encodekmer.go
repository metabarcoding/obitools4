package obikmer

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
// Parameters:
//   - seq: DNA sequence as a byte slice (case insensitive, supports A, C, G, T, U)
//   - k: k-mer size (must be between 1 and 32)
//   - buffer: optional pre-allocated buffer for results. If nil, a new slice is created.
//
// Returns:
//   - slice of uint64 encoded k-mers
//   - nil if sequence is shorter than k or k is invalid
func EncodeKmers(seq []byte, k int, buffer *[]uint64) []uint64 {
	if k < 1 || k > 32 || len(seq) < k {
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

// ReverseComplement computes the reverse complement of an encoded k-mer.
// The k-mer is encoded with 2 bits per nucleotide (A=00, C=01, G=10, T=11).
// The complement is: A↔T (00↔11), C↔G (01↔10), which is simply XOR with 11.
// The reverse swaps the order of 2-bit pairs.
//
// Parameters:
//   - kmer: the encoded k-mer
//   - k: the k-mer size (number of nucleotides)
//
// Returns:
//   - the reverse complement of the k-mer
func ReverseComplement(kmer uint64, k int) uint64 {
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
// Parameters:
//   - seq: DNA sequence as a byte slice (case insensitive, supports A, C, G, T, U)
//   - k: k-mer size (must be between 1 and 32)
//   - buffer: optional pre-allocated buffer for results. If nil, a new slice is created.
//
// Returns:
//   - slice of uint64 normalized k-mers
//   - nil if sequence is shorter than k or k is invalid
func EncodeNormalizedKmers(seq []byte, k int, buffer *[]uint64) []uint64 {
	if k < 1 || k > 32 || len(seq) < k {
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
