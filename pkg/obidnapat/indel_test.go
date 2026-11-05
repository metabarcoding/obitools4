package obidnapat

import (
	"math/rand"
	"sort"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
)

// oracleIndel is the definition of the search with indels, by dynamic
// programming (Sellers): d[i][j] is the fewest errors to align the first i
// positions of the pattern with a piece of the text that ends at its j-th base,
// the piece starting anywhere. A substitution, a deletion of a position of the
// pattern or an insertion of a base after a position costs 1, and is forbidden at
// a position that does not allow errors (#). The first column holds the leading
// deletions, which are not subject to #, as in obiapat. It shares nothing with the
// bit-parallel kernel but the parsed positions.
func oracleIndel(p ApatPattern, seq []byte, begin, length int, circular bool) [][3]int {
	const inf = 1 << 20
	n, ps := len(seq), p.c.positions
	m := len(ps)
	if begin < 0 {
		begin = 0
	}
	if begin >= n || circular && m > n {
		return nil
	}
	limit := n
	if length >= 0 && begin+length < n {
		limit = begin + length
	}
	to := limit
	if circular {
		to = limit - 1 + m
	}
	cost := func(i int) int { // of an error at the position i
		if ps[i].oblig {
			return inf
		}
		return 1
	}

	var res [][3]int
	prev := make([]int, m+1) // column j-1
	cur := make([]int, m+1)
	for i := range prev {
		prev[i] = i // leading deletions
	}
	for v := begin; v < to; v++ {
		b := seq[v%n]
		cur[0] = 0
		for i := 1; i <= m; i++ {
			pos := ps[i-1]
			sub := prev[i-1]
			if !(pos.wildcard || letterSet[b]&pos.set != 0) {
				sub += cost(i - 1)
			}
			cur[i] = min(sub, prev[i]+cost(i-1), cur[i-1]+cost(i-1))
		}
		if e := cur[m]; e <= p.c.maxErr {
			if s := v - m + 1; s >= begin {
				res = append(res, [3]int{s, s + m, e})
			}
		}
		prev, cur = cur, prev
	}
	return res
}

// TestIndelAgainstOracle checks the search with indels against the dynamic
// programming, linear and circular, with regions, ambiguity codes, # and wildcards.
func TestIndelAgainstOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(31))
	found, empty := 0, 0
	for iter := 0; iter < iterations(6000); iter++ {
		pat := randomPattern(rng, 1+rng.Intn(30), true)
		e := 1 + rng.Intn(10)
		p, err := MakeApatPattern(pat, e, true)
		if err != nil {
			continue // more than 63 positions with the wildcards at the ends
		}
		alphabet := []string{"acgt", "acgtn", "acgtrykmn-", "ac"}[rng.Intn(4)]
		seq := randomSequence(rng, alphabet, rng.Intn(200), concrete(rng, p))
		begin, length := -1+rng.Intn(60), -1
		if rng.Intn(2) == 0 {
			length = rng.Intn(150)
		}
		for _, circular := range []bool{false, true} {
			want := oracleIndel(p, seq, begin, length, circular)
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
	if found < 3000 || empty < 500 {
		t.Errorf("the draws are not varied enough: %d matches, %d empty results", found, empty)
	}
}

// TestIndelAgainstObiapat compares the search with indels with obiapat's
// ManberIndel on sequences over acgt, brought to the conventions of this package
// (the nominal window must lie in the region; a circular sequence reports each
// start once). The two must agree window by window, nominal starts included.
func TestIndelAgainstObiapat(t *testing.T) {
	rng := rand.New(rand.NewSource(32))
	found := 0
	for iter := 0; iter < iterations(5000); iter++ {
		pat := randomPattern(rng, 1+rng.Intn(30), true)
		e := 1 + rng.Intn(8)
		p, err := MakeApatPattern(pat, e, true)
		if err != nil {
			continue
		}
		legacy, err := obiapat.MakeApatPattern(pat, e, true)
		if err != nil {
			t.Fatalf("obiapat refuses %q: %v", pat, err)
		}
		seq := randomSequence(rng, "acgt", rng.Intn(250), concrete(rng, p))
		bs := newSeq(seq)

		for _, circular := range []bool{false, true} {
			if circular && p.Len() > len(seq) {
				continue
			}
			begin, length := 0, -1
			if !circular {
				begin = rng.Intn(40)
				if rng.Intn(2) == 0 {
					length = rng.Intn(200)
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
			for _, m := range legacy.FindAllIndex(ls, begin, length) {
				switch {
				case circular && m[0] >= 0 && m[0] < n:
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

// TestIndelBestMatchAgainstObiapat compares BestMatch, with its realignment by
// obialign.LocatePattern, with obiapat's, on sequences over acgt. The cases where
// obiapat lists windows outside the region (before it, or past its end), which
// this package leaves out, are skipped: the two start from different lists.
func TestIndelBestMatchAgainstObiapat(t *testing.T) {
	rng := rand.New(rand.NewSource(33))
	compared, refined, unmatched := 0, 0, 0
	for iter := 0; iter < iterations(5000); iter++ {
		pat := randomPattern(rng, 1+rng.Intn(30), false)
		e := 1 + rng.Intn(5)
		p, err := MakeApatPattern(pat, e, true)
		if err != nil {
			continue
		}
		legacy, err := obiapat.MakeApatPattern(pat, e, true)
		if err != nil {
			t.Fatalf("obiapat refuses %q: %v", pat, err)
		}
		seq := randomSequence(rng, "acgt", 1+rng.Intn(120), concrete(rng, p))
		bs := newSeq(seq)
		begin, length := rng.Intn(20), -1
		if rng.Intn(2) == 0 {
			length = rng.Intn(100)
		}
		ls, err := obiapat.MakeApatSequence(newSeq(append([]byte(nil), seq...)), false)
		if err != nil {
			t.Fatal(err)
		}
		n := len(seq)
		limit := n
		if length >= 0 && begin+length < n {
			limit = begin + length
		}
		clean := true
		for _, m := range legacy.FindAllIndex(ls, begin, length) {
			if m[0] < begin || m[1] > limit {
				clean = false
			}
		}
		if !clean {
			continue
		}
		ws, we, wn, wm := legacy.BestMatch(ls, begin, length)
		gs, ge, gn, gm := p.BestMatch(bs, begin, length)
		if gm != wm || gm && (gs != ws || ge != we || gn != wn) {
			t.Fatalf("pattern %q e=%d begin=%d length=%d seq=%q\n new     %d %d %d %v\n obiapat %d %d %d %v",
				pat, e, begin, length, seq, gs, ge, gn, gm, ws, we, wn, wm)
		}
		compared++
		if gm && gn > 0 {
			refined++
		}
		if !gm {
			unmatched++
		}
	}
	if compared < 1000 || refined < 200 {
		t.Errorf("the draws are not varied enough: %d compared, %d realigned, %d without match", compared, refined, unmatched)
	}
	t.Logf("%d compared, %d realigned, %d without match", compared, refined, unmatched)
}
