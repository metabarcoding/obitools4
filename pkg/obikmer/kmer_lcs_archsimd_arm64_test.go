//go:build goexperiment.simd && arm64

package obikmer

import (
	"math/rand"
	"testing"
)

func TestKmerLCSBatchNEONMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(5))
	for _, k := range []int{1, 2, 5, 11, 21, 31} {
		for _, n := range []int{0, 1, 7, 8, 9, 100} {
			query := randomKmer(rng, k)
			partners := make([]uint64, n)
			for i := range partners {
				partners[i] = randomKmer(rng, k)
			}
			out := make([]KmerLCSResult, n)
			kmerLCSBatchNEON(query, partners, k, out)
			for i, p := range partners {
				if want := KmerLCS(query, p, k); out[i] != want {
					t.Fatalf("k=%d n=%d partner %d: got %+v, want %+v", k, n, i, out[i], want)
				}
			}
		}
	}
}

func BenchmarkKmerLCSNEON21(b *testing.B) { benchBatch(b, 21, kmerLCSBatchNEON) }
func BenchmarkKmerLCSNEON11(b *testing.B) { benchBatch(b, 11, kmerLCSBatchNEON) }
