package obikmer_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obigraph"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
)

func allKmers(k int) []uint64 {
	kmers := make([]uint64, 0, 1<<(2*uint(k)))
	for km := range obikmer.IterConstrainedKmers(k, obikmer.NoKmerConstraints()) {
		kmers = append(kmers, km)
	}
	return kmers
}

func TestDissimilarityFilterMatchesDistances(t *testing.T) {
	const k = 7
	rng := rand.New(rand.NewSource(1))
	kmers := make([]uint64, 300)
	for i := range kmers {
		kmers[i] = rng.Uint64() & (1<<(2*k) - 1)
	}
	candidates := make([]int32, 0, len(kmers))
	for i := 1; i < len(kmers); i++ {
		candidates = append(candidates, int32(i))
	}

	for _, metric := range []obikmer.KmerMetric{obikmer.HammingMetric, obikmer.LCSMetric} {
		for _, dmin := range []int{1, 3, 5} {
			filter := obikmer.NewKmerDissimilarityFilter(kmers, k, metric, dmin)
			got := filter(0, candidates, nil)

			var want []int32
			for _, c := range candidates {
				var d int
				if metric == obikmer.HammingMetric {
					d = obikmer.HammingDistance(kmers[0], kmers[c], k)
				} else {
					d = obikmer.KmerLCS(kmers[0], kmers[c], k).Distance()
				}
				if d >= dmin {
					want = append(want, c)
				}
			}
			if len(got) != len(want) {
				t.Fatalf("%v d>=%d: got %d vertices, want %d", metric, dmin, len(got), len(want))
			}
			for i := range got {
				if got[i] != want[i] {
					t.Fatalf("%v d>=%d: mismatch at %d", metric, dmin, i)
				}
			}
		}
	}
}

// A family of 5-mers pairwise at Hamming distance >= 3, built with FindClique.
func TestFindClique_HammingFamily(t *testing.T) {
	const k, dmin, target = 5, 3, 32
	kmers := allKmers(k)
	filter := obikmer.NewKmerDissimilarityFilter(kmers, k, obikmer.HammingMetric, dmin)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	r := obigraph.FindClique(ctx, len(kmers), filter, target, obigraph.CliqueOptions{Workers: 2})
	if !r.Reached {
		t.Fatalf("family of %d not found (best %d, exhausted %v)", target, len(r.Clique), r.Exhausted)
	}
	for i, a := range r.Clique {
		for _, b := range r.Clique[i+1:] {
			if obikmer.HammingDistance(kmers[a], kmers[b], k) < dmin {
				t.Fatalf("%d and %d are too close", kmers[a], kmers[b])
			}
		}
	}
}

func TestFindClique_LCSFamily(t *testing.T) {
	const k, dmin, target = 6, 3, 20
	kmers := allKmers(k)
	filter := obikmer.NewKmerDissimilarityFilter(kmers, k, obikmer.LCSMetric, dmin)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	r := obigraph.FindClique(ctx, len(kmers), filter, target, obigraph.CliqueOptions{Workers: 2})
	if !r.Reached {
		t.Fatalf("family of %d not found (best %d, exhausted %v)", target, len(r.Clique), r.Exhausted)
	}
	for i, a := range r.Clique {
		for _, b := range r.Clique[i+1:] {
			if obikmer.KmerLCS(kmers[a], kmers[b], k).Distance() < dmin {
				t.Fatalf("%d and %d are too close", kmers[a], kmers[b])
			}
		}
	}
}

func TestParseKmerMetric(t *testing.T) {
	if m, err := obikmer.ParseKmerMetric("LCS"); err != nil || m != obikmer.LCSMetric {
		t.Fatal("lcs not parsed")
	}
	if m, err := obikmer.ParseKmerMetric("hamming"); err != nil || m != obikmer.HammingMetric {
		t.Fatal("hamming not parsed")
	}
	if _, err := obikmer.ParseKmerMetric("levenshtein"); err == nil {
		t.Fatal("unknown metric must be refused")
	}
}
