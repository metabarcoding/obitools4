package obidnapat

import (
	"math/rand"
	"testing"
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
