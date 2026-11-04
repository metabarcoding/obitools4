package obidnapat

import (
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// oracle is the simplest possible search, window by window, base by base: the
// definition of the matches, against which the bit-parallel kernels are checked.
// It shares nothing with them but the parsed positions, which the tests below
// check on their own (hand written cases, and obiapat).
func oracle(p ApatPattern, seq []byte, begin, length int, circular bool) [][3]int {
	var res [][3]int
	n, ps := len(seq), p.c.positions
	m := len(ps)
	if begin < 0 {
		begin = 0
	}
	if begin >= n || (circular && m > n) {
		return nil
	}
	limit := n
	if length >= 0 && begin+length < n {
		limit = begin + length
	}
	for s := begin; s < limit; s++ {
		if !circular && s+m > limit {
			break
		}
		errs, ok := 0, true
		for i, pos := range ps {
			j := s + i
			if circular {
				j %= n
			}
			if pos.wildcard || letterSet[seq[j]]&pos.set != 0 {
				continue
			}
			if pos.oblig {
				ok = false
				break
			}
			errs++
		}
		if ok && errs <= p.c.maxErr {
			res = append(res, [3]int{s, s + m, errs})
		}
	}
	return res
}

// iterations is the number of random draws of a test; OBIDNAPAT_ITER raises it.
func iterations(def int) int {
	if v, err := strconv.Atoi(os.Getenv("OBIDNAPAT_ITER")); err == nil && v > 0 {
		return v
	}
	return def
}

func sameMatches(a, b [][3]int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// randomPattern writes a pattern of up to maxLen positions in the syntax of the
// package: IUPAC letters, classes, negations, # and wildcards at the ends.
func randomPattern(rng *rand.Rand, maxLen int, flanks bool) string {
	const letters = "ACGTACGTACGTRYSWKMBDHVN"
	var b strings.Builder
	if flanks && rng.Intn(3) == 0 {
		b.WriteString(strings.Repeat("N", rng.Intn(12)))
	}
	for i, n := 0, 1+rng.Intn(maxLen); i < n; i++ {
		switch r := rng.Intn(12); {
		case r == 0:
			b.WriteString("[" + string(letters[rng.Intn(len(letters))]) + string(letters[rng.Intn(len(letters))]) + "]")
		case r == 1:
			b.WriteString("!" + string(letters[rng.Intn(len(letters))]))
		default:
			b.WriteByte(letters[rng.Intn(len(letters))])
		}
		if rng.Intn(15) == 0 {
			b.WriteByte('#')
		}
	}
	if flanks && rng.Intn(3) == 0 {
		b.WriteString(strings.Repeat("N", rng.Intn(12)))
	}
	return b.String()
}

// randomSequence draws a sequence over alphabet, and plants copies of window.
func randomSequence(rng *rand.Rand, alphabet string, n int, plant string) []byte {
	s := make([]byte, n)
	for i := range s {
		s[i] = alphabet[rng.Intn(len(alphabet))]
	}
	for k := 0; k < 3 && plant != "" && len(plant) < n; k++ {
		copy(s[rng.Intn(n-len(plant)+1):], plant)
	}
	return s
}

// concrete turns a pattern into a sequence that it matches, with some errors.
func concrete(rng *rand.Rand, p ApatPattern) string {
	b := make([]byte, 0, p.Len())
	for _, pos := range p.c.positions {
		c := "acgt"[rng.Intn(4)]
		for tries := 0; tries < 8 && !pos.wildcard && letterSet[c]&pos.set == 0; tries++ {
			c = "acgt"[rng.Intn(4)]
		}
		if rng.Intn(10) == 0 {
			c = "acgt"[rng.Intn(4)]
		}
		b = append(b, c)
	}
	return string(b)
}

func newSeq(s []byte) *obiseq.BioSequence { return obiseq.NewBioSequence("s", s, "") }

// TestAgainstOracle checks the kernels (every number of errors, linear and
// circular, regions, wildcards at the ends, # ) against the naive search, on
// sequences that carry ambiguity codes and characters outside the alphabet.
func TestAgainstOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	found, empty := 0, 0
	defer func() {
		if found < 5000 || empty < 1000 { // (for the default number of draws)
			t.Errorf("the draws are not varied enough: %d matches, %d empty results", found, empty)
		}
	}()
	for iter := 0; iter < iterations(6000); iter++ {
		pat := randomPattern(rng, 1+rng.Intn(40), true)
		e := rng.Intn(11) // 0..10: the unrolled kernels (up to 8 levels) and the generic one
		p, err := MakeApatPattern(pat, e, false)
		if err != nil {
			t.Fatalf("pattern %q: %v", pat, err)
		}
		alphabet := []string{"acgt", "acgtn", "acgtrykmn-", "ac"}[rng.Intn(4)]
		seq := randomSequence(rng, alphabet, rng.Intn(260), concrete(rng, p))
		begin, length := -1+rng.Intn(60), -1
		if rng.Intn(2) == 0 {
			length = rng.Intn(200)
		}
		for _, circular := range []bool{false, true} {
			want := oracle(p, seq, begin, length, circular)
			got := p.c.search(nil, seq, begin, length, circular)
			if !sameMatches(got, want) {
				t.Fatalf("pattern %q e=%d circular=%v begin=%d length=%d seq=%q\n got  %v\n want %v",
					pat, e, circular, begin, length, seq, got, want)
			}
			found += len(want)
			if len(want) == 0 {
				empty++
			}
		}
	}
}
