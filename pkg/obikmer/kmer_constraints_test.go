package obikmer

import (
	"math"
	"math/rand"
	"testing"
)

// IterConstrainedKmers must enumerate exactly the k-mers accepted by Accepts.
func TestIterConstrainedKmersMatchesAccepts(t *testing.T) {
	rng := rand.New(rand.NewSource(10))
	patterns := []string{"", "nnnnnn", "yrnnnn", "cnnnnn", "annnnt"}

	for trial := 0; trial < 60; trial++ {
		k := 6
		c := NoKmerConstraints()
		for b := range 4 {
			if rng.Intn(3) == 0 {
				c.MaxCount[b] = rng.Intn(4)
			}
		}
		if rng.Intn(2) == 0 {
			c.MaxGC = rng.Intn(5)
		}
		if rng.Intn(2) == 0 {
			c.MaxRun = 1 + rng.Intn(3)
		}
		if rng.Intn(2) == 0 {
			c.MinRun = 2 + rng.Intn(3)
		}
		if p := patterns[rng.Intn(len(patterns))]; p != "" {
			c.Accepted, _ = ParseIUPACMasks(p)
		}
		if p := patterns[rng.Intn(len(patterns))]; p != "" && rng.Intn(2) == 0 {
			c.Rejected, _ = ParseIUPACMasks(p)
		}

		var want []uint64
		for km := uint64(0); km < 1<<(2*uint(k)); km++ {
			if c.Accepts(km, k) {
				want = append(want, km)
			}
		}
		got := collectConstrained(k, c)
		if len(got) != len(want) {
			t.Fatalf("trial %d %+v: got %d k-mers, want %d", trial, c, len(got), len(want))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("trial %d %+v: mismatch at %d: %d != %d", trial, c, i, got[i], want[i])
			}
		}
	}
}

func TestConstraintsGCAndRuns(t *testing.T) {
	enc := func(s string) uint64 { return EncodeKmer([]byte(s), len(s)) }

	c := NoKmerConstraints()
	c.MaxGC = 2
	if !c.Accepts(enc("ACGTAA"), 6) || c.Accepts(enc("ACGCAA"), 6) {
		t.Fatal("MaxGC not enforced")
	}

	c = NoKmerConstraints()
	c.MinRun = 3
	if !c.Accepts(enc("ACCCGT"), 6) || c.Accepts(enc("ACCGGT"), 6) {
		t.Fatal("MinRun not enforced")
	}

	c = NoKmerConstraints()
	c.Rejected, _ = ParseIUPACMasks("cnnnnn") // no C in first position
	if c.Accepts(enc("CAAAAA"), 6) || !c.Accepts(enc("AAAAAA"), 6) {
		t.Fatal("Rejected not enforced")
	}
}

func TestRestrictAlphabetAndPatterns(t *testing.T) {
	c := NoKmerConstraints()
	if err := c.RestrictAlphabet("ACT"); err != nil {
		t.Fatal(err)
	}
	if got := len(collectConstrained(4, c)); got != 81 {
		t.Fatalf("expected 3^4 = 81 k-mers, got %d", got)
	}
	c = NoKmerConstraints()
	if err := c.RestrictAlphabet("H"); err != nil { // H = A, C or T
		t.Fatal(err)
	}
	if c.MaxCount != [4]int{-1, -1, 0, -1} {
		t.Fatalf("unexpected counts %v", c.MaxCount)
	}
	if err := c.RestrictAlphabet("AX"); err == nil {
		t.Fatal("invalid IUPAC code must be refused")
	}
	if _, err := ParseIUPACMasks("acgz"); err == nil {
		t.Fatal("invalid IUPAC code must be refused")
	}
	// wrong pattern length
	c = NoKmerConstraints()
	c.Accepted, _ = ParseIUPACMasks("nnn")
	if got := collectConstrained(4, c); len(got) != 0 {
		t.Fatalf("pattern of wrong length must give nothing, got %d", len(got))
	}
}

func TestHammingBound(t *testing.T) {
	cases := []struct {
		q, k, d int
		want    float64
	}{
		{4, 7, 3, 16384.0 / 22},
		{4, 8, 4, 16384.0 / 22}, // d even: reduces to (7, 3)
		{4, 9, 5, 262144.0 / 352},
		{3, 9, 3, 19683.0 / 19},
		{4, 5, 1, 1024},
	}
	for _, c := range cases {
		if got := HammingBound(c.q, c.k, c.d); math.Abs(got-c.want) > 1e-6 {
			t.Errorf("HammingBound(%d,%d,%d) = %v, want %v", c.q, c.k, c.d, got, c.want)
		}
	}
}
