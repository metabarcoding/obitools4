package obidnapat

import (
	"math/rand"
	"sort"
	"strings"
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
	return oracleIndelRaw(p, seq, begin, length, circular, true)
}

// oracleIndelRaw is oracleIndel; without strict, the windows whose nominal start
// falls before the region are kept too.
func oracleIndelRaw(p ApatPattern, seq []byte, begin, length int, circular, strict bool) [][3]int {
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
			if s := v - m + 1; (s >= begin || !strict) && naiveRealBases(seq, s, s+m, circular) > e {
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
		if e >= p.Len() {
			continue // obiapat has no rule on the real bases of a window (see noise.go)
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

// truthStart is the reference for the start of a window: the latest start s >= lo
// such that the bases seq[s..pos] are aligned with the whole pattern for at most e
// errors, by a global dynamic programming for each s (an insertion before the
// first position is not allowed: the first base is then the start). Same costs as
// the search: a substitution, a deletion or an insertion after a position costs 1
// and is forbidden at a position marked #.
func truthStart(p ApatPattern, seq []byte, pos, e, lo int) (start int, ok bool) {
	const inf = 1 << 20
	ps := p.c.positions
	m := len(ps)
	cost := func(i int) int {
		if ps[i].oblig {
			return inf
		}
		return 1
	}
	for s := pos; s >= max(lo, 0); s-- {
		text := seq[s : pos+1]
		g := make([][]int, m+1)
		for i := range g {
			g[i] = make([]int, len(text)+1)
		}
		for j := 1; j <= len(text); j++ {
			g[0][j] = inf
		}
		for i := 1; i <= m; i++ {
			g[i][0] = g[i-1][0] + cost(i-1)
			for j := 1; j <= len(text); j++ {
				sub := g[i-1][j-1]
				if !(ps[i-1].wildcard || letterSet[text[j-1]]&ps[i-1].set != 0) {
					sub += cost(i - 1)
				}
				g[i][j] = min(sub, g[i-1][j]+cost(i-1), g[i][j-1]+cost(i-1))
			}
		}
		if g[m][len(text)] <= e {
			return s, true
		}
	}
	return 0, false
}

// TestBackwardStartAgainstTruth checks the start found by the backward search
// against the dynamic programming, for every window with errors that the forward
// search reports. Without # the two searches must agree on the existence of a start
// too.
func TestBackwardStartAgainstTruth(t *testing.T) {
	rng := rand.New(rand.NewSource(34))
	checked, found := 0, 0
	for iter := 0; iter < iterations(4000); iter++ {
		pat := randomPattern(rng, 1+rng.Intn(25), true)
		withSharp := strings.Contains(pat, "#")
		e := 1 + rng.Intn(6)
		p, err := MakeApatPattern(pat, e, true)
		if err != nil {
			continue
		}
		alphabet := []string{"acgt", "acgt", "acgtn", "acgtrykmn-"}[rng.Intn(4)]
		seq := randomSequence(rng, alphabet, rng.Intn(150), concrete(rng, p))
		begin := rng.Intn(30)
		for _, h := range p.c.search(nil, seq, begin, -1, false) {
			if h[2] == 0 {
				continue
			}
			got, ok := p.c.backwardStart(seq, h[1]-1, h[2], begin)
			want, wok := truthStart(p, seq, h[1]-1, h[2], begin)
			if ok != wok && !withSharp || ok && (!wok || got != want) {
				t.Fatalf("pattern %q e=%d begin=%d seq=%q window %v: backward %d,%v truth %d,%v",
					pat, e, begin, seq, h, got, ok, want, wok)
			}
			checked++
			if ok {
				found++
			}
		}
	}
	if checked < 1500 || found < 1500 {
		t.Errorf("the draws are not varied enough: %d windows, %d starts found", checked, found)
	}
	t.Logf("%d windows with errors, %d starts found", checked, found)
}

// mutate applies one or two random insertions, deletions or substitutions.
func mutate(rng *rand.Rand, s string) string {
	b := []byte(s)
	for k := rng.Intn(3); k > 0 && len(b) > 2; k-- {
		i := rng.Intn(len(b))
		switch rng.Intn(3) {
		case 0:
			b = append(b[:i], b[i+1:]...)
		case 1:
			b = append(b[:i], append([]byte{"acgt"[rng.Intn(4)]}, b[i:]...)...)
		default:
			b[i] = "acgt"[rng.Intn(4)]
		}
	}
	return string(b)
}

// expectedAllMatches is the reference of AllMatches for a pattern with indels on a
// linear sequence: the windows of the dynamic programming, the ones that overlap
// reduced to the best (the rule of FilterBestMatch, written again here), then the
// real start of each window with errors from truthStart.
func expectedAllMatches(p ApatPattern, seq []byte, begin, length int) [][3]int {
	var picked [][3]int
	var best [3]int
	have := false
	for _, w := range oracleIndelRaw(p, seq, begin, length, false, false) {
		switch {
		case !have:
			best, have = w, true
		case w[0]-w[2] < best[1]+best[2]: // overlaps the best one so far
			if w[2] < best[2] {
				best = w
			}
		default:
			picked = append(picked, best)
			best = w
		}
	}
	if have {
		picked = append(picked, best)
	}
	var res [][3]int
	for _, w := range picked {
		if w[2] > 0 {
			if s, ok := truthStart(p, seq, w[1]-1, w[2], begin); ok {
				w[0] = s
			}
		}
		if w[0] >= begin {
			res = append(res, w)
		}
	}
	return res
}

// TestIndelAllMatchesAgainstOracle checks AllMatches (real starts) of a pattern
// with indels against the reference above, with # and wildcards too.
func TestIndelAllMatchesAgainstOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(36))
	found, shifted := 0, 0
	for iter := 0; iter < iterations(5000); iter++ {
		pat := randomPattern(rng, 1+rng.Intn(25), true)
		e := 1 + rng.Intn(5)
		p, err := MakeApatPattern(pat, e, true)
		if err != nil {
			continue
		}
		alphabet := []string{"acgt", "acgt", "acgtn", "acgtrykmn-"}[rng.Intn(4)]
		seq := randomSequence(rng, alphabet, rng.Intn(200), concrete(rng, p))
		if len(seq) > 40 && rng.Intn(2) == 0 { // a second copy of the site, with a mutation
			copy(seq[rng.Intn(len(seq)-30):], mutate(rng, concrete(rng, p)))
		}
		begin, length := rng.Intn(30), -1
		if rng.Intn(3) == 0 {
			length = rng.Intn(150)
		}
		want := expectedAllMatches(p, seq, begin, length)
		got := p.AllMatches(newSeq(seq), begin, length)
		if !sameMatches(got, nilIfEmpty(want)) {
			t.Fatalf("pattern %q e=%d begin=%d length=%d seq=%q\n got  %v\n want %v", pat, e, begin, length, seq, got, want)
		}
		found += len(got)
		for _, w := range got {
			if w[0] != w[1]-p.Len() {
				shifted++
			}
		}
	}
	if found < 2000 || shifted < 300 {
		t.Errorf("the draws are not varied enough: %d windows, %d with a start other than the nominal one", found, shifted)
	}
}

// TestIndelAllMatchesAgainstObiapat compares the ends and the errors of AllMatches
// with obiapat's FilterBestMatch (the grouping of the overlapping windows, whose
// starts are the nominal ones), on sequences over acgt and where none of obiapat's
// windows falls partly outside the region. (obiapat's own AllMatches realigns each
// window with obialign.LocatePattern, which moves its end and changes its errors,
// and drops it when that fails: it is compared on real reads, see realdata_test.go.)
func TestIndelAllMatchesAgainstObiapat(t *testing.T) {
	rng := rand.New(rand.NewSource(37))
	compared := 0
	for iter := 0; iter < iterations(5000); iter++ {
		pat := randomPattern(rng, 1+rng.Intn(25), true)
		e := 1 + rng.Intn(5)
		p, err := MakeApatPattern(pat, e, true)
		if err != nil {
			continue
		}
		legacy, err := obiapat.MakeApatPattern(pat, e, true)
		if err != nil {
			t.Fatal(err)
		}
		if e >= p.Len() {
			continue // obiapat has no rule on the real bases of a window (see noise.go)
		}
		seq := randomSequence(rng, "acgt", 20+rng.Intn(200), concrete(rng, p))
		begin := rng.Intn(20)
		ls, err := obiapat.MakeApatSequence(newSeq(append([]byte(nil), seq...)), false)
		if err != nil {
			t.Fatal(err)
		}
		clean := true
		for _, m := range legacy.FindAllIndex(ls, begin, -1) {
			if m[0] < begin {
				clean = false
			}
		}
		if !clean {
			continue
		}
		want := legacy.FilterBestMatch(ls, begin, -1)
		got := p.AllMatches(newSeq(seq), begin, -1)
		if len(got) != len(want) {
			t.Fatalf("pattern %q e=%d begin=%d seq=%q\n got  %v\n obiapat %v", pat, e, begin, seq, got, want)
		}
		for i := range got {
			if got[i][1] != want[i][1] || got[i][2] != want[i][2] {
				t.Fatalf("pattern %q e=%d begin=%d seq=%q\n got  %v\n obiapat %v", pat, e, begin, seq, got, want)
			}
		}
		compared += len(got)
	}
	if compared < 1500 {
		t.Errorf("only %d windows compared", compared)
	}
}

// TestBestMatch: BestMatch is the first window with the fewest errors of AllMatches,
// which the reference expectedAllMatches gives for a pattern with indels; for a
// pattern without, it is obiapat's BestMatch.
func TestBestMatch(t *testing.T) {
	rng := rand.New(rand.NewSource(41))
	withErrors, plain := 0, 0
	for iter := 0; iter < iterations(6000); iter++ {
		indel := rng.Intn(3) > 0
		pat := randomPattern(rng, 1+rng.Intn(25), true)
		e := rng.Intn(5)
		p, err := MakeApatPattern(pat, e, indel)
		if err != nil {
			continue
		}
		seq := randomSequence(rng, "acgt", rng.Intn(200), mutate(rng, concrete(rng, p)))
		begin, length := rng.Intn(30), -1
		if rng.Intn(3) == 0 {
			length = rng.Intn(150)
		}
		gs, ge, gn, gm := p.BestMatch(newSeq(seq), begin, length)
		if p.c.indel {
			all := expectedAllMatches(p, seq, begin, length)
			best := -1
			for i, w := range all {
				if best < 0 || w[2] < all[best][2] {
					best = i
				}
			}
			if gm != (best >= 0) || gm && [3]int{gs, ge, gn} != all[best] {
				t.Fatalf("pattern %q e=%d begin=%d length=%d seq=%q: BestMatch %v %v, reference %v", pat, e, begin, length, seq, [3]int{gs, ge, gn}, gm, all)
			}
			if gm && gn > 0 {
				withErrors++
			}
			continue
		}
		if e >= p.Len() {
			continue // obiapat has no rule on the real bases of a window (see noise.go)
		}
		legacy, err := obiapat.MakeApatPattern(pat, e, false)
		if err != nil {
			t.Fatal(err)
		}
		ls, _ := obiapat.MakeApatSequence(newSeq(append([]byte(nil), seq...)), false)
		ws, we, wn, wm := legacy.BestMatch(ls, begin, length)
		// obiapat looks 64 bases past the end of the region, this package does not
		if len(seq) > 0 && (length < 0 || begin+length >= len(seq)) {
			if gm != wm || gm && (gs != ws || ge != we || gn != wn) {
				t.Fatalf("pattern %q e=%d begin=%d seq=%q: BestMatch %d %d %d %v, obiapat %d %d %d %v", pat, e, begin, seq, gs, ge, gn, gm, ws, we, wn, wm)
			}
			plain++
		}
	}
	if withErrors < 300 || plain < 300 {
		t.Errorf("the draws are not varied enough: %d matches with errors (indels), %d without indels", withErrors, plain)
	}
}

// naiveRealBases counts, base by base, the bases of the nominal window [s, e)
// that are not N (the part outside a linear sequence does not count): a window
// must hold more of them than errors.
func naiveRealBases(seq []byte, s, e int, circular bool) int {
	n, count := len(seq), 0
	for i := s; i < e; i++ {
		j := i
		if circular {
			j = ((i % n) + n) % n
		} else if i < 0 || i >= n {
			continue
		}
		if seq[j] != 'n' && seq[j] != 'N' {
			count++
		}
	}
	return count
}
