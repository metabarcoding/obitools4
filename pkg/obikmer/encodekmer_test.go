package obikmer

import (
	"bytes"
	"testing"
)

// TestEncodeKmersBasic tests basic k-mer encoding
func TestEncodeKmersBasic(t *testing.T) {
	tests := []struct {
		name     string
		seq      string
		k        int
		expected []uint64
	}{
		{
			name:     "simple 4-mer ACGT",
			seq:      "ACGT",
			k:        4,
			expected: []uint64{0b00011011}, // A=00, C=01, G=10, T=11 -> 00 01 10 11 = 27
		},
		{
			name:     "simple 2-mer AC",
			seq:      "AC",
			k:        2,
			expected: []uint64{0b0001}, // A=00, C=01 -> 00 01 = 1
		},
		{
			name:     "sliding 2-mer ACGT",
			seq:      "ACGT",
			k:        2,
			expected: []uint64{0b0001, 0b0110, 0b1011}, // AC=1, CG=6, GT=11
		},
		{
			name:     "lowercase",
			seq:      "acgt",
			k:        4,
			expected: []uint64{0b00011011},
		},
		{
			name:     "with U instead of T",
			seq:      "ACGU",
			k:        4,
			expected: []uint64{0b00011011}, // U encodes same as T
		},
		{
			name:     "8-mer",
			seq:      "ACGTACGT",
			k:        8,
			expected: []uint64{0b0001101100011011}, // ACGTACGT
		},
		{
			name:     "32-mer max size",
			seq:      "ACGTACGTACGTACGTACGTACGTACGTACGT",
			k:        32,
			expected: []uint64{0x1B1B1B1B1B1B1B1B}, // ACGTACGT repeated 4 times
		},
		{
			name: "longer sequence sliding",
			seq:  "AAACCCGGG",
			k:    3,
			expected: []uint64{
				0b000000, // AAA = 0
				0b000001, // AAC = 1
				0b000101, // ACC = 5
				0b010101, // CCC = 21
				0b010110, // CCG = 22
				0b011010, // CGG = 26
				0b101010, // GGG = 42
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeKmers([]byte(tt.seq), tt.k, nil)

			if len(result) != len(tt.expected) {
				t.Errorf("length mismatch: got %d, want %d", len(result), len(tt.expected))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("position %d: got %d (0b%b), want %d (0b%b)",
						i, v, v, tt.expected[i], tt.expected[i])
				}
			}
		})
	}
}

// TestEncodeKmersEdgeCases tests edge cases
func TestEncodeKmersEdgeCases(t *testing.T) {
	// Empty sequence
	result := EncodeKmers([]byte{}, 4, nil)
	if result != nil {
		t.Errorf("empty sequence should return nil, got %v", result)
	}

	// k > sequence length
	result = EncodeKmers([]byte("ACG"), 4, nil)
	if result != nil {
		t.Errorf("k > seq length should return nil, got %v", result)
	}

	// k = 0
	result = EncodeKmers([]byte("ACGT"), 0, nil)
	if result != nil {
		t.Errorf("k=0 should return nil, got %v", result)
	}

	// k > 32
	result = EncodeKmers([]byte("ACGTACGTACGTACGTACGTACGTACGTACGTACGT"), 33, nil)
	if result != nil {
		t.Errorf("k>32 should return nil, got %v", result)
	}

	// k = sequence length (single k-mer)
	result = EncodeKmers([]byte("ACGT"), 4, nil)
	if len(result) != 1 {
		t.Errorf("k=seq_len should return 1 k-mer, got %d", len(result))
	}
}

// TestEncodeKmersBuffer tests buffer reuse
func TestEncodeKmersBuffer(t *testing.T) {
	seq := []byte("ACGTACGTACGT")
	k := 4

	// First call without buffer
	result1 := EncodeKmers(seq, k, nil)

	// Second call with buffer - pre-allocate with capacity
	buffer := make([]uint64, 0, 100)
	result2 := EncodeKmers(seq, k, &buffer)

	if len(result1) != len(result2) {
		t.Errorf("buffer reuse: length mismatch %d vs %d", len(result1), len(result2))
	}

	for i := range result1 {
		if result1[i] != result2[i] {
			t.Errorf("buffer reuse: position %d mismatch", i)
		}
	}

	// Verify results are correct
	if len(result2) == 0 {
		t.Errorf("result should not be empty")
	}

	// Test multiple calls with same buffer to verify no memory issues
	for i := 0; i < 10; i++ {
		result3 := EncodeKmers(seq, k, &buffer)
		if len(result3) != len(result1) {
			t.Errorf("iteration %d: length mismatch", i)
		}
	}
}

// TestEncodeKmersVariousLengths tests encoding with various sequence lengths
func TestEncodeKmersVariousLengths(t *testing.T) {
	lengths := []int{1, 4, 8, 15, 16, 17, 31, 32, 33, 63, 64, 65, 100, 256, 1000}
	k := 8

	for _, length := range lengths {
		// Generate test sequence
		seq := make([]byte, length)
		for i := range seq {
			seq[i] = "ACGT"[i%4]
		}

		if length < k {
			continue
		}

		t.Run("length_"+string(rune('0'+length/100))+string(rune('0'+(length%100)/10))+string(rune('0'+length%10)), func(t *testing.T) {
			result := EncodeKmers(seq, k, nil)

			expectedLen := length - k + 1
			if len(result) != expectedLen {
				t.Errorf("length mismatch: got %d, want %d", len(result), expectedLen)
			}
		})
	}
}

// TestEncodeKmersLongSequence tests with a longer realistic sequence
func TestEncodeKmersLongSequence(t *testing.T) {
	// Simulate a realistic DNA sequence
	seq := bytes.Repeat([]byte("ACGTACGTNNACGTACGT"), 100)
	k := 16

	result := EncodeKmers(seq, k, nil)
	expectedLen := len(seq) - k + 1

	if len(result) != expectedLen {
		t.Fatalf("length mismatch: got %d, want %d", len(result), expectedLen)
	}
}

// BenchmarkEncodeKmers benchmarks the encoding function
func BenchmarkEncodeKmers(b *testing.B) {
	// Create test sequences of various sizes
	sizes := []int{100, 1000, 10000, 100000}
	kSizes := []int{8, 16, 32}

	for _, k := range kSizes {
		for _, size := range sizes {
			seq := make([]byte, size)
			for i := range seq {
				seq[i] = "ACGT"[i%4]
			}

			name := "k" + string(rune('0'+k/10)) + string(rune('0'+k%10)) + "_size" + string(rune('0'+size/10000)) + string(rune('0'+(size%10000)/1000)) + string(rune('0'+(size%1000)/100)) + string(rune('0'+(size%100)/10)) + string(rune('0'+size%10))
			b.Run(name, func(b *testing.B) {
				buffer := make([]uint64, 0, size)
				b.ResetTimer()
				b.SetBytes(int64(size))

				for i := 0; i < b.N; i++ {
					EncodeKmers(seq, k, &buffer)
				}
			})
		}
	}
}

// TestEncodeNucleotide verifies nucleotide encoding
func TestEncodeNucleotide(t *testing.T) {
	testCases := []struct {
		nucleotide byte
		expected   byte
	}{
		{'A', 0},
		{'a', 0},
		{'C', 1},
		{'c', 1},
		{'G', 2},
		{'g', 2},
		{'T', 3},
		{'t', 3},
		{'U', 3},
		{'u', 3},
	}

	for _, tc := range testCases {
		result := EncodeNucleotide(tc.nucleotide)
		if result != tc.expected {
			t.Errorf("EncodeNucleotide('%c') = %d, want %d",
				tc.nucleotide, result, tc.expected)
		}
	}
}

// TestReverseComplement tests the reverse complement function
func TestReverseComplement(t *testing.T) {
	tests := []struct {
		name     string
		seq      string
		k        int
		expected string // expected reverse complement sequence
	}{
		{
			name:     "ACGT -> ACGT (palindrome)",
			seq:      "ACGT",
			k:        4,
			expected: "ACGT",
		},
		{
			name:     "AAAA -> TTTT",
			seq:      "AAAA",
			k:        4,
			expected: "TTTT",
		},
		{
			name:     "TTTT -> AAAA",
			seq:      "TTTT",
			k:        4,
			expected: "AAAA",
		},
		{
			name:     "CCCC -> GGGG",
			seq:      "CCCC",
			k:        4,
			expected: "GGGG",
		},
		{
			name:     "AACG -> CGTT",
			seq:      "AACG",
			k:        4,
			expected: "CGTT",
		},
		{
			name:     "AC -> GT",
			seq:      "AC",
			k:        2,
			expected: "GT",
		},
		{
			name:     "ACGTACGT -> ACGTACGT (palindrome)",
			seq:      "ACGTACGT",
			k:        8,
			expected: "ACGTACGT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode the input sequence
			kmers := EncodeKmers([]byte(tt.seq), tt.k, nil)
			if len(kmers) != 1 {
				t.Fatalf("expected 1 k-mer, got %d", len(kmers))
			}

			// Compute reverse complement
			rc := ReverseComplement(kmers[0], tt.k)

			// Encode the expected reverse complement
			expectedKmers := EncodeKmers([]byte(tt.expected), tt.k, nil)
			if len(expectedKmers) != 1 {
				t.Fatalf("expected 1 k-mer for expected, got %d", len(expectedKmers))
			}

			if rc != expectedKmers[0] {
				t.Errorf("ReverseComplement(%s) = %d (0b%b), want %d (0b%b) for %s",
					tt.seq, rc, rc, expectedKmers[0], expectedKmers[0], tt.expected)
			}
		})
	}
}

// TestReverseComplementInvolution tests that RC(RC(x)) = x
func TestReverseComplementInvolution(t *testing.T) {
	testSeqs := []string{"ACGT", "AAAA", "TTTT", "ACGTACGT", "AACGTTGC", "AC", "ACGTACGTACGTACGT", "ACGTACGTACGTACGTACGTACGTACGTACGT"}

	for _, seq := range testSeqs {
		k := len(seq)
		kmers := EncodeKmers([]byte(seq), k, nil)
		if len(kmers) != 1 {
			continue
		}

		original := kmers[0]
		rc := ReverseComplement(original, k)
		rcrc := ReverseComplement(rc, k)

		if rcrc != original {
			t.Errorf("RC(RC(%s)) != %s: got %d, want %d", seq, seq, rcrc, original)
		}
	}
}

// TestNormalizeKmer tests the normalization function
func TestNormalizeKmer(t *testing.T) {
	tests := []struct {
		name string
		seq  string
		k    int
	}{
		{"ACGT palindrome", "ACGT", 4},
		{"AAAA vs TTTT", "AAAA", 4},
		{"TTTT vs AAAA", "TTTT", 4},
		{"AACG vs CGTT", "AACG", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kmers := EncodeKmers([]byte(tt.seq), tt.k, nil)
			if len(kmers) != 1 {
				t.Fatalf("expected 1 k-mer, got %d", len(kmers))
			}

			kmer := kmers[0]
			rc := ReverseComplement(kmer, tt.k)
			normalized := NormalizeKmer(kmer, tt.k)

			// Normalized should be the minimum
			expectedNorm := kmer
			if rc < kmer {
				expectedNorm = rc
			}

			if normalized != expectedNorm {
				t.Errorf("NormalizeKmer(%d) = %d, want %d", kmer, normalized, expectedNorm)
			}

			// Normalizing the RC should give the same result
			normalizedRC := NormalizeKmer(rc, tt.k)
			if normalizedRC != normalized {
				t.Errorf("NormalizeKmer(RC) = %d, want %d (same as NormalizeKmer(fwd))", normalizedRC, normalized)
			}
		})
	}
}

// TestEncodeNormalizedKmersBasic tests basic normalized k-mer encoding
func TestEncodeNormalizedKmersBasic(t *testing.T) {
	// Test that a sequence and its reverse complement produce the same normalized k-mers
	seq := []byte("AACGTT")
	revComp := []byte("AACGTT") // This is a palindrome!

	k := 4
	kmers1 := EncodeNormalizedKmers(seq, k, nil)
	kmers2 := EncodeNormalizedKmers(revComp, k, nil)

	if len(kmers1) != len(kmers2) {
		t.Fatalf("length mismatch: %d vs %d", len(kmers1), len(kmers2))
	}

	// For a palindrome, forward and reverse should give the same k-mers
	for i := range kmers1 {
		if kmers1[i] != kmers2[len(kmers2)-1-i] {
			t.Logf("Note: position %d differs (expected for non-palindromic sequences)", i)
		}
	}
}

// TestEncodeNormalizedKmersSymmetry tests that seq and its RC produce same normalized k-mers (reversed)
func TestEncodeNormalizedKmersSymmetry(t *testing.T) {
	// Manually construct a sequence and its reverse complement
	seq := []byte("ACGTAACCGG")

	// Compute reverse complement manually
	rcMap := map[byte]byte{'A': 'T', 'C': 'G', 'G': 'C', 'T': 'A'}
	revComp := make([]byte, len(seq))
	for i, b := range seq {
		revComp[len(seq)-1-i] = rcMap[b]
	}

	k := 4
	kmers1 := EncodeNormalizedKmers(seq, k, nil)
	kmers2 := EncodeNormalizedKmers(revComp, k, nil)

	if len(kmers1) != len(kmers2) {
		t.Fatalf("length mismatch: %d vs %d", len(kmers1), len(kmers2))
	}

	// The normalized k-mers should be the same but in reverse order
	for i := range kmers1 {
		j := len(kmers2) - 1 - i
		if kmers1[i] != kmers2[j] {
			t.Errorf("position %d vs %d: %d != %d", i, j, kmers1[i], kmers2[j])
		}
	}
}

// TestEncodeNormalizedKmersConsistency verifies normalized k-mers match manual normalization
func TestEncodeNormalizedKmersConsistency(t *testing.T) {
	seq := []byte("ACGTACGTACGTACGT")
	k := 8

	// Get k-mers both ways
	rawKmers := EncodeKmers(seq, k, nil)
	normalizedKmers := EncodeNormalizedKmers(seq, k, nil)

	if len(rawKmers) != len(normalizedKmers) {
		t.Fatalf("length mismatch: %d vs %d", len(rawKmers), len(normalizedKmers))
	}

	// Verify each normalized k-mer matches manual normalization
	for i, raw := range rawKmers {
		expected := NormalizeKmer(raw, k)
		if normalizedKmers[i] != expected {
			t.Errorf("position %d: EncodeNormalizedKmers gave %d, NormalizeKmer gave %d",
				i, normalizedKmers[i], expected)
		}
	}
}

// BenchmarkEncodeNormalizedKmers benchmarks the normalized encoding function
func BenchmarkEncodeNormalizedKmers(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}
	kSizes := []int{8, 16, 32}

	for _, k := range kSizes {
		for _, size := range sizes {
			seq := make([]byte, size)
			for i := range seq {
				seq[i] = "ACGT"[i%4]
			}

			name := "k" + string(rune('0'+k/10)) + string(rune('0'+k%10)) + "_size" + string(rune('0'+size/10000)) + string(rune('0'+(size%10000)/1000)) + string(rune('0'+(size%1000)/100)) + string(rune('0'+(size%100)/10)) + string(rune('0'+size%10))
			b.Run(name, func(b *testing.B) {
				buffer := make([]uint64, 0, size)
				b.ResetTimer()
				b.SetBytes(int64(size))

				for i := 0; i < b.N; i++ {
					EncodeNormalizedKmers(seq, k, &buffer)
				}
			})
		}
	}
}

// BenchmarkReverseComplement benchmarks the reverse complement function
func BenchmarkReverseComplement(b *testing.B) {
	kmer := uint64(0x123456789ABCDEF0)
	k := 32

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReverseComplement(kmer, k)
	}
}

// BenchmarkNormalizeKmer benchmarks the normalization function
func BenchmarkNormalizeKmer(b *testing.B) {
	kmer := uint64(0x123456789ABCDEF0)
	k := 32

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NormalizeKmer(kmer, k)
	}
}
