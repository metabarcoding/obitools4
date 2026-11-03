package obikmer

import (
	"math/rand"
	"slices"
	"strings"
	"sync"
	"testing"
)

type bruteInfo struct {
	seqs     map[uint32]int // sequence id -> occurrences
	repeated bool
}

// bruteForce is the oracle: every window of k unambiguous bases, circular
// junction included, encoded with EncodeCanonicalKmer.
func bruteForce(seqs [][]byte, circular []bool, k int) map[uint64]*bruteInfo {
	res := map[uint64]*bruteInfo{}
	for id, s := range seqs {
		if len(s) < k {
			continue
		}
		if circular[id] {
			s = append(slices.Clone(s), s[:k-1]...)
		}
	window:
		for i := 0; i+k <= len(s); i++ {
			w := s[i : i+k]
			for _, b := range w {
				if !unambiguousBase[b] {
					continue window
				}
			}
			kmer := EncodeCanonicalKmer(w, k)
			info := res[kmer]
			if info == nil {
				info = &bruteInfo{seqs: map[uint32]int{}}
				res[kmer] = info
			}
			info.seqs[uint32(id)]++
			if info.seqs[uint32(id)] > 1 {
				info.repeated = true
			}
		}
	}
	return res
}

func randomSeqs(rng *rand.Rand, n, minLen, maxLen int, withN bool) [][]byte {
	alphabet := "acgtACGT"
	if withN {
		alphabet = "acgtACGTacgtACGTNnR"
	}
	// A shared motif and its reverse complement make k-mers common to several
	// sequences on both strands.
	motif := make([]byte, 40)
	for i := range motif {
		motif[i] = "ACGT"[rng.Intn(4)]
	}
	rc := make([]byte, len(motif))
	for i, b := range motif {
		rc[len(motif)-1-i] = map[byte]byte{'A': 'T', 'C': 'G', 'G': 'C', 'T': 'A'}[b]
	}

	out := make([][]byte, n)
	for i := range out {
		l := minLen + rng.Intn(maxLen-minLen+1)
		s := make([]byte, l)
		for j := range s {
			s[j] = alphabet[rng.Intn(len(alphabet))]
		}
		mode := rng.Intn(4)
		if l < 60 { // too short to host the motifs
			mode = 3
		}
		switch mode {
		case 0:
			copy(s[l/4:], motif)
		case 1:
			copy(s[l/4:], rc)
		case 2: // internal repeat, both strands
			copy(s[2:], motif[:20])
			copy(s[l-25:], rc[:20])
		}
		out[i] = s
	}
	return out
}

func TestSeqKmerCounterAgainstBruteForce(t *testing.T) {
	cases := []struct {
		name    string
		k, m, p int
		withN   bool
		circ    bool
	}{
		{"k7-linear", 7, 3, 8, false, false},
		{"k7-circular-N", 7, 3, 5, true, true},
		{"k18-linear-N", 18, 8, 16, true, false},
		{"k18-circular", 18, 8, 7, false, true},
		{"k31-p1", 31, 10, 1, false, true},
		{"k4-small-m", 4, 1, 3, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(int64(tc.k*100 + tc.p)))
			seqs := randomSeqs(rng, 30, 5, 400, tc.withN)
			circ := make([]bool, len(seqs))
			for i := range circ {
				circ[i] = tc.circ && i%3 != 0
			}

			c, err := NewSeqKmerCounter(t.TempDir(), tc.k, tc.m, tc.p)
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()

			// Concurrent insertion, shuffled order, ids not contiguous.
			var wg sync.WaitGroup
			for i := range seqs {
				wg.Add(1)
				go func() {
					defer wg.Done()
					c.AddSequence(uint32(i), seqs[i], circ[i])
				}()
			}
			wg.Wait()

			want := bruteForce(seqs, circ, tc.k)
			for _, q := range []int{1, 2, 3, 10} {
				got, err := c.CountAtLeast(q)
				if err != nil {
					t.Fatal(err)
				}
				if !slices.IsSortedFunc(got, func(a, b SeqKmerCount) int {
					switch {
					case a.Kmer < b.Kmer:
						return -1
					case a.Kmer > b.Kmer:
						return 1
					}
					return 0
				}) {
					t.Fatal("result is not sorted by k-mer")
				}

				n := 0
				for kmer, info := range want {
					if len(info.seqs) >= q {
						n++
						_ = kmer
					}
				}
				if len(got) != n {
					t.Fatalf("q=%d: %d k-mers, want %d", q, len(got), n)
				}
				for _, g := range got {
					info := want[g.Kmer]
					if info == nil {
						t.Fatalf("q=%d: unexpected k-mer %d", q, g.Kmer)
					}
					if int(g.Sequences) != len(info.seqs) || g.Repeated != info.repeated {
						t.Fatalf("q=%d kmer %d: got (%d,%v) want (%d,%v)",
							q, g.Kmer, g.Sequences, g.Repeated, len(info.seqs), info.repeated)
					}
				}
			}
		})
	}
}

// A k-mer appearing on both strands of one sequence counts once but is Repeated.
func TestSeqKmerCounterBothStrandsOneSequence(t *testing.T) {
	const k = 8
	fwd := "ACGGTCAT"
	rc := "ATGACCGT"
	seq := []byte("TTTT" + fwd + "GGGGGGGG" + rc + "CCCC")

	c, err := NewSeqKmerCounter(t.TempDir(), k, 3, 4)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.AddSequence(7, seq, false)

	got, err := c.CountAtLeast(1)
	if err != nil {
		t.Fatal(err)
	}
	target := EncodeCanonicalKmer([]byte(fwd), k)
	found := false
	for _, g := range got {
		if g.Kmer == target {
			found = true
			if g.Sequences != 1 || !g.Repeated {
				t.Errorf("got %+v, want 1 sequence and Repeated", g)
			}
		}
	}
	if !found {
		t.Fatal("k-mer not found")
	}
}

// Ambiguous bases must never be part of a k-mer, and the circular junction
// must produce the wrap-around k-mers.
func TestSeqKmerCounterAmbiguityAndCircular(t *testing.T) {
	const k = 5
	c, err := NewSeqKmerCounter(t.TempDir(), k, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	c.AddSequence(0, []byte("ACGTNACGT"), false) // runs ACGT (<k) and ACGT (<k): nothing
	c.AddSequence(1, []byte("AACCGGTT"), true)   // 8 circular windows

	got, err := c.CountAtLeast(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("expected k-mers from the circular sequence")
	}
	var circ []string
	for i := 0; i < 8; i++ {
		var sb strings.Builder
		for j := 0; j < k; j++ {
			sb.WriteByte("AACCGGTT"[(i+j)%8])
		}
		circ = append(circ, sb.String())
	}
	want := map[uint64]bool{}
	for _, w := range circ {
		want[EncodeCanonicalKmer([]byte(w), k)] = true
	}
	if len(got) != len(want) {
		t.Fatalf("got %d k-mers, want %d", len(got), len(want))
	}
	for _, g := range got {
		if !want[g.Kmer] || g.Sequences != 1 {
			t.Errorf("unexpected %+v", g)
		}
	}
}

func TestSeqKmerCounterParameters(t *testing.T) {
	if _, err := NewSeqKmerCounter(t.TempDir(), 1, 1, 4); err == nil {
		t.Error("k=1 must fail")
	}
	if _, err := NewSeqKmerCounter(t.TempDir(), 32, 5, 4); err == nil {
		t.Error("k=32 must fail")
	}
	if _, err := NewSeqKmerCounter(t.TempDir(), 10, 10, 4); err == nil {
		t.Error("m>=k must fail")
	}
	if _, err := NewSeqKmerCounter(t.TempDir(), 10, 4, 0); err == nil {
		t.Error("zero partitions must fail")
	}
	c, err := NewSeqKmerCounter(t.TempDir(), 10, 0, 4) // default minimizer
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("second Close must be harmless: %v", err)
	}
}
