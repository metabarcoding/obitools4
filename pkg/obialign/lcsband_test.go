package obialign

import (
	"math/rand"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// bandDP is the oracle: a plain DP over the whole matrix, written
// independently of the production code, in which only the cells with a
// diagonal offset j-i in [kmin, kmax] exist (kmin=-lB, kmax=lA when
// unbounded). It returns the optimal LCS score and the shortest alignment
// length achieving it among the alignments staying in the band.
func bandDP(a, b []byte, kmin, kmax int) (int, int) {
	lA, lB := len(a), len(b)
	type cell struct {
		s, l int
		ok   bool
	}
	better := func(x, y cell) bool {
		return !y.ok || (x.ok && (x.s > y.s || (x.s == y.s && x.l < y.l)))
	}
	m := make([][]cell, lB+1)
	for i := range m {
		m[i] = make([]cell, lA+1)
	}
	for i := 0; i <= lB; i++ {
		for j := 0; j <= lA; j++ {
			if k := j - i; k < kmin || k > kmax {
				continue
			}
			if i == 0 || j == 0 {
				m[i][j] = cell{0, i + j, true}
				continue
			}
			var best cell
			if d := m[i-1][j-1]; d.ok {
				best = cell{d.s, d.l + 1, true}
				if obiseq.SameIUPACNuc(a[j-1], b[i-1]) {
					best.s++
				}
			}
			if u := m[i-1][j]; u.ok && better(cell{u.s, u.l + 1, true}, best) {
				best = cell{u.s, u.l + 1, true}
			}
			if l := m[i][j-1]; l.ok && better(cell{l.s, l.l + 1, true}, best) {
				best = cell{l.s, l.l + 1, true}
			}
			m[i][j] = best
		}
	}
	return m[lB][lA].s, m[lB][lA].l
}

// fullDP is the unbanded optimum.
func fullDP(a, b []byte) (int, int) { return bandDP(a, b, -len(b), len(a)) }

// expected is the contract of BandedLCS: the unbounded optimum when it has at
// most maxError errors, (-1,-1) otherwise.
func expected(a, b []byte, maxError int) (int, int) {
	s, l := fullDP(a, b)
	if maxError >= 0 && l-s > maxError {
		return -1, -1
	}
	return s, l
}

func randomFrom(rng *rand.Rand, alphabet string, n int) []byte {
	s := make([]byte, n)
	for i := range s {
		s[i] = alphabet[rng.Intn(len(alphabet))]
	}
	return s
}

func mutateFrom(rng *rand.Rand, alphabet string, s []byte, rate float64) []byte {
	out := make([]byte, 0, len(s)+8)
	for _, c := range s {
		if rng.Float64() >= rate {
			out = append(out, c)
			continue
		}
		switch rng.Intn(3) {
		case 0:
			out = append(out, alphabet[rng.Intn(len(alphabet))])
		case 1:
		case 2:
			out = append(out, c, alphabet[rng.Intn(len(alphabet))])
		}
	}
	return out
}

var testAlphabets = []struct {
	name, letters string
	regular       bool // only IUPAC letters of non-zero mask
}{
	{"acgt", "ACGT", true},
	{"low-complexity", "AC", true},
	{"mixed-case", "ACGTacgt", true},
	{"iupac", "ACGTRYSWKMBDHVNUacgtrywsn", true},
	{"exotic", "ACGT-.EXz*", false},
}

// TestBandedLCSDifferential checks the scalar reference against the oracle,
// and the dispatcher (SIMD kernel when built with GOEXPERIMENT=simd) against
// the scalar reference, for random pairs of unequal lengths and many bands.
func TestBandedLCSDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	ws := new(bandedLCSWorkspace)
	for _, alpha := range testAlphabets {
		for it := 0; it < 3000; it++ {
			lA := 1 + rng.Intn(90)
			a := randomFrom(rng, alpha.letters, lA)
			var b []byte
			if rng.Intn(3) > 0 {
				b = mutateFrom(rng, alpha.letters, a, 0.02+rng.Float64()*0.5)
			} else {
				b = randomFrom(rng, alpha.letters, 1+rng.Intn(90))
			}
			if len(b) == 0 {
				continue
			}
			if len(a) < len(b) {
				a, b = b, a
			}
			delta := len(a) - len(b)
			for _, maxError := range []int{-1, 0, 1, 2, 3, 5, 8, 13, 21, 40, 200, delta, delta + 1, delta + 2, delta + 3} {
				ws, wl := expected(a, b, maxError)
				ss, sl := bandedLCS(a, b, maxError, new(bandedLCSWorkspace), false)
				if ss != ws || sl != wl {
					t.Fatalf("%s scalar maxError=%d a=%q b=%q: got (%d,%d), oracle (%d,%d)",
						alpha.name, maxError, a, b, ss, sl, ws, wl)
				}
				ds, dl := bandedLCS(a, b, maxError, new(bandedLCSWorkspace), haveSIMD)
				if ds != ss || dl != sl {
					t.Fatalf("%s dispatch(simd=%v) maxError=%d a=%q b=%q: got (%d,%d), scalar (%d,%d)",
						alpha.name, haveSIMD, maxError, a, b, ds, dl, ss, sl)
				}
			}
		}
	}
	_ = ws
}

// TestBandedLCSLongPairs covers lengths well above the vector width, with
// reused workspaces (stale buffers must never leak between calls).
func TestBandedLCSLongPairs(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	ws := new(bandedLCSWorkspace)
	for it := 0; it < 60; it++ {
		a := randomFrom(rng, "ACGT", 150+rng.Intn(500))
		b := mutateFrom(rng, "ACGT", a, 0.01+rng.Float64()*0.2)
		if len(a) < len(b) {
			a, b = b, a
		}
		for _, maxError := range []int{-1, 4, 20, 60, 150} {
			wantS, wantL := expected(a, b, maxError)
			gotS, gotL := bandedLCS(a, b, maxError, ws, haveSIMD)
			if gotS != wantS || gotL != wantL {
				t.Fatalf("maxError=%d lens=%d/%d: got (%d,%d), oracle (%d,%d)",
					maxError, len(a), len(b), gotS, gotL, wantS, wantL)
			}
		}
	}
}

// TestBandedLCSMatchesLegacy compares with the historical implementation
// (FastLCSEGFScoreByte, not end-gap-free) wherever that one is exact: no
// limit, or a result within the limit.
func TestBandedLCSMatchesLegacy(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	for it := 0; it < 4000; it++ {
		a := randomFrom(rng, "ACGTRYN", 1+rng.Intn(120))
		b := mutateFrom(rng, "ACGTRYN", a, rng.Float64()*0.4)
		if len(b) == 0 {
			continue
		}
		for _, maxError := range []int{-1, 2, 7, 25} {
			ls, ll, err := BandedLCS(a, b, maxError)
			if err != nil {
				t.Fatal(err)
			}
			if ls < 0 {
				continue
			}
			os, ol, _ := FastLCSEGFScoreByte(a, b, maxError, false, nil)
			if os != ls || ol != ll {
				t.Fatalf("maxError=%d a=%q b=%q: BandedLCS (%d,%d), legacy (%d,%d)",
					maxError, a, b, ls, ll, os, ol)
			}
		}
	}
}

func TestBandedLCSEdgeCases(t *testing.T) {
	cases := []struct {
		a, b     string
		maxError int
		s, l     int
	}{
		{"", "", -1, 0, 0},
		{"ACGT", "", -1, 0, 4},
		{"ACGT", "", 3, -1, -1},
		{"ACGT", "", 4, 0, 4},
		{"ACGT", "ACGT", 0, 4, 4},
		{"ACGT", "ACGA", 0, -1, -1},
		{"ACGT", "ACGA", 1, 3, 4},
		{"ACGTT", "ACGT", 0, -1, -1}, // length difference above the limit
		{"ACGTT", "ACGT", 1, 4, 5},
	}
	for _, c := range cases {
		s, l, err := BandedLCS([]byte(c.a), []byte(c.b), c.maxError)
		if err != nil || s != c.s || l != c.l {
			t.Errorf("%q/%q maxError=%d: got (%d,%d,%v), want (%d,%d)", c.a, c.b, c.maxError, s, l, err, c.s, c.l)
		}
		// the arguments are symmetric
		s, l, _ = BandedLCS([]byte(c.b), []byte(c.a), c.maxError)
		if s != c.s || l != c.l {
			t.Errorf("swapped %q/%q maxError=%d: got (%d,%d), want (%d,%d)", c.b, c.a, c.maxError, s, l, c.s, c.l)
		}
	}
}

func TestBandedLCSTooLong(t *testing.T) {
	long := make([]byte, MaxBandedLCSLength+1)
	for i := range long {
		long[i] = 'A'
	}
	if _, _, err := BandedLCS(long, []byte("ACGT"), -1); err != ErrSequenceTooLong {
		t.Errorf("A too long: err = %v", err)
	}
	if _, _, err := BandedLCS([]byte("ACGT"), long, 5); err != ErrSequenceTooLong {
		t.Errorf("B too long: err = %v", err)
	}
	// the maximum length itself is accepted
	ok := long[:MaxBandedLCSLength]
	if s, l, err := BandedLCS(ok, ok, 0); err != nil || s != len(ok) || l != len(ok) {
		t.Errorf("max length: got (%d,%d,%v)", s, l, err)
	}
}

// TestBandedLCSSecondPass checks the exactness over many bands, also on
// low-complexity sequences where the best LCS is not the closest alignment,
// and reports how often the band had to be widened (never more than once).
func TestBandedLCSSecondPass(t *testing.T) {
	rng := rand.New(rand.NewSource(5))
	for _, alpha := range testAlphabets[:4] {
		for _, rate := range []float64{0.02, 0.1, 0.3} {
			ws := new(bandedLCSWorkspace)
			var total int
			for it := 0; it < 1500; it++ {
				a := randomFrom(rng, alpha.letters, 20+rng.Intn(120))
				b := mutateFrom(rng, alpha.letters, a, rate)
				if len(b) == 0 {
					continue
				}
				if len(a) < len(b) {
					a, b = b, a
				}
				for _, maxError := range []int{1, 3, 6, 10, 20, 40} {
					want, wantL := expected(a, b, maxError)
					got, gotL := bandedLCS(a, b, maxError, ws, haveSIMD)
					if got != want || gotL != wantL {
						t.Fatalf("%s maxError=%d a=%q b=%q: got (%d,%d), want (%d,%d)",
							alpha.name, maxError, a, b, got, gotL, want, wantL)
					}
					total++
				}
			}
			t.Logf("%-15s mutation %.2f: %d bounded calls, %d widened (%.1f%%)",
				alpha.name, rate, total, ws.widened, 100*float64(ws.widened)/float64(total))
		}
	}
}

// TestFastLCSScoreWrapper checks the BioSequence entry point.
func TestFastLCSScoreWrapper(t *testing.T) {
	sa := obiseq.NewBioSequence("a", []byte("ACGTACGTAC"), "")
	sb := obiseq.NewBioSequence("b", []byte("ACGTTCGTAC"), "")
	if s, l := FastLCSScore(sa, sb, -1, nil); s != 9 || l != 10 {
		t.Errorf("got (%d,%d), want (9,10)", s, l)
	}
	if s, l := FastLCSScore(sa, sb, 0, nil); s != -1 || l != -1 {
		t.Errorf("got (%d,%d), want (-1,-1)", s, l)
	}
}

// TestBandedLCSAtLeast checks the exactness of the bound expressed on the
// score: the optimum when it reaches minLCS, (-1,-1) otherwise.
func TestBandedLCSAtLeast(t *testing.T) {
	rng := rand.New(rand.NewSource(13))
	var buf BandedLCSBuffer
	for _, alpha := range testAlphabets {
		for it := 0; it < 1500; it++ {
			a := randomFrom(rng, alpha.letters, 1+rng.Intn(100))
			b := mutateFrom(rng, alpha.letters, a, rng.Float64()*0.5)
			if len(b) == 0 {
				continue
			}
			os, ol := fullDP(a, b)
			for _, minLCS := range []int{-3, 0, 1, os - 1, os, os + 1, len(b) / 2, len(b), len(b) + 1} {
				ws, wl := os, ol
				if os < minLCS {
					ws, wl = -1, -1
				}
				for _, got := range [][2]int{func() [2]int { s, l, _ := buf.AtLeast(a, b, minLCS); return [2]int{s, l} }(),
					func() [2]int { s, l, _ := BandedLCSAtLeast(b, a, minLCS); return [2]int{s, l} }()} {
					if got[0] != ws || got[1] != wl {
						t.Fatalf("%s minLCS=%d a=%q b=%q: got (%d,%d), want (%d,%d)",
							alpha.name, minLCS, a, b, got[0], got[1], ws, wl)
					}
				}
			}
		}
	}
	if _, _, err := buf.AtLeast(make([]byte, MaxBandedLCSLength+1), []byte("a"), 0); err != ErrSequenceTooLong {
		t.Errorf("too long: err = %v", err)
	}
}

// TestBandedLCSScoreAtLeast checks the score-only entry point, with sequences
// long enough for the early stop of the kernel to be exercised.
func TestBandedLCSScoreAtLeast(t *testing.T) {
	rng := rand.New(rand.NewSource(17))
	var buf BandedLCSBuffer
	check := func(name string, a, b []byte) {
		os, _ := fullDP(a, b)
		for _, minLCS := range []int{-1, 0, os - 3, os - 1, os, os + 1, os + 5, len(b) - 1, len(b), (len(b) * 9) / 10, len(b) + 1} {
			want := os
			if os < minLCS {
				want = -1
			}
			got, err := buf.ScoreAtLeast(a, b, minLCS)
			if err != nil || got != want {
				t.Fatalf("%s minLCS=%d a=%q b=%q: got %d (%v), want %d", name, minLCS, a, b, got, err, want)
			}
		}
	}
	for _, alpha := range testAlphabets {
		for it := 0; it < 800; it++ {
			a := randomFrom(rng, alpha.letters, 1+rng.Intn(60))
			b := mutateFrom(rng, alpha.letters, a, rng.Float64()*0.6)
			if len(b) > 0 {
				check(alpha.name, a, b)
			}
		}
	}
	for it := 0; it < 150; it++ {
		a := randomFrom(rng, "ACGT", 100+rng.Intn(400))
		b := mutateFrom(rng, "ACGT", a, rng.Float64()*0.3)
		if it%3 == 0 {
			b = randomFrom(rng, "ACGT", 50+rng.Intn(300))
		}
		if len(b) > 0 {
			check("long", a, b)
		}
	}
}
