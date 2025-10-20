package obikmer

import "testing"

func TestNormalize(t *testing.T) {
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

		// Test avec k=4
		{"k=4 acgt", "acgt", "acgt"},
		{"k=4 cgta", "cgta", "acgt"},
		{"k=4 gtac", "gtac", "acgt"},
		{"k=4 tacg", "tacg", "acgt"},
		{"k=4 tgca", "tgca", "atgc"},

		// Test avec k=6
		{"k=6 aaaaaa", "aaaaaa", "aaaaaa"},
		{"k=6 tttttt", "tttttt", "tttttt"},

		// Test avec k>6 (calcul à la volée)
		{"k=7 aaaaaaa", "aaaaaaa", "aaaaaaa"},
		{"k=7 tgcatgc", "tgcatgc", "atgctgc"},
		{"k=7 gcatgct", "gcatgct", "atgctgc"},
		{"k=8 acgtacgt", "acgtacgt", "acgtacgt"},
		{"k=8 gtacgtac", "gtacgtac", "acgtacgt"},
		{"k=10 acgtacgtac", "acgtacgtac", "acacgtacgt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.kmer)
			if result != tt.expected {
				t.Errorf("Normalize(%q) = %q, want %q", tt.kmer, result, tt.expected)
			}
		})
	}
}

func TestNormalizeTableConsistency(t *testing.T) {
	// Vérifier que tous les kmers de la table donnent le bon résultat
	// en comparant avec le calcul à la volée
	for kmer, expected := range LexicographicNormalization {
		calculated := getCanonicalCircular(kmer)
		if calculated != expected {
			t.Errorf("Table inconsistency for %q: table=%q, calculated=%q",
				kmer, expected, calculated)
		}
	}
}

func BenchmarkNormalizeSmall(b *testing.B) {
	// Benchmark pour k<=6 (utilise la table)
	kmer := "acgtac"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Normalize(kmer)
	}
}

func BenchmarkNormalizeLarge(b *testing.B) {
	// Benchmark pour k>6 (calcul à la volée)
	kmer := "acgtacgtac"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Normalize(kmer)
	}
}
