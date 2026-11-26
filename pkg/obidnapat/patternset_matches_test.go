package obidnapat

import (
	"math/rand"
	"slices"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// randomMixedSet draws patterns with and without indels, of any length and number
// of errors (0 to 10), and compiles them.
func randomMixedSet(t *testing.T, rng *rand.Rand, maxLen int) []ApatPattern {
	var ps []ApatPattern
	for i, n := 0, 1+rng.Intn(30); i < n; i++ {
		p, err := MakeApatPattern(randomPattern(rng, 1+rng.Intn(maxLen), true), rng.Intn(11), rng.Intn(3) > 0)
		if err != nil {
			continue
		}
		ps = append(ps, p)
	}
	return ps
}

func randomSeqFor(rng *rand.Rand, ps []ApatPattern) []byte {
	alphabet := []string{"acgt", "acgt", "acgtn", "acgtrykmn-", "ac"}[rng.Intn(5)]
	seq := randomSequence(rng, alphabet, rng.Intn(300), "")
	for k := 0; k < 5 && len(seq) > 0; k++ {
		c := []byte(mutate(rng, concrete(rng, ps[rng.Intn(len(ps))])))
		if len(c) < len(seq) {
			copy(seq[rng.Intn(len(seq)-len(c)+1):], c)
		}
	}
	return seq
}

// TestPatternSetAllMatches compares AppendAllMatches, pattern by pattern and
// strand by strand, with the AllMatches of the pattern (and, for the patterns
// with indels, with expectedAllMatches, which does not share the code).
func TestPatternSetAllMatches(t *testing.T) {
	rng := rand.New(rand.NewSource(77))
	found, withIndel, multi := 0, 0, 0
	for iter := 0; iter < iterations(3000); iter++ {
		ps := randomMixedSet(t, rng, 30)
		if len(ps) == 0 {
			continue
		}
		both := rng.Intn(4) > 0
		s, err := MakePatternSet(ps, both)
		if err != nil {
			t.Fatal(err)
		}
		if len(s.iwords) > 0 && len(s.words) > 0 {
			multi++
		}
		seq := randomSeqFor(rng, ps)
		begin, length := -1+rng.Intn(60), -1
		if rng.Intn(2) == 0 {
			length = rng.Intn(250)
		}
		bs := obiseq.NewBioSequence("s", seq, "")
		got := s.AppendAllMatches(nil, bs, begin, length)
		if !slices.IsSortedFunc(got, compareOccurrences) {
			t.Fatalf("not sorted: %v", got)
		}
		total := 0
		for i, p := range ps {
			for _, strand := range []int8{1, -1} {
				q := p
				if strand < 0 {
					if !both {
						continue
					}
					if q, err = p.ReverseComplement(); err != nil {
						t.Fatal(err)
					}
				}
				want := q.AllMatches(bs, begin, length)
				if q.c.indel {
					withIndel++
					if ex := expectedAllMatches(q, seq, begin, length); !sameMatches(nilIfEmpty(ex), want) {
						t.Fatalf("AllMatches differs from the oracle: pattern %q seq %q begin %d length %d\n oracle %v\n all    %v", q.String(), seq, begin, length, ex, want)
					}
				}
				var mine [][3]int
				for _, o := range got {
					if o.Pattern == i && o.Strand == strand {
						mine = append(mine, [3]int{o.Start, o.End, o.Errors})
					}
				}
				if !sameMatches(mine, want) {
					t.Fatalf("pattern %d %q (indel %v, %d errors) strand %d, seq %q begin %d length %d, %d words + %d with indels\n set %v\n one %v",
						i, q.String(), q.c.indel, q.c.maxErr, strand, seq, begin, length, len(s.words), len(s.iwords), mine, want)
				}
				total += len(want)
			}
		}
		if total != len(got) {
			t.Fatalf("%d occurrences, %d expected", len(got), total)
		}
		found += total
	}
	t.Logf("%d windows, %d pattern-strands with indels, %d sets with both kinds of words", found, withIndel, multi)
	if found < 5000 || multi < 200 {
		t.Errorf("not varied enough")
	}
}

// TestPatternSetIndelRaw: AppendAll and AppendAllCircular give, for a pattern with
// indels, the matches of AppendAllIndex (AppendAllIndexCircular).
func TestPatternSetIndelRaw(t *testing.T) {
	rng := rand.New(rand.NewSource(78))
	for iter := 0; iter < iterations(1500); iter++ {
		ps := randomMixedSet(t, rng, 30)
		if len(ps) == 0 {
			continue
		}
		s, err := MakePatternSet(ps, true)
		if err != nil {
			t.Fatal(err)
		}
		seq := randomSeqFor(rng, ps)
		bs := obiseq.NewBioSequence("s", seq, "")
		begin, length := -1+rng.Intn(60), -1
		if rng.Intn(2) == 0 {
			length = rng.Intn(250)
		}
		for _, circular := range []bool{false, true} {
			var got []Occurrence
			if circular {
				got = s.AppendAllCircular(nil, bs, begin, length)
			} else {
				got = s.AppendAll(nil, bs, begin, length)
			}
			total := 0
			for i, p := range ps {
				for _, strand := range []int8{1, -1} {
					q := p
					if strand < 0 {
						q, _ = p.ReverseComplement()
					}
					var want [][3]int
					if circular {
						want = q.AppendAllIndexCircular(nil, bs, begin, length)
					} else {
						want = q.AppendAllIndex(nil, bs, begin, length)
					}
					var mine [][3]int
					for _, o := range got {
						if o.Pattern == i && o.Strand == strand {
							mine = append(mine, [3]int{o.Start, o.End, o.Errors})
						}
					}
					if !sameMatches(mine, nilIfEmpty(want)) && !(len(mine) == 0 && len(want) == 0) {
						t.Fatalf("pattern %q strand %d circular %v seq %q begin %d length %d\n set %v\n one %v", q.String(), strand, circular, seq, begin, length, mine, want)
					}
					total += len(want)
				}
			}
			if total != len(got) {
				t.Fatalf("%d occurrences, %d expected", len(got), total)
			}
		}
	}
}

// TestPatternSetIndelBatches: many patterns with the same number of errors, so
// that the words with indels run together (2 or 3 in one loop), against the
// search of each pattern alone; every level written out (1 to 8) and the generic ones above.
func TestPatternSetIndelBatches(t *testing.T) {
	defer func() { forcedBatch = 0 }()
	rng := rand.New(rand.NewSource(79))
	for level := 1; level <= 10; level++ {
		for iter := 0; iter < iterations(600); iter++ {
			forcedBatch = iter % 4 // 0: the usual size, then 1, 2 and 3 words in a loop
			var ps []ApatPattern
			for len(ps) < 4+rng.Intn(14) {
				p, err := MakeApatPattern(randomPattern(rng, 8+rng.Intn(25), true), level, true)
				if err == nil {
					ps = append(ps, p)
				}
			}
			s, err := MakePatternSet(ps, true)
			if err != nil {
				t.Fatal(err)
			}
			seq := randomSeqFor(rng, ps)
			bs := obiseq.NewBioSequence("s", seq, "")
			begin, length := -1+rng.Intn(40), -1
			if rng.Intn(3) == 0 {
				length = rng.Intn(250)
			}
			got := s.AppendAllMatches(nil, bs, begin, length)
			total := 0
			for i, p := range ps {
				for _, strand := range []int8{1, -1} {
					q := p
					if strand < 0 {
						q, _ = p.ReverseComplement()
					}
					want := q.AllMatches(bs, begin, length)
					var mine [][3]int
					for _, o := range got {
						if o.Pattern == i && o.Strand == strand {
							mine = append(mine, [3]int{o.Start, o.End, o.Errors})
						}
					}
					if !sameMatches(mine, want) {
						t.Fatalf("level %d, pattern %q strand %d, seq %q (%d words)\n set %v\n one %v", level, q.String(), strand, seq, len(s.iwords), mine, want)
					}
					total += len(want)
				}
			}
			if total != len(got) {
				t.Fatalf("%d occurrences, %d expected", len(got), total)
			}
		}
	}
}
