package obidnapat

import (
	"strings"
	"testing"
)

// A window must hold more real bases (not N) than errors.
func TestNoiseRule(t *testing.T) {
	for _, indel := range []bool{false, true} {
		p, err := MakeApatPattern("ACGTACGT", 2, indel)
		if err != nil {
			t.Fatal(err)
		}
		// 3 real bases and 5 N: matched with 0 errors, but not on real bases
		// alone beyond the errors... 3 > 0 keeps it; 2 real bases with 2 errors
		// does not.
		seq := []byte(strings.Repeat("n", 5000) + "acgtacgt" + strings.Repeat("n", 5000))
		for _, w := range p.FindAllIndex(newSeq(seq), 0, -1) {
			if r := naiveRealBases(seq, w[0], w[1], false); r <= w[2] {
				t.Errorf("indel=%v: window %v holds %d real bases for %d errors", indel, w, r, w[2])
			}
		}
		// a window of N only is never reported
		if got := p.FindAllIndex(newSeq([]byte(strings.Repeat("n", 100))), 0, -1); got != nil {
			t.Errorf("indel=%v: windows of N: %v", indel, got)
		}
	}
}

// A pattern is not matched by its errors alone, even without N.
func TestNoiseRuleShortPattern(t *testing.T) {
	p, _ := MakeApatPattern("A", 1, false)
	got := p.FindAllIndex(newSeq([]byte("acgt")), 0, -1)
	if len(got) != 1 || got[0] != [3]int{0, 1, 0} {
		t.Errorf("got %v, want only the exact match", got)
	}
}

// The same occurrences from a pattern alone and from a PatternSet, on a
// sequence with runs of N.
func TestNoiseRulePatternSet(t *testing.T) {
	pats := []string{"ACGTACGTAC", "GGTTAACC", "TTTTT", "ACGT"}
	var ps []ApatPattern
	for _, s := range pats {
		p, err := MakeApatPattern(s, 3, false)
		if err != nil {
			t.Fatal(err)
		}
		ps = append(ps, p)
	}
	set, err := MakePatternSet(ps, true)
	if err != nil {
		t.Fatal(err)
	}
	seq := []byte("nnnnnnnnnnacgtacgtacnnnnnnnnnnnnnggttaaccnnnnnnnnttttnnnnnnacgtnnn")
	for _, circ := range []bool{false, true} {
		var want [][3]int
		got := set.AppendAll(nil, newSeq(seq), 0, -1)
		if circ {
			got = set.AppendAllCircular(nil, newSeq(seq), 0, -1)
		}
		for _, o := range got {
			if r := naiveRealBases(seq, o.Start, o.End, circ); r <= o.Errors {
				t.Errorf("circular=%v: %+v holds %d real bases", circ, o, r)
			}
			if o.Strand == 1 {
				want = append(want, [3]int{o.Pattern, o.Start, o.End})
			}
		}
		for i, p := range ps {
			var single [][3]int
			if circ {
				single = p.FindAllIndexCircular(newSeq(seq), 0, -1)
			} else {
				single = p.FindAllIndex(newSeq(seq), 0, -1)
			}
			var fromSet [][3]int
			for _, o := range got {
				if o.Pattern == i && o.Strand == 1 {
					fromSet = append(fromSet, [3]int{o.Start, o.End, o.Errors})
				}
			}
			if !sameMatches(single, nilIfEmpty(fromSet)) {
				t.Errorf("circular=%v pattern %q: single %v, set %v", circ, pats[i], single, fromSet)
			}
		}
		_ = want
	}
}
