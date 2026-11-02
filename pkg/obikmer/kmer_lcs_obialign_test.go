package obikmer_test

import (
	"math/rand"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// KmerLCS must reproduce obialign.FastLCSScore on the decoded sequences.
func TestKmerLCSMatchesFastLCSScore(t *testing.T) {
	rng := rand.New(rand.NewSource(4))
	for _, k := range []int{1, 2, 4, 9, 15, 21, 31} {
		mask := uint64(1)<<(2*uint(k)) - 1
		for n := 0; n < 300; n++ {
			a, b := rng.Uint64()&mask, rng.Uint64()&mask
			sa := obiseq.NewBioSequence("a", obikmer.DecodeKmer(a, k, nil), "")
			sb := obiseq.NewBioSequence("b", obikmer.DecodeKmer(b, k, nil), "")
			lcs, alen := obialign.FastLCSScore(sa, sb, -1, nil)
			got := obikmer.KmerLCS(a, b, k)
			if got.LCS != lcs || got.AlignLen != alen {
				t.Fatalf("k=%d %s/%s: KmerLCS=%+v, FastLCSScore=(%d,%d)",
					k, sa.String(), sb.String(), got, lcs, alen)
			}
		}
	}
}
