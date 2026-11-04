package obidnapat

import (
	"strings"
	"testing"
)

func TestLetterSets(t *testing.T) {
	want := map[byte]uint8{
		'a': baseA, 'c': baseC, 'g': baseG, 't': baseT, 'u': baseT,
		'r': baseA | baseG, 'y': baseC | baseT, 's': baseC | baseG, 'w': baseA | baseT,
		'k': baseG | baseT, 'm': baseA | baseC,
		'b': baseC | baseG | baseT, 'd': baseA | baseG | baseT,
		'h': baseA | baseC | baseT, 'v': baseA | baseC | baseG, 'n': anyBase,
	}
	for c := 0; c < 256; c++ {
		lower := byte(c)
		if c >= 'A' && c <= 'Z' {
			lower = byte(c) + 0x20 // either case reads the same
		}
		if got := letterSet[c]; got != want[lower] {
			t.Errorf("letterSet[%q] = %04b, want %04b", rune(c), got, want[lower])
		}
	}
}

func TestParse(t *testing.T) {
	good := map[string]int{
		"ACGT": 4, "acgt": 4, "ACGU": 4, "ARYN": 4, "[AG]CT": 3, "![AG]CT": 3, "!A": 1,
		"A#CG": 3, "[AG]#T": 2, "!!A": 1, "ANX": 3, "NNNACGTNNN": 10,
	}
	for pat, n := range good {
		ps, err := parse(pat)
		if err != nil || len(ps) != n {
			t.Errorf("parse(%q) = %d positions, %v; want %d", pat, len(ps), err, n)
		}
	}
	for _, pat := range []string{"", "#A", "A##", "[", "[]", "[A", "[A!]", "[A#]", "!", "A!", "AE", "A-C", "A]", "[[A]]"} {
		if _, err := parse(pat); err == nil {
			t.Errorf("parse(%q) must fail", pat)
		}
	}
}

func TestPositions(t *testing.T) {
	ps, _ := parse("![AG]#N[ACGT]!N")
	type exp struct {
		set       uint8
		wild, obl bool
	}
	want := []exp{{baseC | baseT, false, true}, {anyBase, true, false}, {anyBase, true, false}, {0, false, false}}
	for i, w := range want {
		if ps[i].set != w.set || ps[i].wildcard != w.wild || ps[i].oblig != w.obl {
			t.Errorf("position %d (%q) = %+v, want %+v", i, ps[i].text, ps[i], w)
		}
	}
}

func TestLengthLimit(t *testing.T) {
	for _, tc := range []struct {
		pat string
		ok  bool
	}{
		{strings.Repeat("A", 63), true},
		{strings.Repeat("A", 64), false},
		{strings.Repeat("N", 40) + strings.Repeat("A", 63) + strings.Repeat("N", 40), true}, // the limit is on the core
		{strings.Repeat("N", 40) + strings.Repeat("A", 64) + strings.Repeat("N", 40), false},
		{strings.Repeat("N", 100), true},
	} {
		_, err := MakeApatPattern(tc.pat, 2, false)
		if (err == nil) != tc.ok {
			t.Errorf("pattern of %d symbols: error %v, want ok=%v", len(tc.pat), err, tc.ok)
		}
	}
	if _, err := MakeApatPattern("ACGT", 1, true); err == nil {
		t.Error("indels must be refused for now")
	}
}

func TestReverseComplement(t *testing.T) {
	for pat, want := range map[string]string{
		"AACG":         "CGTT",
		"[AG]CTN":      "NAG[TC]",
		"!AC#T":        "AG#!T",
		"NNACGTNN":     "NNACGTNN",
		"RYKMBDHVSWUX": "XAWSBDHVKMRY",
	} {
		p, err := MakeApatPattern(pat, 0, false)
		if err != nil {
			t.Fatal(err)
		}
		rc, err := p.ReverseComplement()
		if err != nil || rc.String() != want {
			t.Errorf("rc(%q) = %q, %v; want %q", pat, rc.String(), err, want)
		}
		back, _ := rc.ReverseComplement()
		if pat != "RYKMBDHVSWUX" && back.String() != strings.ToUpper(pat) { // U comes back as T
			t.Errorf("rc(rc(%q)) = %q", pat, back.String())
		}
	}
}

func TestSemantics(t *testing.T) {
	find := func(pat string, e int, seq string) [][3]int {
		p, err := MakeApatPattern(pat, e, false)
		if err != nil {
			t.Fatal(err)
		}
		return p.FindAllIndex(newSeq([]byte(seq)), 0, -1)
	}
	check := func(got, want [][3]int, what string) {
		t.Helper()
		if !sameMatches(got, want) {
			t.Errorf("%s: got %v, want %v", what, got, want)
		}
	}
	check(find("ACGTAC", 0, "ttacgtactt"), [][3]int{{2, 8, 0}}, "exact")
	check(find("ACNTAC", 0, "ttacgtactt"), [][3]int{{2, 8, 0}}, "pattern N")
	check(find("ACRTAC", 0, "ttacgtactt"), [][3]int{{2, 8, 0}}, "pattern R on g")
	check(find("ACRTAC", 0, "ttactactt"), nil, "pattern R on c")
	// the ambiguity codes of the sequence are no longer errors
	check(find("ACGTAC", 0, "ttacntactt"), [][3]int{{2, 8, 0}}, "sequence n")
	check(find("ACGTAC", 0, "ttacrtactt"), [][3]int{{2, 8, 0}}, "sequence r on pattern g")
	check(find("ACCTAC", 0, "ttacrtactt"), nil, "sequence r on pattern c: A/G never gives C")
	check(find("ACGTAC", 0, "ttacguactt"), [][3]int{{2, 8, 0}}, "sequence u is t")
	check(find("ACGTAC", 0, "ttac-tactt"), nil, "a dash is no base")
	check(find("ACGTAC", 1, "ttac-tactt"), [][3]int{{2, 8, 1}}, "a dash is an error")
	// # forbids an error
	check(find("ACG#TAC", 1, "ttacctactt"), nil, "error at a # position")
	check(find("ACGT#AC", 1, "ttacctactt"), [][3]int{{2, 8, 1}}, "error away from #")
	// wildcards around the core
	check(find("NNACGTNN", 0, "ttacgtactt"), [][3]int{{0, 8, 0}}, "flanks fit")
	check(find("NNACGTNN", 0, "acgtacttt"), nil, "no room for the left flank")
	check(find("NNACGTNN", 0, "tttacgt"), nil, "no room for the right flank")
	check(find("NNACGTNN", 0, "ttttacgtacttt"), [][3]int{{2, 10, 0}}, "flanks")
	check(find("NNACGTNN", 0, "ttacgtac"), [][3]int{{0, 8, 0}}, "flanks, exact fit")
	// the strands
	check(find("AACCGG", 1, "ttaaccgctt"), [][3]int{{2, 8, 1}}, "one error")
}
