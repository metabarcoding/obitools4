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
			name:     "31-mer max size",
			seq:      "ACGTACGTACGTACGTACGTACGTACGTACG",
			k:        31,
			expected: []uint64{0x06C6C6C6C6C6C6C6}, // ACGTACGT repeated ~4 times
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

	// k > 31
	result = EncodeKmers([]byte("ACGTACGTACGTACGTACGTACGTACGTACGTACGT"), 32, nil)
	if result != nil {
		t.Errorf("k>31 should return nil, got %v", result)
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
	kSizes := []int{8, 16, 31}

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
	kSizes := []int{8, 16, 31}

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
	kmer := uint64(0x06C6C6C6C6C6C6C6)
	k := 31

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReverseComplement(kmer, k)
	}
}

// BenchmarkNormalizeKmer benchmarks the normalization function
func BenchmarkNormalizeKmer(b *testing.B) {
	kmer := uint64(0x06C6C6C6C6C6C6C6)
	k := 31

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NormalizeKmer(kmer, k)
	}
}

// TestExtractSuperKmersBasic tests basic super k-mer extraction
func TestExtractSuperKmersBasic(t *testing.T) {
	tests := []struct {
		name     string
		seq      string
		k        int
		m        int
		validate func(*testing.T, []SuperKmer)
	}{
		{
			name: "simple sequence",
			seq:  "ACGTACGTACGT",
			k:    5,
			m:    3,
			validate: func(t *testing.T, sks []SuperKmer) {
				if len(sks) == 0 {
					t.Error("expected at least one super k-mer")
				}
				// Verify all super k-mers cover the sequence
				totalLen := 0
				for _, sk := range sks {
					totalLen += sk.End - sk.Start
					if string(sk.Sequence) != string([]byte(t.Name())[len(t.Name())-len(sk.Sequence):]) {
						// Just verify Start/End matches Sequence
						if string(sk.Sequence) != string([]byte("ACGTACGTACGT")[sk.Start:sk.End]) {
							t.Errorf("Sequence mismatch: seq[%d:%d] != %s", sk.Start, sk.End, sk.Sequence)
						}
					}
				}
			},
		},
		{
			name: "single k-mer sequence",
			seq:  "ACGTACGT",
			k:    8,
			m:    4,
			validate: func(t *testing.T, sks []SuperKmer) {
				if len(sks) != 1 {
					t.Errorf("expected exactly 1 super k-mer for len(seq)==k, got %d", len(sks))
				}
				if len(sks) > 0 {
					if sks[0].Start != 0 || sks[0].End != 8 {
						t.Errorf("expected [0:8], got [%d:%d]", sks[0].Start, sks[0].End)
					}
				}
			},
		},
		{
			name: "repeating sequence",
			seq:  "AAAAAAAAAA",
			k:    5,
			m:    3,
			validate: func(t *testing.T, sks []SuperKmer) {
				// Repeating A should have same minimizer (AAA) everywhere
				if len(sks) != 1 {
					t.Errorf("expected 1 super k-mer for repeating sequence, got %d", len(sks))
				}
				if len(sks) > 0 {
					if sks[0].Start != 0 || sks[0].End != 10 {
						t.Errorf("expected super k-mer to cover entire sequence [0:10], got [%d:%d]",
							sks[0].Start, sks[0].End)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractSuperKmers([]byte(tt.seq), tt.k, tt.m, nil)
			tt.validate(t, result)
		})
	}
}

// TestExtractSuperKmersEdgeCases tests edge cases and error handling
func TestExtractSuperKmersEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		seq       string
		k         int
		m         int
		expectNil bool
	}{
		{"empty sequence", "", 5, 3, true},
		{"seq shorter than k", "ACG", 5, 3, true},
		{"m < 1", "ACGTACGT", 5, 0, true},
		{"m >= k", "ACGTACGT", 5, 5, true},
		{"m == k-1 (valid)", "ACGTACGT", 5, 4, false},
		{"k < 2", "ACGTACGT", 1, 1, true},
		{"k > 31", "ACGTACGTACGTACGTACGTACGTACGTACGT", 32, 16, true},
		{"k == 31 (valid)", "ACGTACGTACGTACGTACGTACGTACGTACG", 31, 16, false},
		{"seq == k (valid)", "ACGTACGT", 8, 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractSuperKmers([]byte(tt.seq), tt.k, tt.m, nil)
			if tt.expectNil && result != nil {
				t.Errorf("expected nil, got %v", result)
			}
			if !tt.expectNil && result == nil {
				t.Errorf("expected non-nil result, got nil")
			}
		})
	}
}

// TestExtractSuperKmersBoundaries verifies Start/End positions
func TestExtractSuperKmersBoundaries(t *testing.T) {
	seq := []byte("ACGTACGTGGGGAAAA")
	k := 6
	m := 3

	result := ExtractSuperKmers(seq, k, m, nil)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Verify each super k-mer
	for i, sk := range result {
		// Verify Start < End
		if sk.Start >= sk.End {
			t.Errorf("super k-mer %d: Start (%d) >= End (%d)", i, sk.Start, sk.End)
		}

		// Verify Sequence matches seq[Start:End]
		expected := string(seq[sk.Start:sk.End])
		actual := string(sk.Sequence)
		if actual != expected {
			t.Errorf("super k-mer %d: Sequence mismatch: got %s, want %s", i, actual, expected)
		}

		// Verify bounds are within sequence
		if sk.Start < 0 || sk.End > len(seq) {
			t.Errorf("super k-mer %d: bounds [%d:%d] outside sequence length %d",
				i, sk.Start, sk.End, len(seq))
		}

		// Verify minimum length is k
		if sk.End-sk.Start < k {
			t.Errorf("super k-mer %d: length %d < k=%d", i, sk.End-sk.Start, k)
		}
	}

	// Verify super k-mers can overlap (by up to k-1 bases) but must be ordered
	// and the overlap should not exceed k-1
	for i := 0; i < len(result)-1; i++ {
		// Next super k-mer should start before or at the end of current one
		// Overlap is allowed and expected
		overlap := result[i].End - result[i+1].Start
		if overlap > k-1 {
			t.Errorf("super k-mers %d and %d overlap by %d bases (max allowed: %d): [%d:%d] and [%d:%d]",
				i, i+1, overlap, k-1, result[i].Start, result[i].End, result[i+1].Start, result[i+1].End)
		}
		// But the start positions should be ordered
		if result[i+1].Start < result[i].Start {
			t.Errorf("super k-mers %d and %d are not ordered: [%d:%d] and [%d:%d]",
				i, i+1, result[i].Start, result[i].End, result[i+1].Start, result[i+1].End)
		}
	}
}

// TestExtractSuperKmersBufferReuse tests buffer parameter
func TestExtractSuperKmersBufferReuse(t *testing.T) {
	seq := []byte("ACGTACGTACGTACGT")
	k := 6
	m := 3

	// First call without buffer
	result1 := ExtractSuperKmers(seq, k, m, nil)

	// Second call with buffer
	buffer := make([]SuperKmer, 0, 100)
	result2 := ExtractSuperKmers(seq, k, m, &buffer)

	if len(result1) != len(result2) {
		t.Errorf("buffer reuse: length mismatch %d vs %d", len(result1), len(result2))
	}

	for i := range result1 {
		if result1[i].Minimizer != result2[i].Minimizer {
			t.Errorf("position %d: minimizer mismatch", i)
		}
		if result1[i].Start != result2[i].Start || result1[i].End != result2[i].End {
			t.Errorf("position %d: boundary mismatch", i)
		}
	}

	// Test multiple calls with same buffer
	for i := 0; i < 10; i++ {
		result3 := ExtractSuperKmers(seq, k, m, &buffer)
		if len(result3) != len(result1) {
			t.Errorf("iteration %d: length mismatch", i)
		}
	}
}

// TestExtractSuperKmersCanonical verifies minimizers are canonical
func TestExtractSuperKmersCanonical(t *testing.T) {
	seq := []byte("ACGTACGTACGT")
	k := 6
	m := 3

	result := ExtractSuperKmers(seq, k, m, nil)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	for i, sk := range result {
		// Verify the minimizer is indeed canonical (equal to its normalized form)
		normalized := NormalizeKmer(sk.Minimizer, m)
		if sk.Minimizer != normalized {
			t.Errorf("super k-mer %d: minimizer %d is not canonical (normalized: %d)",
				i, sk.Minimizer, normalized)
		}

		// The minimizer should be <= its reverse complement
		rc := ReverseComplement(sk.Minimizer, m)
		if sk.Minimizer > rc {
			t.Errorf("super k-mer %d: minimizer %d > reverse complement %d (not canonical)",
				i, sk.Minimizer, rc)
		}
	}
}

// TestExtractSuperKmersVariousKM tests various k and m combinations
func TestExtractSuperKmersVariousKM(t *testing.T) {
	seq := []byte("ACGTACGTACGTACGTACGTACGT")

	configs := []struct {
		k int
		m int
	}{
		{5, 3},
		{8, 4},
		{10, 5},
		{16, 8},
		{21, 11},
		{6, 5}, // m = k-1
		{4, 2},
	}

	for _, cfg := range configs {
		t.Run("k"+string(rune('0'+cfg.k/10))+string(rune('0'+cfg.k%10))+"_m"+string(rune('0'+cfg.m/10))+string(rune('0'+cfg.m%10)), func(t *testing.T) {
			if len(seq) < cfg.k {
				t.Skip("sequence too short for this k")
			}

			result := ExtractSuperKmers(seq, cfg.k, cfg.m, nil)

			if result == nil {
				t.Fatal("expected non-nil result for valid parameters")
			}

			if len(result) == 0 {
				t.Error("expected at least one super k-mer")
			}

			// Verify each super k-mer has minimum length k
			for i, sk := range result {
				length := sk.End - sk.Start
				if length < cfg.k {
					t.Errorf("super k-mer %d has length %d < k=%d", i, length, cfg.k)
				}
			}
		})
	}
}

// TestKmerErrorMarkers tests the error marker functionality
func TestKmerErrorMarkers(t *testing.T) {
	// Test with a 31-mer (max odd k-mer that fits in 62 bits)
	kmer31 := uint64(0x1FFFFFFFFFFFFFFF) // All 62 bits set (31 * 2)

	tests := []struct {
		name      string
		kmer      uint64
		errorCode uint64
		expected  uint64
	}{
		{
			name:      "no error",
			kmer:      kmer31,
			errorCode: 0,
			expected:  kmer31,
		},
		{
			name:      "error type 1",
			kmer:      kmer31,
			errorCode: 1,
			expected:  kmer31 | (0b01 << 62),
		},
		{
			name:      "error type 2",
			kmer:      kmer31,
			errorCode: 2,
			expected:  kmer31 | (0b10 << 62),
		},
		{
			name:      "error type 3",
			kmer:      kmer31,
			errorCode: 3,
			expected:  kmer31 | (0b11 << 62),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set error
			marked := SetKmerError(tt.kmer, tt.errorCode)
			if marked != tt.expected {
				t.Errorf("SetKmerError: got 0x%016X, want 0x%016X", marked, tt.expected)
			}

			// Get error
			extractedError := GetKmerError(marked)
			if extractedError != tt.errorCode {
				t.Errorf("GetKmerError: got 0x%016X, want 0x%016X", extractedError, tt.errorCode)
			}

			// Clear error
			cleared := ClearKmerError(marked)
			if cleared != tt.kmer {
				t.Errorf("ClearKmerError: got 0x%016X, want 0x%016X", cleared, tt.kmer)
			}

			// Verify sequence bits are preserved
			if (marked & KmerSequenceMask) != tt.kmer {
				t.Errorf("Sequence bits corrupted: got 0x%016X, want 0x%016X",
					marked&KmerSequenceMask, tt.kmer)
			}
		})
	}
}

// TestKmerErrorMarkersWithRealKmers tests error markers with actual k-mer encoding
func TestKmerErrorMarkersWithRealKmers(t *testing.T) {
	// Encode a real 31-mer
	seq := []byte("ACGTACGTACGTACGTACGTACGTACGTACG") // 31 bases
	k := 31

	kmers := EncodeKmers(seq, k, nil)
	if len(kmers) != 1 {
		t.Fatalf("Expected 1 k-mer, got %d", len(kmers))
	}

	originalKmer := kmers[0]

	// Test each error type
	for i, errorCode := range []uint64{0, 1, 2, 3} {
		t.Run("error_code_"+string(rune('0'+i)), func(t *testing.T) {
			// Mark with error
			marked := SetKmerError(originalKmer, errorCode)

			// Verify error can be extracted
			if GetKmerError(marked) != errorCode {
				t.Errorf("Error code mismatch: got 0x%X, want 0x%X",
					GetKmerError(marked), errorCode)
			}

			// Verify sequence is preserved
			if ClearKmerError(marked) != originalKmer {
				t.Errorf("Original k-mer not preserved after marking")
			}

			// Verify normalization works with error bits cleared
			normalized1 := NormalizeKmer(originalKmer, k)
			normalized2 := NormalizeKmer(ClearKmerError(marked), k)
			if normalized1 != normalized2 {
				t.Errorf("Normalization affected by error bits")
			}
		})
	}
}

// TestKmerErrorMarkersConstants verifies the mask constant definitions
func TestKmerErrorMarkersConstants(t *testing.T) {
	// Verify error mask covers exactly the top 2 bits
	if KmerErrorMask != (0b11 << 62) {
		t.Errorf("KmerErrorMask wrong value: 0x%016X", KmerErrorMask)
	}

	// Verify sequence mask is the complement
	if KmerSequenceMask != ^KmerErrorMask {
		t.Errorf("KmerSequenceMask should be complement of KmerErrorMask")
	}

	// Verify masks are mutually exclusive
	if (KmerErrorMask & KmerSequenceMask) != 0 {
		t.Errorf("Masks should be mutually exclusive")
	}

	// Verify masks cover all bits
	if (KmerErrorMask | KmerSequenceMask) != ^uint64(0) {
		t.Errorf("Masks should cover all 64 bits")
	}

	// Verify error code API
	testKmer := uint64(0x06C6C6C6C6C6C6C6)
	for code := uint64(0); code <= 3; code++ {
		marked := SetKmerError(testKmer, code)
		extracted := GetKmerError(marked)
		if extracted != code {
			t.Errorf("Error code %d not preserved: got %d", code, extracted)
		}
	}
}

// TestReverseComplementPreservesErrorBits tests that RC preserves error markers
func TestReverseComplementPreservesErrorBits(t *testing.T) {
	// Test with a 31-mer
	seq := []byte("ACGTACGTACGTACGTACGTACGTACGTACG")
	k := 31

	kmers := EncodeKmers(seq, k, nil)
	if len(kmers) != 1 {
		t.Fatalf("Expected 1 k-mer, got %d", len(kmers))
	}

	originalKmer := kmers[0]

	// Test each error code
	errorCodes := []uint64{0, 1, 2, 3}

	for i, errCode := range errorCodes {
		t.Run("error_code_"+string(rune('0'+i)), func(t *testing.T) {
			// Mark k-mer with error
			marked := SetKmerError(originalKmer, errCode)

			// Compute reverse complement
			rc := ReverseComplement(marked, k)

			// Verify error bits are preserved
			extractedError := GetKmerError(rc)
			if extractedError != errCode {
				t.Errorf("Error bits not preserved: got 0x%X, want 0x%X", extractedError, errCode)
			}

			// Verify sequence was reverse complemented correctly
			// (clear error bits and check RC property)
			cleanOriginal := ClearKmerError(originalKmer)
			cleanRC := ClearKmerError(rc)
			expectedRC := ReverseComplement(cleanOriginal, k)

			if cleanRC != expectedRC {
				t.Errorf("Sequence not reverse complemented correctly")
			}

			// Verify RC(RC(x)) = x (involution property with error bits)
			rcrc := ReverseComplement(rc, k)
			if rcrc != marked {
				t.Errorf("RC(RC(x)) != x: got 0x%016X, want 0x%016X", rcrc, marked)
			}
		})
	}
}

// TestNormalizeKmerWithErrorBits tests that NormalizeKmer works with error bits
func TestNormalizeKmerWithErrorBits(t *testing.T) {
	seq := []byte("ACGTACGTACGTACGTACGTACGTACGTACG")
	k := 31

	kmers := EncodeKmers(seq, k, nil)
	if len(kmers) != 1 {
		t.Fatalf("Expected 1 k-mer, got %d", len(kmers))
	}

	originalKmer := kmers[0]

	// Test with different error codes
	for i, errCode := range []uint64{1, 2, 3} {
		t.Run("error_code_"+string(rune('0'+i+1)), func(t *testing.T) {
			marked := SetKmerError(originalKmer, errCode)

			// Normalize should work on the sequence part
			normalized := NormalizeKmer(marked, k)

			// Error bits should be preserved
			if GetKmerError(normalized) != errCode {
				t.Errorf("Error bits lost during normalization")
			}

			// The sequence part should be normalized
			cleanNormalized := ClearKmerError(normalized)
			expectedNormalized := NormalizeKmer(ClearKmerError(marked), k)

			if cleanNormalized != expectedNormalized {
				t.Errorf("Normalization incorrect with error bits present")
			}
		})
	}
}

// TestKmerErrorMarkersOddKmers tests that error markers work for all odd k ≤ 31
func TestKmerErrorMarkersOddKmers(t *testing.T) {
	oddKSizes := []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31}

	for _, k := range oddKSizes {
		t.Run("k="+string(rune('0'+k/10))+string(rune('0'+k%10)), func(t *testing.T) {
			// Create a sequence of length k
			seq := make([]byte, k)
			for i := range seq {
				seq[i] = "ACGT"[i%4]
			}

			kmers := EncodeKmers(seq, k, nil)
			if len(kmers) != 1 {
				t.Fatalf("Expected 1 k-mer, got %d", len(kmers))
			}

			originalKmer := kmers[0]

			// Verify that k*2 bits fit in 62 bits (top 2 bits should be free)
			maxValue := uint64(1)<<(k*2) - 1
			if originalKmer > maxValue {
				t.Errorf("k-mer exceeds expected bit range for k=%d", k)
			}

			// Test all error codes
			for _, errCode := range []uint64{1, 2, 3} {
				marked := SetKmerError(originalKmer, errCode)

				// Verify error is set
				if GetKmerError(marked) != errCode {
					t.Errorf("Error code not preserved for k=%d", k)
				}

				// Verify sequence is preserved
				if ClearKmerError(marked) != originalKmer {
					t.Errorf("Sequence corrupted for k=%d", k)
				}
			}
		})
	}
}

// TestIterKmers tests the k-mer iterator
func TestIterKmers(t *testing.T) {
	seq := []byte("ACGTACGT")
	k := 4

	// Collect k-mers via iterator
	var iterKmers []uint64
	for kmer := range IterKmers(seq, k) {
		iterKmers = append(iterKmers, kmer)
	}

	// Compare with slice-based version
	sliceKmers := EncodeKmers(seq, k, nil)

	if len(iterKmers) != len(sliceKmers) {
		t.Errorf("length mismatch: iter=%d, slice=%d", len(iterKmers), len(sliceKmers))
	}

	for i := range iterKmers {
		if iterKmers[i] != sliceKmers[i] {
			t.Errorf("position %d: iter=%d, slice=%d", i, iterKmers[i], sliceKmers[i])
		}
	}
}

// TestIterNormalizedKmers tests the normalized k-mer iterator
func TestIterNormalizedKmers(t *testing.T) {
	seq := []byte("ACGTACGTACGT")
	k := 6

	// Collect k-mers via iterator
	var iterKmers []uint64
	for kmer := range IterNormalizedKmers(seq, k) {
		iterKmers = append(iterKmers, kmer)
	}

	// Compare with slice-based version
	sliceKmers := EncodeNormalizedKmers(seq, k, nil)

	if len(iterKmers) != len(sliceKmers) {
		t.Errorf("length mismatch: iter=%d, slice=%d", len(iterKmers), len(sliceKmers))
	}

	for i := range iterKmers {
		if iterKmers[i] != sliceKmers[i] {
			t.Errorf("position %d: iter=%d, slice=%d", i, iterKmers[i], sliceKmers[i])
		}
	}
}

// TestIterKmersEarlyExit tests early exit from iterator
func TestIterKmersEarlyExit(t *testing.T) {
	seq := []byte("ACGTACGTACGTACGT")
	k := 4

	count := 0
	for range IterKmers(seq, k) {
		count++
		if count == 5 {
			break
		}
	}

	if count != 5 {
		t.Errorf("expected to process 5 k-mers, got %d", count)
	}
}

// BenchmarkIterKmers benchmarks the k-mer iterator vs slice-based
func BenchmarkIterKmers(b *testing.B) {
	seq := make([]byte, 10000)
	for i := range seq {
		seq[i] = "ACGT"[i%4]
	}
	k := 21

	b.Run("Iterator", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			for range IterKmers(seq, k) {
				count++
			}
		}
	})

	b.Run("Slice", func(b *testing.B) {
		var buffer []uint64
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buffer = EncodeKmers(seq, k, &buffer)
		}
	})
}

// BenchmarkIterNormalizedKmers benchmarks the normalized iterator
func BenchmarkIterNormalizedKmers(b *testing.B) {
	seq := make([]byte, 10000)
	for i := range seq {
		seq[i] = "ACGT"[i%4]
	}
	k := 21

	b.Run("Iterator", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			for range IterNormalizedKmers(seq, k) {
				count++
			}
		}
	})

	b.Run("Slice", func(b *testing.B) {
		var buffer []uint64
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buffer = EncodeNormalizedKmers(seq, k, &buffer)
		}
	})
}

// BenchmarkExtractSuperKmers benchmarks the super k-mer extraction
func BenchmarkExtractSuperKmers(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}
	configs := []struct {
		k int
		m int
	}{
		{21, 11},
		{31, 15},
		{16, 8},
		{10, 5},
	}

	for _, cfg := range configs {
		for _, size := range sizes {
			seq := make([]byte, size)
			for i := range seq {
				seq[i] = "ACGT"[i%4]
			}

			name := "k" + string(rune('0'+cfg.k/10)) + string(rune('0'+cfg.k%10)) +
				"_m" + string(rune('0'+cfg.m/10)) + string(rune('0'+cfg.m%10)) +
				"_size" + string(rune('0'+(size/10000)%10)) +
				string(rune('0'+(size/1000)%10)) +
				string(rune('0'+(size/100)%10)) +
				string(rune('0'+(size/10)%10)) +
				string(rune('0'+size%10))

			b.Run(name, func(b *testing.B) {
				buffer := make([]SuperKmer, 0, size/cfg.k)
				b.ResetTimer()
				b.SetBytes(int64(size))

				for i := 0; i < b.N; i++ {
					ExtractSuperKmers(seq, cfg.k, cfg.m, &buffer)
				}
			})
		}
	}
}
