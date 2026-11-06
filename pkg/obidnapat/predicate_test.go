package obidnapat

import (
	"math/rand"
	"strings"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
)

// TestIsMatchingIndelAgainstOracle: IsMatching of a pattern with indels is true when
// the dynamic programming finds a window, whatever the nominal start of the window.
func TestIsMatchingIndelAgainstOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(61))
	yes, atTheStart := 0, 0
	for iter := 0; iter < iterations(6000); iter++ {
		pat := randomPattern(rng, 1+rng.Intn(25), true)
		p, err := MakeApatPattern(pat, 1+rng.Intn(5), true)
		if err != nil {
			continue
		}
		seq := randomSequence(rng, "acgt", rng.Intn(120), mutate(rng, concrete(rng, p)))
		if rng.Intn(3) == 0 && len(seq) > 0 { // the pattern at the very start, with its first bases cut
			cut := 1 + rng.Intn(3)
			if c := concrete(rng, p); len(c) > cut {
				seq = append([]byte(c[cut:]), seq...)
			}
		}
		begin, length := 0, -1
		if rng.Intn(2) == 0 {
			begin, length = rng.Intn(20), rng.Intn(100)
		}
		want := len(oracleIndelRaw(p, seq, begin, length, false, false)) > 0
		if got := p.IsMatching(newSeq(seq), begin, length); got != want {
			t.Fatalf("pattern %q e=%d begin=%d length=%d seq=%q: %v, want %v", pat, p.MaxErrors(), begin, length, seq, got, want)
		}
		if want {
			yes++
			if nominal := oracleIndel(p, seq, begin, length, false); len(nominal) == 0 {
				atTheStart++ // found only through a window that starts before the region
			}
		}
	}
	if yes < 1500 || atTheStart < 50 {
		t.Errorf("the draws are not varied enough: %d matches, %d only with a nominal start before the region", yes, atTheStart)
	}
}

// TestIsPatternMatchSequenceAgainstObiapat: the predicate gives the answer of obiapat's
// (sequences over acgt, so the two readings of the ambiguity codes cannot differ).
// A pattern with a negation is compared on its own strand only: obiapat gets the
// reverse complement of "!" wrong, the matches of this package were checked by hand.
func TestIsPatternMatchSequenceAgainstObiapat(t *testing.T) {
	rng := rand.New(rand.NewSource(62))
	yes, no := 0, 0
	for iter := 0; iter < iterations(6000); iter++ {
		pat := randomPattern(rng, 1+rng.Intn(25), false)
		e, indel, both := rng.Intn(5), rng.Intn(2) == 0, rng.Intn(2) == 0
		if both {
			// obiapat reverse complements a negation badly ("!C#" becomes "g!#"): not compared
			pat = strings.ReplaceAll(pat, "!", "")
		}
		if q, err := MakeApatPattern(pat, e, indel); err != nil || e >= q.Len() {
			continue // (obiapat has no rule on the real bases of a window, see noise.go)
		}
		mine := IsPatternMatchSequence(pat, e, both, indel)
		theirs := obiapat.IsPatternMatchSequence(pat, e, both, indel)
		p, _ := MakeApatPattern(pat, e, indel)
		site := concrete(rng, p)
		if rng.Intn(2) == 0 {
			rc, _ := p.ReverseComplement()
			site = concrete(rng, rc)
		}
		seq := newSeq(randomSequence(rng, "acgt", rng.Intn(150), mutate(rng, site)))
		if indel && seq.Len() < p.Len()+e {
			continue // too short: a nominal window hangs over the start, and its missing bases are not real ones (see noise.go)
		}
		if got, want := mine(seq), theirs(seq); got != want {
			t.Fatalf("pattern %q e=%d indels=%v both=%v seq=%q: %v, obiapat %v", pat, e, indel, both, seq.String(), got, want)
		}
		if mine(seq) {
			yes++
		} else {
			no++
		}
	}
	if yes < 1000 || no < 1000 {
		t.Errorf("the draws are not varied enough: %d matches, %d without", yes, no)
	}
}
