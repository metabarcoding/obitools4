package obikmer

import (
	"fmt"
	"testing"
)

func TestEncodeDecodeKmer(t *testing.T) {
	tests := []struct {
		kmer string
		code int
	}{
		{"a", 0},
		{"c", 1},
		{"g", 2},
		{"t", 3},
		{"aa", 0},
		{"ac", 1},
		{"ca", 4},
		{"acgt", 27}, // 0b00011011
		{"cgta", 108}, // 0b01101100
		{"tttt", 255}, // 0b11111111
	}

	for _, tt := range tests {
		t.Run(tt.kmer, func(t *testing.T) {
			// Test encoding
			encoded := EncodeKmer(tt.kmer)
			if encoded != tt.code {
				t.Errorf("EncodeKmer(%q) = %d, want %d", tt.kmer, encoded, tt.code)
			}

			// Test decoding
			decoded := DecodeKmer(tt.code, len(tt.kmer))
			if decoded != tt.kmer {
				t.Errorf("DecodeKmer(%d, %d) = %q, want %q", tt.code, len(tt.kmer), decoded, tt.kmer)
			}
		})
	}
}

func TestNormalizeInt(t *testing.T) {
	tests := []struct {
		name     string
		kmer     string
		expected string
	}{
		// Test avec k=1
		{"k=1 a", "a", "a"},
		{"k=1 c", "c", "c"},

		// Test avec k=2
		{"k=2 ca", "ca", "ac"},
		{"k=2 ac", "ac", "ac"},
		{"k=2 ta", "ta", "at"},

		// Test avec k=4 - toutes les rotations de "acgt"
		{"k=4 acgt", "acgt", "acgt"},
		{"k=4 cgta", "cgta", "acgt"},
		{"k=4 gtac", "gtac", "acgt"},
		{"k=4 tacg", "tacg", "acgt"},

		// Test avec k=4 - rotations de "tgca"
		{"k=4 tgca", "tgca", "atgc"},
		{"k=4 gcat", "gcat", "atgc"},
		{"k=4 catg", "catg", "atgc"},
		{"k=4 atgc", "atgc", "atgc"},

		// Test avec k=3 - rotations de "atg"
		{"k=3 atg", "atg", "atg"},
		{"k=3 tga", "tga", "atg"},
		{"k=3 gat", "gat", "atg"},

		// Test avec k=6
		{"k=6 aaaaaa", "aaaaaa", "aaaaaa"},
		{"k=6 tttttt", "tttttt", "tttttt"},

		// Test avec k>6 (calcul à la volée)
		{"k=7 aaaaaaa", "aaaaaaa", "aaaaaaa"},
		{"k=7 tgcatgc", "tgcatgc", "atgctgc"},
		{"k=7 gcatgct", "gcatgct", "atgctgc"},
		{"k=8 acgtacgt", "acgtacgt", "acgtacgt"},
		{"k=8 gtacgtac", "gtacgtac", "acgtacgt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kmerCode := EncodeKmer(tt.kmer)
			expectedCode := EncodeKmer(tt.expected)

			result := NormalizeInt(kmerCode, len(tt.kmer))

			if result != expectedCode {
				resultKmer := DecodeKmer(result, len(tt.kmer))
				t.Errorf("NormalizeInt(%q) = %q (code %d), want %q (code %d)",
					tt.kmer, resultKmer, result, tt.expected, expectedCode)
			}
		})
	}
}

func TestNormalizeIntConsistencyWithString(t *testing.T) {
	// Vérifier que NormalizeInt donne le même résultat que Normalize
	// pour tous les k-mers de taille 1 à 4 (pour ne pas trop ralentir les tests)
	bases := []byte{'a', 'c', 'g', 't'}

	var testKmers func(current string, maxSize int)
	testKmers = func(current string, maxSize int) {
		if len(current) > 0 {
			// Test normalization
			normalizedStr := Normalize(current)
			normalizedStrCode := EncodeKmer(normalizedStr)

			kmerCode := EncodeKmer(current)
			normalizedInt := NormalizeInt(kmerCode, len(current))

			if normalizedInt != normalizedStrCode {
				normalizedIntStr := DecodeKmer(normalizedInt, len(current))
				t.Errorf("Inconsistency for %q: Normalize=%q (code %d), NormalizeInt=%q (code %d)",
					current, normalizedStr, normalizedStrCode, normalizedIntStr, normalizedInt)
			}
		}

		if len(current) < maxSize {
			for _, base := range bases {
				testKmers(current+string(base), maxSize)
			}
		}
	}

	testKmers("", 4) // Test jusqu'à k=4 pour rester raisonnable
}

func TestCircularRotations(t *testing.T) {
	// Test que toutes les rotations circulaires donnent le même canonical
	tests := []struct {
		kmers []string
		canonical string
	}{
		{[]string{"atg", "tga", "gat"}, "atg"},
		{[]string{"acgt", "cgta", "gtac", "tacg"}, "acgt"},
		{[]string{"tgca", "gcat", "catg", "atgc"}, "atgc"},
	}

	for _, tt := range tests {
		canonicalCode := EncodeKmer(tt.canonical)

		for _, kmer := range tt.kmers {
			kmerCode := EncodeKmer(kmer)
			result := NormalizeInt(kmerCode, len(kmer))

			if result != canonicalCode {
				resultKmer := DecodeKmer(result, len(kmer))
				t.Errorf("NormalizeInt(%q) = %q, want %q", kmer, resultKmer, tt.canonical)
			}
		}
	}
}

func BenchmarkNormalizeIntSmall(b *testing.B) {
	// Benchmark pour k<=6 (utilise la table)
	kmer := "acgtac"
	kmerCode := EncodeKmer(kmer)
	kmerSize := len(kmer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NormalizeInt(kmerCode, kmerSize)
	}
}

func BenchmarkNormalizeIntLarge(b *testing.B) {
	// Benchmark pour k>6 (calcul à la volée)
	kmer := "acgtacgtac"
	kmerCode := EncodeKmer(kmer)
	kmerSize := len(kmer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NormalizeInt(kmerCode, kmerSize)
	}
}

func BenchmarkEncodeKmer(b *testing.B) {
	kmer := "acgtacgt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeKmer(kmer)
	}
}

func TestCanonicalKmerCount(t *testing.T) {
	// Test exact counts for k=1 to 6
	tests := []struct {
		k        int
		expected int
	}{
		{1, 4},
		{2, 10},
		{3, 24},
		{4, 70},
		{5, 208},
		{6, 700},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("k=%d", tt.k), func(t *testing.T) {
			result := CanonicalKmerCount(tt.k)
			if result != tt.expected {
				t.Errorf("CanonicalKmerCount(%d) = %d, want %d", tt.k, result, tt.expected)
			}
		})
	}

	// Verify counts match table sizes
	for k := 1; k <= 6; k++ {
		// Count unique canonical codes in the table
		uniqueCodes := make(map[int]bool)
		for _, canonicalCode := range LexicographicNormalizationInt[k] {
			uniqueCodes[canonicalCode] = true
		}

		expected := len(uniqueCodes)
		result := CanonicalKmerCount(k)

		if result != expected {
			t.Errorf("CanonicalKmerCount(%d) = %d, but table has %d unique canonical codes",
				k, result, expected)
		}
	}
}

func TestNecklaceCountFormula(t *testing.T) {
	// Verify Moreau's formula gives the same results as hardcoded values for k=1 to 6
	// and compute exact values for k=7+
	tests := []struct {
		k        int
		expected int
	}{
		{1, 4},
		{2, 10},
		{3, 24},
		{4, 70},
		{5, 208},
		{6, 700},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("k=%d", tt.k), func(t *testing.T) {
			result := necklaceCount(tt.k, 4)
			if result != tt.expected {
				t.Errorf("necklaceCount(%d, 4) = %d, want %d", tt.k, result, tt.expected)
			}
		})
	}
}

func TestNecklaceCountByBruteForce(t *testing.T) {
	// Verify necklace count for k=7 and k=8 by brute force
	// Generate all 4^k k-mers and count unique normalized ones
	bases := []byte{'a', 'c', 'g', 't'}

	for _, k := range []int{7, 8} {
		t.Run(fmt.Sprintf("k=%d", k), func(t *testing.T) {
			unique := make(map[int]bool)

			// Generate all possible k-mers
			var generate func(current int, depth int)
			generate = func(current int, depth int) {
				if depth == k {
					// Normalize and add to set
					normalized := NormalizeInt(current, k)
					unique[normalized] = true
					return
				}

				for _, base := range bases {
					newCode := (current << 2) | int(EncodeNucleotide(base))
					generate(newCode, depth+1)
				}
			}

			generate(0, 0)

			bruteForceCount := len(unique)
			formulaCount := necklaceCount(k, 4)

			if bruteForceCount != formulaCount {
				t.Errorf("For k=%d: brute force count = %d, formula count = %d",
					k, bruteForceCount, formulaCount)
			}

			t.Logf("k=%d: unique canonical k-mers = %d (formula matches brute force)", k, bruteForceCount)
		})
	}
}

func TestEulerTotient(t *testing.T) {
	tests := []struct {
		n        int
		expected int
	}{
		{1, 1},
		{2, 1},
		{3, 2},
		{4, 2},
		{5, 4},
		{6, 2},
		{7, 6},
		{8, 4},
		{9, 6},
		{10, 4},
		{12, 4},
		{15, 8},
		{20, 8},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("φ(%d)", tt.n), func(t *testing.T) {
			result := eulerTotient(tt.n)
			if result != tt.expected {
				t.Errorf("eulerTotient(%d) = %d, want %d", tt.n, result, tt.expected)
			}
		})
	}
}

func TestDivisors(t *testing.T) {
	tests := []struct {
		n        int
		expected []int
	}{
		{1, []int{1}},
		{2, []int{1, 2}},
		{6, []int{1, 2, 3, 6}},
		{12, []int{1, 2, 3, 4, 6, 12}},
		{15, []int{1, 3, 5, 15}},
		{20, []int{1, 2, 4, 5, 10, 20}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("divisors(%d)", tt.n), func(t *testing.T) {
			result := divisors(tt.n)
			if len(result) != len(tt.expected) {
				t.Errorf("divisors(%d) = %v, want %v", tt.n, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("divisors(%d) = %v, want %v", tt.n, result, tt.expected)
					return
				}
			}
		})
	}
}
