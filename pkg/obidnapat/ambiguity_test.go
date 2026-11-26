package obidnapat

import (
	"math/rand"
	"testing"
)

// globalCost is the fewest errors to align the whole text with the whole
// pattern (same costs as the search; nothing inserted before the first position).
func globalCost(p ApatPattern, text []byte) int {
	const inf = 1 << 20
	ps := p.c.positions
	m := len(ps)
	cost := func(i int) int {
		if ps[i].oblig {
			return inf
		}
		return 1
	}
	prev := make([]int, len(text)+1)
	cur := make([]int, len(text)+1)
	for j := 1; j <= len(text); j++ {
		prev[j] = inf
	}
	for i := 1; i <= m; i++ {
		cur[0] = prev[0] + cost(i-1)
		for j := 1; j <= len(text); j++ {
			sub := prev[j-1]
			if !(ps[i-1].wildcard || letterSet[text[j-1]]&ps[i-1].set != 0) {
				sub += cost(i - 1)
			}
			cur[j] = min(sub, prev[j]+cost(i-1), cur[j-1]+cost(i-1))
		}
		prev, cur = cur, prev
	}
	return prev[len(text)]
}

// startRange gives the earliest and the latest start of a window that ends at pos
// and aligns with the whole pattern for at most e errors.
func startRange(p ApatPattern, seq []byte, pos, e, lo int) (earliest, latest int, ok bool) {
	for s := pos; s >= max(lo, 0); s-- {
		if globalCost(p, seq[s:pos+1]) <= e {
			if !ok {
				latest = s
			}
			earliest, ok = s, true
		}
	}
	return
}

// TestAmbiguityOfTheStart measures, on simulated reads, the share of the primer
// matches whose start (or end) is not determined: several windows of the same
// minimal cost. It asserts nothing but sanity: it is a measure.
func TestAmbiguityOfTheStart(t *testing.T) {
	type model struct {
		name         string
		sub, del, in float64
	}
	models := []model{
		{"substitutions 1%", 0.01, 0, 0},
		{"substitutions 1%, indels 0.2% each", 0.01, 0.002, 0.002},
		{"oligo synthesis: sub 0.5%, deletions 1%, insertions 0.2%", 0.005, 0.01, 0.002},
		{"degraded: sub 3%, indels 1% each", 0.03, 0.01, 0.01},
	}
	nReads := iterations(3000)
	for _, md := range models {
		rng := rand.New(rand.NewSource(41))
		var total, matched, withErr, ambStart, ambEnd, ambAny, widthSum int
		for r := 0; r < nReads; r++ {
			p, err := MakeApatPattern(randomACGT(rng, 20), 2, true)
			if err != nil {
				t.Fatal(err)
			}
			var site []byte
			for _, pos := range p.c.positions {
				c := "acgt"[rng.Intn(4)]
				for letterSet[c]&pos.set == 0 {
					c = "acgt"[rng.Intn(4)]
				}
				x := rng.Float64()
				switch {
				case x < md.del:
					continue
				case x < md.del+md.sub:
					c = "acgt"[rng.Intn(4)]
				}
				if rng.Float64() < md.in {
					site = append(site, "acgt"[rng.Intn(4)])
				}
				site = append(site, c)
			}
			seq := randomSequence(rng, "acgt", 150, "")
			at := 20 + rng.Intn(60)
			copy(seq[at:], site)
			total++

			start, end, nerr, ok := p.BestMatch(newSeq(seq), 0, -1)
			if !ok {
				continue
			}
			matched++
			if nerr == 0 {
				continue
			}
			withErr++
			earliest, latest, found := startRange(p, seq, end-1, nerr, 0)
			if !found || latest != start {
				t.Fatalf("backward start %d, oracle latest %d (found %v)", start, latest, found)
			}
			sa := earliest != latest
			// the ends: other hits of the same minimal cost next to this one
			ea := false
			for _, h := range p.c.search(nil, seq, 0, -1, false) {
				if h[2] == nerr && h[1] != end && h[1] >= end-3 && h[1] <= end+3 {
					ea = true
				}
			}
			if sa {
				ambStart++
				widthSum += latest - earliest
			}
			if ea {
				ambEnd++
			}
			if sa || ea {
				ambAny++
			}
		}
		pc := func(n, d int) float64 {
			if d == 0 {
				return 0
			}
			return 100 * float64(n) / float64(d)
		}
		t.Logf("%-58s | reads %d, primer found %d, with errors %d (%.1f%% of found) | start ambiguous %d (%.1f%% of found, %.1f%% of those with errors, mean width %.1f) | end ambiguous %d (%.1f%% of found) | either %.1f%% of found",
			md.name, total, matched, withErr, pc(withErr, matched), ambStart, pc(ambStart, matched), pc(ambStart, withErr),
			float64(widthSum)/float64(max(ambStart, 1)), ambEnd, pc(ambEnd, matched), pc(ambAny, matched))
	}
}

func randomACGT(rng *rand.Rand, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = "ACGT"[rng.Intn(4)]
	}
	return string(b)
}
