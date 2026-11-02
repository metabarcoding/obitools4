package obikmer

import (
	"math/rand"
	"testing"
)

func randomKmer(rng *rand.Rand, k int) uint64 {
	return rng.Uint64() & (uint64(1)<<(2*uint(k)) - 1)
}

func TestHammingDistance(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for _, k := range []int{1, 2, 7, 21, 31} {
		for n := 0; n < 200; n++ {
			a, b := randomKmer(rng, k), randomKmer(rng, k)
			sa, sb := DecodeKmer(a, k, nil), DecodeKmer(b, k, nil)
			want := 0
			for i := range sa {
				if sa[i] != sb[i] {
					want++
				}
			}
			if got := HammingDistance(a, b, k); got != want {
				t.Fatalf("k=%d %s/%s: got %d, want %d", k, sa, sb, got, want)
			}
		}
	}
}

func TestHammingDistanceIgnoresHighBits(t *testing.T) {
	if got := HammingDistance(0xFF00, 0x0000, 4); got != 0 {
		t.Fatalf("bits above 2k must be ignored, got %d", got)
	}
}

func TestKmerLCSKnown(t *testing.T) {
	enc := func(s string) uint64 { return EncodeKmer([]byte(s), len(s)) }
	cases := []struct {
		a, b string
		want KmerLCSResult
	}{
		{"ACGT", "ACGT", KmerLCSResult{4, 4}},
		{"AAAA", "CCCC", KmerLCSResult{0, 4}},
		{"ACGT", "ACGA", KmerLCSResult{3, 4}},
		{"ACGT", "CGTA", KmerLCSResult{3, 5}}, // one shift: 1 indel on each side
	}
	for _, c := range cases {
		if got := KmerLCS(enc(c.a), enc(c.b), 4); got != c.want {
			t.Errorf("%s/%s: got %+v, want %+v", c.a, c.b, got, c.want)
		}
	}
	if got := KmerLCS(0, 0, 32); got.LCS != -1 {
		t.Errorf("k=32 must be rejected, got %+v", got)
	}
}

// The batch implementation (SIMD when built with GOEXPERIMENT=simd) must agree
// with the scalar reference for all k and for batch sizes around the lane count.
func TestKmerLCSBatchMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	for _, k := range []int{1, 2, 3, 5, 8, 16, 21, 30, 31} {
		for _, n := range []int{0, 1, 3, 7, 8, 9, 16, 17, 100} {
			query := randomKmer(rng, k)
			partners := make([]uint64, n)
			for i := range partners {
				partners[i] = randomKmer(rng, k)
			}
			if n > 2 {
				partners[0] = query // identical
			}
			out := make([]KmerLCSResult, n)
			if !KmerLCSBatch(query, partners, k, out) {
				t.Fatalf("k=%d n=%d: batch refused", k, n)
			}
			for i, p := range partners {
				if want := KmerLCS(query, p, k); out[i] != want {
					t.Fatalf("k=%d n=%d partner %d: got %+v, want %+v", k, n, i, out[i], want)
				}
			}
		}
	}
}

func TestKmerLCSBatchRejectsBadInput(t *testing.T) {
	if KmerLCSBatch(0, []uint64{1}, 0, make([]KmerLCSResult, 1)) {
		t.Fatal("k=0 must be rejected")
	}
	if KmerLCSBatch(0, []uint64{1, 2}, 4, make([]KmerLCSResult, 1)) {
		t.Fatal("short output must be rejected")
	}
}

func benchBatch(b *testing.B, k int, f func(uint64, []uint64, int, []KmerLCSResult)) {
	rng := rand.New(rand.NewSource(3))
	partners := make([]uint64, 1024)
	for i := range partners {
		partners[i] = randomKmer(rng, k)
	}
	out := make([]KmerLCSResult, len(partners))
	query := randomKmer(rng, k)
	b.ResetTimer()
	for b.Loop() {
		f(query, partners, k, out)
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*len(partners)), "ns/pair")
}

func BenchmarkKmerLCSScalar21(b *testing.B) { benchBatch(b, 21, kmerLCSBatchScalar) }
func BenchmarkKmerLCSBatch21(b *testing.B)  { benchBatch(b, 21, kmerLCSBatch) }
func BenchmarkKmerLCSScalar11(b *testing.B) { benchBatch(b, 11, kmerLCSBatchScalar) }
func BenchmarkKmerLCSBatch11(b *testing.B)  { benchBatch(b, 11, kmerLCSBatch) }
