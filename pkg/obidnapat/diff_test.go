package obidnapat

import (
	"math/rand"
	"sort"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
)

// TestAgainstObiapat compares the matches with those of obiapat on the common
// ground: sequences over acgt, where the two ways to read the ambiguity codes
// of a sequence cannot differ. The legacy matches are brought to the
// conventions of this package (the region is strict; a circular sequence
// reports each start once).
func TestAgainstObiapat(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	found := 0
	for iter := 0; iter < iterations(5000); iter++ {
		pat := randomPattern(rng, 1+rng.Intn(40), true)
		e := rng.Intn(11) // 0..10: the unrolled kernels (up to 8 levels) and the generic one
		legacy, err := obiapat.MakeApatPattern(pat, e, false)
		if err != nil {
			t.Fatalf("obiapat refuses %q: %v", pat, err)
		}
		p, err := MakeApatPattern(pat, e, false)
		if err != nil {
			t.Fatalf("pattern %q: %v", pat, err)
		}
		seq := randomSequence(rng, "acgt", rng.Intn(300), concrete(rng, p))
		bs := newSeq(seq)

		for _, circular := range []bool{false, true} {
			if circular && p.Len() > len(seq) {
				continue // a window reading the circle more than once: refused here, not by obiapat
			}
			begin, length := 0, -1
			if !circular {
				begin = rng.Intn(40)
				if rng.Intn(2) == 0 {
					length = rng.Intn(250)
				}
			}
			ls, err := obiapat.MakeApatSequence(newSeq(append([]byte(nil), seq...)), circular)
			if err != nil {
				t.Fatal(err)
			}
			n := len(seq)
			limit := n
			if length >= 0 && begin+length < n {
				limit = begin + length
			}
			var want [][3]int
			seen := map[int]bool{}
			for _, m := range legacy.FindAllIndex(ls, begin, length) {
				switch {
				case circular && m[0] >= 0 && m[0] < n && !seen[m[0]]:
					seen[m[0]] = true
					want = append(want, m)
				case !circular && m[0] >= begin && m[1] <= limit:
					want = append(want, m)
				}
			}
			sort.Slice(want, func(i, j int) bool { return want[i][0] < want[j][0] })

			var got [][3]int
			if circular {
				got = p.FindAllIndexCircular(bs, begin, length)
			} else {
				got = p.FindAllIndex(bs, begin, length)
			}
			if !sameMatches(got, nilIfEmpty(want)) {
				t.Fatalf("pattern %q e=%d circular=%v begin=%d length=%d seq=%q\n new     %v\n obiapat %v",
					pat, e, circular, begin, length, seq, got, want)
			}
			found += len(got)
		}
	}
	if found < 3000 {
		t.Errorf("the draws hold only %d matches", found)
	}
}
