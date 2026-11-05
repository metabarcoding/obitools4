package obidnapat

import (
	"math/rand"
	"sort"
	"strings"
	"testing"
)

// expectedSet is the reference of a PatternSet: the naive search of every
// pattern and strand taken separately, merged and sorted by (Start, Pattern,
// strand, +1 first) with the standard library sort, not with the code under test.
func expectedSet(t *testing.T, patterns []ApatPattern, both bool, seq []byte, begin, length int, circular bool) []Occurrence {
	var want []Occurrence
	add := func(i int, strand int8, p ApatPattern) {
		for _, m := range oracle(p, seq, begin, length, circular) {
			want = append(want, Occurrence{i, strand, m[0], m[1], m[2]})
		}
	}
	for i, p := range patterns {
		add(i, 1, p)
		if both {
			rc, err := p.ReverseComplement()
			if err != nil {
				t.Fatal(err)
			}
			add(i, -1, rc)
		}
	}
	sort.Slice(want, func(a, b int) bool {
		x, y := want[a], want[b]
		if x.Start != y.Start {
			return x.Start < y.Start
		}
		if x.Pattern != y.Pattern {
			return x.Pattern < y.Pattern
		}
		return x.Strand > y.Strand
	})
	return want
}

func sameOccurrences(a, b []Occurrence) bool {
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

// exactPattern writes a pattern of n positions whose ends are not wildcards.
func exactPattern(rng *rand.Rand, n int) string {
	const letters = "ACGTACGTACGTRYSWKMBDHV"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rng.Intn(len(letters))]
		if rng.Intn(10) == 0 {
			b[i] = 'N' // internal wildcards only: the ends stay in the core
		}
	}
	b[0], b[n-1] = letters[rng.Intn(4)], letters[rng.Intn(4)]
	return string(b)
}

func compileAll(t *testing.T, texts []string, errs []int) []ApatPattern {
	ps := make([]ApatPattern, len(texts))
	for i, tx := range texts {
		p, err := MakeApatPattern(tx, errs[i], false)
		if err != nil {
			t.Fatalf("pattern %q: %v", tx, err)
		}
		ps[i] = p
	}
	return ps
}

// testVector tells checkSet to run the words of 2 levels on SIMD vectors.
var testVector bool

// checkSet draws a sequence and a region, and compares the set with the
// reference, linear and circular. It returns the number of occurrences.
func checkSet(t *testing.T, rng *rand.Rand, ps []ApatPattern, both bool, label string) int {
	s, err := makePatternSet(ps, both, testVector)
	if err != nil {
		t.Fatal(err)
	}
	alphabet := []string{"acgt", "acgt", "acgtn", "acgtrykmn-", "ac"}[rng.Intn(5)]
	seq := randomSequence(rng, alphabet, rng.Intn(300), "")
	for k := 0; k < 4 && len(seq) > 0; k++ { // copies of the patterns
		c := concrete(rng, ps[rng.Intn(len(ps))])
		if len(c) < len(seq) {
			copy(seq[rng.Intn(len(seq)-len(c)+1):], c)
		}
	}
	begin, length := -1+rng.Intn(60), -1
	if rng.Intn(2) == 0 {
		length = rng.Intn(250)
	}
	found := 0
	for _, circular := range []bool{false, true} {
		want := expectedSet(t, ps, both, seq, begin, length, circular)
		got := s.appendAll(nil, seq, begin, length, circular)
		if !sameOccurrences(got, want) {
			var pats []string
			for _, p := range ps {
				pats = append(pats, p.String())
			}
			t.Fatalf("%s: patterns %q both=%v circular=%v begin=%d length=%d seq=%q words=%d\n got  %v\n want %v",
				label, pats, both, circular, begin, length, seq, s.Words(), got, want)
		}
		found += len(want)
	}
	return found
}

// TestPatternSetAgainstOracle checks sets of 1 to 40 patterns of any length and
// error level, one or two strands, linear and circular, regions, against the
// naive search of each pattern taken separately.
func TestPatternSetAgainstOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(21))
	found, multiWord := 0, 0
	for iter := 0; iter < iterations(3000); iter++ {
		var texts []string
		var errs []int
		sameErr := rng.Intn(2) == 0
		e0 := rng.Intn(11)
		for k, n := 0, 1+rng.Intn(40); k < n; k++ {
			tx := randomPattern(rng, []int{6, 12, 20, 63}[rng.Intn(4)], true)
			if _, err := MakeApatPattern(tx, 0, false); err != nil {
				continue // a core of more than 63 positions
			}
			texts = append(texts, tx)
			if sameErr {
				errs = append(errs, e0)
			} else {
				errs = append(errs, rng.Intn(11))
			}
		}
		if len(texts) == 0 {
			continue
		}
		ps := compileAll(t, texts, errs)
		found += checkSet(t, rng, ps, rng.Intn(2) == 0, "random set")
		if s, _ := MakePatternSet(ps, false); s.Words() > 1 {
			multiWord++
		}
	}
	if found < 3000 || multiWord < 100 {
		t.Errorf("the draws are not varied enough: %d occurrences, %d sets on several words", found, multiWord)
	}
}

// TestPatternSetWordBoundary packs patterns whose cores total exactly 63, 64 and
// 65 positions, with the same number of errors (so that they share words), and
// checks the number of words as well as the occurrences.
func TestPatternSetWordBoundary(t *testing.T) {
	rng := rand.New(rand.NewSource(22))
	for _, total := range []int{63, 64, 65} {
		for iter := 0; iter < iterations(300); iter++ {
			lens := make([]int, 2+rng.Intn(4))
			for i := range lens {
				lens[i] = 1
			}
			for left := total - len(lens); left > 0; {
				if i := rng.Intn(len(lens)); lens[i] < MaxCoreLength {
					lens[i]++
					left--
				}
			}
			var texts []string
			var errs []int
			e := rng.Intn(2) // one level count for all: they share words
			for _, l := range lens {
				texts = append(texts, exactPattern(rng, l))
				errs = append(errs, e)
			}
			ps := compileAll(t, texts, errs)
			s, err := MakePatternSet(ps, false)
			if err != nil {
				t.Fatal(err)
			}
			if u := s.Words() + len(s.solo); total <= wordBits && u != 1 || total > wordBits && u != 2 {
				t.Fatalf("cores %v (total %d): %d words, %d alone", lens, total, s.Words(), len(s.solo))
			}
			checkSet(t, rng, ps, false, "boundary")
		}
	}
}

// TestPatternSetSingle: a set of one pattern gives the matches of the pattern.
func TestPatternSetSingle(t *testing.T) {
	rng := rand.New(rand.NewSource(23))
	for iter := 0; iter < iterations(2000); iter++ {
		tx := randomPattern(rng, 1+rng.Intn(40), true)
		p, err := MakeApatPattern(tx, rng.Intn(9), false)
		if err != nil {
			continue
		}
		s, _ := MakePatternSet([]ApatPattern{p}, false)
		seq := randomSequence(rng, "acgtn", rng.Intn(200), concrete(rng, p))
		for _, circular := range []bool{false, true} {
			want := p.c.search(nil, seq, -1, -1, circular)
			got := s.appendAll(nil, seq, -1, -1, circular)
			if len(got) != len(want) {
				t.Fatalf("pattern %q: %v against %v", tx, got, want)
			}
			for i := range got {
				if got[i] != (Occurrence{0, 1, want[i][0], want[i][1], want[i][2]}) {
					t.Fatalf("pattern %q: %v against %v", tx, got, want)
				}
			}
		}
	}
}

// TestPatternSetNoLeak puts exact cores side by side in one word. What the end
// of a core pushes into the first bit of the next one must not start a window.
func TestPatternSetNoLeak(t *testing.T) {
	ps := compileAll(t, []string{"ACGT", "ACGT", "TTTT"}, []int{0, 0, 0})
	s, _ := MakePatternSet(ps, false)
	if s.Words() != 1 {
		t.Fatalf("%d words", s.Words())
	}
	for _, c := range []struct {
		seq  string
		want []Occurrence
	}{
		{"ACGTTTT", []Occurrence{{0, 1, 0, 4, 0}, {1, 1, 0, 4, 0}, {2, 1, 3, 7, 0}}},
		{"ACGACGT", []Occurrence{{0, 1, 3, 7, 0}, {1, 1, 3, 7, 0}}},
		{"ACGTACGT", []Occurrence{{0, 1, 0, 4, 0}, {1, 1, 0, 4, 0}, {0, 1, 4, 8, 0}, {1, 1, 4, 8, 0}}},
		{"ACGTTT", nil},
	} {
		got := s.appendAll(nil, []byte(c.seq), 0, -1, false)
		if c.seq == "ACGTTT" { // TTTT is not there, ACGT is
			c.want = []Occurrence{{0, 1, 0, 4, 0}, {1, 1, 0, 4, 0}}
		}
		if !sameOccurrences(got, c.want) {
			t.Fatalf("%s: got %v want %v", c.seq, got, c.want)
		}
	}
}

// TestPatternSetOrder: the occurrences come by start, whatever the order of the
// ends, then by pattern, then by strand.
func TestPatternSetOrder(t *testing.T) {
	ps := compileAll(t, []string{"ACGTACGTAC", "GT"}, []int{0, 0})
	s, _ := MakePatternSet(ps, false)
	got := s.FindAll(newSeq([]byte("ACGTACGTAC")), 0, -1)
	// the long one starts first but ends last
	want := []Occurrence{{0, 1, 0, 10, 0}, {1, 1, 2, 4, 0}, {1, 1, 6, 8, 0}}
	if !sameOccurrences(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}

	// a palindrome is found once per strand, the direct one first
	ps = compileAll(t, []string{"ACGT"}, []int{0})
	s, _ = MakePatternSet(ps, true)
	got = s.FindAll(newSeq([]byte("TACGTA")), 0, -1)
	want = []Occurrence{{0, 1, 1, 5, 0}, {0, -1, 1, 5, 0}}
	if !sameOccurrences(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if s.FindAll(newSeq(nil), 0, -1) != nil {
		t.Fatal("an empty sequence has no occurrence")
	}
}

// TestPatternSetVector: sets whose words all search 2 levels, so that groups
// of words run on SIMD vectors when the program is built for it (GOEXPERIMENT=simd);
// they must give the occurrences of the scalar kernels. Short, very degenerate
// patterns make the ends frequent, long ones fill a word each.
func TestPatternSetVector(t *testing.T) {
	testVector = true
	defer func() { testVector = false }()
	rng := rand.New(rand.NewSource(24))
	groups := 0
	for iter := 0; iter < iterations(1500); iter++ {
		var texts []string
		var errs []int
		for k, n := 0, 2+rng.Intn(20); k < n; k++ {
			l := []int{4, 8, 20, 33, 50, 63}[rng.Intn(6)]
			texts = append(texts, exactPattern(rng, l))
			errs = append(errs, 2)
		}
		ps := compileAll(t, texts, errs)
		checkSet(t, rng, ps, rng.Intn(2) == 0, "vector")
		s, _ := makePatternSet(ps, true, true)
		groups += len(s.groups)
	}
	if vectorLanes() > 1 && groups == 0 {
		t.Error("no group of words was built")
	}
	t.Logf("%d lanes, %d groups built", vectorLanes(), groups)
}

// TestPatternSetBatches: sets of many patterns of the same number of errors, so
// that the words are advanced together by batches of 2 and 3, for each level of
// 0 to 10 (the kernels written out and the generic one) and each batch size.
func TestPatternSetBatches(t *testing.T) {
	rng := rand.New(rand.NewSource(25))
	defer func(old int) { forcedBatch = old }(forcedBatch)
	batched := 0
	for iter := 0; iter < iterations(1200); iter++ {
		forcedBatch = 1 + iter%3
		e := iter / 3 % 11
		var texts []string
		var errs []int
		for k, n := 0, 3+rng.Intn(14); k < n; k++ {
			// wildcards at the ends: the words of a batch do not share their windows
			fl := func() string { return strings.Repeat("N", rng.Intn(7)) }
			texts = append(texts, fl()+exactPattern(rng, []int{9, 20, 33, 50, 62}[rng.Intn(5)])+fl())
			errs = append(errs, e)
		}
		ps := compileAll(t, texts, errs)
		checkSet(t, rng, ps, rng.Intn(2) == 0, "batches")
		s, _ := MakePatternSet(ps, true)
		for _, b := range s.batches {
			if len(b) > 1 {
				batched++
			}
		}
	}
	if batched < 500 {
		t.Errorf("only %d batches of several words", batched)
	}
}

// TestPatternSetRefusesIndels: a pattern that allows indels would be searched by the
// recurrence of the substitutions, with wrong results: it is refused. A pattern
// compiled with allowsIndel but no error allowed is a strict one and is accepted.
func TestPatternSetRefusesIndels(t *testing.T) {
	withIndels, err := MakeApatPattern("ACGTACGTAC", 2, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := MakePatternSet([]ApatPattern{withIndels}, false); err == nil {
		t.Error("a pattern with indels must be refused")
	}
	strict, _ := MakeApatPattern("ACGTACGTAC", 0, true)
	if _, err := MakePatternSet([]ApatPattern{strict}, true); err != nil {
		t.Errorf("a strict pattern must be accepted: %v", err)
	}
}
