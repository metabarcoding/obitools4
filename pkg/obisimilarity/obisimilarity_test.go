package obisimilarity

import (
	"math/rand"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
)

func randomSeq(rng *rand.Rand, n int, alpha string) []byte {
	s := make([]byte, n)
	for i := range s {
		s[i] = alpha[rng.Intn(len(alpha))]
	}
	return s
}

// commonWordsRef is the definition of commonWords, spelled out.
func commonWordsRef(t1, t2 *Table4) int {
	total := 0
	for i := range t1.Count {
		total += int(min(t1.Count[i], t2.Count[i]))
	}
	return total + min(t1.Overflow, t2.Overflow)
}

func TestCommonWords(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	for it := 0; it < 3000; it++ {
		alpha := []string{"acgt", "ac", "a"}[it%3]
		t1 := NewTable4(randomSeq(rng, rng.Intn(1200), alpha))
		t2 := NewTable4(randomSeq(rng, rng.Intn(1200), alpha))
		if got, want := commonWords(t1, t2), commonWordsRef(t1, t2); got != want {
			t.Fatalf("commonWords = %d, want %d", got, want)
		}
	}
}

// TestNewTable4 checks the table against the 4-mers counted the long way,
// including the saturation of the counters.
func TestNewTable4(t *testing.T) {
	code := map[byte]int{'a': 0, 'c': 1, 't': 2, 'g': 3}
	rng := rand.New(rand.NewSource(4))
	for it := 0; it < 500; it++ {
		alpha := []string{"acgt", "a", "ac"}[it%3]
		s := randomSeq(rng, rng.Intn(900), alpha)
		var counts [256]int
		for i := 0; i+4 <= len(s); i++ {
			w := 0
			for _, c := range s[i : i+4] {
				w = w<<2 | code[c]
			}
			counts[w]++
		}
		tab := NewTable4(s)
		over := 0
		for w, c := range counts {
			if c > 255 {
				over += c - 255
				c = 255
			}
			if int(tab.Count[w]) != c {
				t.Fatalf("word %d: got %d, want %d", w, tab.Count[w], c)
			}
		}
		if tab.Overflow != over {
			t.Fatalf("overflow = %d, want %d", tab.Overflow, over)
		}
	}
}

// TestBlockCommon checks the sums of a block of transposed tables against
// commonWords.
func TestBlockCommon(t *testing.T) {
	width := simdBlockWidth()
	if width == 0 {
		t.Skip("blocks of tables are used with GOEXPERIMENT=simd only")
	}
	rng := rand.New(rand.NewSource(8))
	for it := 0; it < 200; it++ {
		alpha := []string{"acgt", "ac", "a"}[it%3]
		block := newTableBlock(width)
		tabs := make([]*Table4, width)
		for j := range tabs {
			tabs[j] = NewTable4(randomSeq(rng, rng.Intn(600), alpha))
			block.set(width, j, tabs[j])
		}
		q := NewTable4(randomSeq(rng, rng.Intn(260), alpha))
		words := SparseWords(q)
		if words == nil {
			continue
		}
		out := make([]uint8, 64)
		block.common(width, words, out)
		for j, tab := range tabs {
			want := commonWordsRef(q, tab) - min(q.Overflow, tab.Overflow)
			if int(out[j]) != want {
				t.Fatalf("lane %d: %d, want %d", j, out[j], want)
			}
		}
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		in        string
		accept    bool
		kept      bool
		ambiguous bool
	}{
		{"acgt", false, true, false}, {"acgt", true, true, false},
		{"acgtn", false, false, false}, {"acgtn", true, true, true},
		{"acgtx", true, false, false}, {"nacgt", false, false, false},
		{"acgt-", true, false, false}, {"acgu", true, true, true},
		{"ACGT", true, false, false}, {"", true, false, false},
		{"rykmswbdhvn", true, true, true}, {"rykmswbdhvn", false, false, false},
	}
	for _, c := range cases {
		kept, amb := Classify([]byte(c.in), c.accept)
		if kept != c.kept || amb != c.ambiguous {
			t.Errorf("Classify(%q, %v) = (%v, %v), want kept=%v ambiguous=%v", c.in, c.accept, kept, amb, c.kept, c.ambiguous)
		}
	}
}

func randomEntries(rng *rand.Rand, n int) []Entry {
	return BuildEntries(n, func(int) ([]byte, int, bool) {
		return randomSeq(rng, 8+rng.Intn(80), "acgt"), 1 + rng.Intn(20), false
	}, 2)
}

// TestLengthTable checks that the values precomputed by length give the
// scores of the direct computation, in every mode.
func TestLengthTable(t *testing.T) {
	rng := rand.New(rand.NewSource(33))
	// sequences made of the same few roots, so that many pairs are similar
	roots := [][]byte{randomSeq(rng, 70, "acgt"), randomSeq(rng, 55, "acgt")}
	entries := BuildEntries(120, func(i int) ([]byte, int, bool) {
		var s []byte
		for _, c := range roots[i%2] {
			switch r := rng.Intn(30); {
			case r == 0:
				s = append(s, "acgt"[rng.Intn(4)])
			case r == 1:
			default:
				s = append(s, c)
			}
		}
		if i < 20 { // very short sequences skip the filter
			s = s[:8+rng.Intn(10)]
		}
		return s, 1 + rng.Intn(20), false
	}, 2)

	var buf obialign.BandedLCSBuffer
	for _, ref := range []Reference{RefAlignment, RefLongest, RefShortest} {
		for _, p := range []Params{
			{Threshold: 0.9, Normalized: true},
			{Threshold: 0.8, Normalized: true, Distance: true},
			{Threshold: 30},
			{Threshold: 6, Distance: true},
			{Normalized: true}, // no threshold
		} {
			p.Reference, p.Ratio = ref, 3
			table, err := NewScoring(p, entries)
			if err != nil {
				t.Fatal(err)
			}
			direct := *table
			direct.byLength = nil // forces the direct computation
			if (ref == RefShortest) != (table.byLength == nil) {
				t.Fatalf("ref %v: table presence is wrong", ref)
			}
			for i := range entries {
				for j := range entries {
					a, aok := direct.PairScore(&entries[i], &entries[j], -1, &buf)
					b, bok := table.PairScore(&entries[i], &entries[j], -1, &buf)
					if a != b || aok != bok {
						t.Fatalf("ref %v params %+v pair %d,%d: (%v,%v) with the table, (%v,%v) directly", ref, p, i, j, b, bok, a, aok)
					}
				}
			}
		}
	}
}

// TestIndexCommon checks that the filter computed by blocks gives the same
// scores as the direct computation, for every pair, whatever the order the
// members are read in.
func TestIndexCommon(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	entries := randomEntries(rng, 150)
	sc, err := NewScoring(Params{Threshold: 0.6, Normalized: true, Reference: RefAlignment, Ratio: 100}, entries)
	if err != nil {
		t.Fatal(err)
	}
	index := NewIndex()
	for i := range entries {
		index.Add(entries[i], i)
	}
	if index.Len() != len(entries) {
		t.Fatalf("Len = %d", index.Len())
	}

	var buf obialign.BandedLCSBuffer
	var commons Commons
	for i := range entries {
		q := &entries[i]
		for _, length := range index.Lengths() {
			b := index.Bucket(length)
			commons.Reset()
			for k := len(b.Members) - 1; k >= 0; k-- { // backwards: blocks are not read in order
				m := &b.Members[k]
				got, gok := sc.PairScore(q, &m.Entry, b.Common(k, q.Sparse, &commons), &buf)
				want, wok := sc.PairScore(q, &entries[m.ID], -1, &buf)
				if got != want || gok != wok {
					t.Fatalf("pair %d,%d: (%v,%v) with the blocks, (%v,%v) directly", i, m.ID, got, gok, want, wok)
				}
			}
		}
	}
}

func TestNewScoringChecks(t *testing.T) {
	for _, p := range []Params{{Threshold: -1}, {Threshold: 1.5, Normalized: true}} {
		if _, err := NewScoring(p, nil); err == nil {
			t.Errorf("%+v must be rejected", p)
		}
	}
	if _, err := NewScoring(Params{Threshold: 12}, nil); err != nil {
		t.Errorf("a raw threshold above 1 is valid: %v", err)
	}
}
