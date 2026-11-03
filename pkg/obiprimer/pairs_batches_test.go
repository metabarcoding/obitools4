package obiprimer

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func TestAmpBatches(t *testing.T) {
	cases := []struct {
		count  []int
		budget int
		want   [][2]int
	}{
		{[]int{3, 3, 3}, 0, [][2]int{{0, 3}}},
		{[]int{3, 3, 3}, 100, [][2]int{{0, 3}}},
		{[]int{3, 3, 3}, 6, [][2]int{{0, 2}, {2, 3}}},
		{[]int{3, 3, 3}, 1, [][2]int{{0, 1}, {1, 2}, {2, 3}}}, // a pair above the budget goes alone
		{[]int{9, 1, 1}, 5, [][2]int{{0, 1}, {1, 3}}},
		{nil, 5, nil},
		{[]int{0, 0}, 5, [][2]int{{0, 2}}},
	}
	for _, c := range cases {
		if got := ampBatches(c.count, c.budget); !reflect.DeepEqual(got, c.want) {
			t.Errorf("ampBatches(%v, %d) = %v, want %v", c.count, c.budget, got, c.want)
		}
	}
}

// The batches of phase 2 only bound the memory: whatever their size, the pairs,
// their indices and their sites on the reference are the same.
func TestAmpBudgetDoesNotChangeTheResult(t *testing.T) {
	s := newSynthetic(10, -1, 5)
	// non-targets sharing barcodes with targets, and counterexamples
	for k, role := range []Role{Other, Other, CounterExample, Other, CounterExample} {
		cp := append([]byte(nil), s.corpus.Seqs[k*2].Sequence()...)
		if role == CounterExample {
			for i := s.pos; i < s.pos+10; i++ { // no forward site: not amplified
				cp[i] = "cgta"[strings.IndexByte("acgt", cp[i])]
			}
		}
		s.corpus.Seqs = append(s.corpus.Seqs, obiseq.NewBioSequence(strings.Repeat("x", k+1), cp, ""))
		s.corpus.Roles = append(s.corpus.Roles, role)
		s.taxa.Of = append(s.taxa.Of, int32(100+k%3))
	}
	// a few unrelated sequences, with a taxon of their own or none
	rng := rand.New(rand.NewSource(9))
	for k := 0; k < 3; k++ {
		s.corpus.Seqs = append(s.corpus.Seqs, obiseq.NewBioSequence("u"+strings.Repeat("y", k), []byte(strings.ToLower(randDNA(rng, 300))), ""))
		s.corpus.Roles = append(s.corpus.Roles, Other)
		s.taxa.Of = append(s.taxa.Of, int32(k-1))
	}

	run := func(budget int) []PairStats {
		p := s.params()
		p.MinSpecificity = 0 // keep every pair: the stress case of the batches
		p.FalsePositive = 1
		p.Reference = "s3"
		p.AmpBudget = budget
		pairs, info, err := Search(s.corpus, s.taxa, p)
		if err != nil {
			t.Fatal(err)
		}
		if info.Pair.Candidates < 2 {
			t.Fatalf("budget %d: only %d pairs reach phase 2b, the test needs several", budget, info.Pair.Candidates)
		}
		return pairs
	}
	want := run(0)
	for _, budget := range []int{1, 7, 50, 1000} {
		if got := run(budget); !reflect.DeepEqual(got, want) {
			t.Errorf("budget %d changes the result:\n got %d pairs\nwant %d pairs", budget, len(got), len(want))
		}
	}
}
