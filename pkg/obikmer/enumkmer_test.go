package obikmer

import (
	"strings"
	"testing"
)

func collectConstrained(k int, c KmerConstraints) []uint64 {
	var out []uint64
	for km := range IterConstrainedKmers(k, c) {
		out = append(out, km)
	}
	return out
}

func TestIterConstrainedKmersUnconstrained(t *testing.T) {
	got := collectConstrained(4, NoKmerConstraints())
	if len(got) != 256 {
		t.Fatalf("expected 256 k-mers, got %d", len(got))
	}
	for i, km := range got {
		if km != uint64(i) {
			t.Fatalf("not in increasing order at %d: %d", i, km)
		}
	}
}

func TestIterConstrainedKmersNoG(t *testing.T) {
	c := NoKmerConstraints()
	c.MaxCount[2] = 0
	got := collectConstrained(5, c)
	if len(got) != 243 {
		t.Fatalf("expected 3^5 = 243 k-mers, got %d", len(got))
	}
	for _, km := range got {
		if strings.Contains(string(DecodeKmer(km, 5, nil)), "g") {
			t.Fatalf("k-mer %d contains G", km)
		}
	}
}

func TestIterConstrainedKmersMaxRun(t *testing.T) {
	c := NoKmerConstraints()
	c.MaxRun = 2
	// a(n) = 3*(a(n-1)+a(n-2)), a(1)=4, a(2)=16
	want := map[int]int{3: 60, 4: 228}
	for k, n := range want {
		if got := len(collectConstrained(k, c)); got != n {
			t.Fatalf("k=%d: expected %d, got %d", k, n, got)
		}
	}
}

func TestIterConstrainedKmersMaxCount(t *testing.T) {
	c := NoKmerConstraints()
	c.MaxCount[0] = 1 // at most one A
	for _, km := range collectConstrained(6, c) {
		if strings.Count(string(DecodeKmer(km, 6, nil)), "a") > 1 {
			t.Fatalf("k-mer %d has more than one A", km)
		}
	}
}

func TestIterConstrainedKmersImpossible(t *testing.T) {
	c := KmerConstraints{MaxCount: [4]int{1, 1, 1, 0}}
	if got := collectConstrained(4, c); len(got) != 0 {
		t.Fatalf("expected no k-mer, got %d", len(got))
	}
	if got := collectConstrained(0, NoKmerConstraints()); len(got) != 0 {
		t.Fatalf("expected no k-mer for k=0")
	}
}

func TestIterConstrainedKmersEarlyStop(t *testing.T) {
	n := 0
	for range IterConstrainedKmers(10, NoKmerConstraints()) {
		n++
		if n == 5 {
			break
		}
	}
	if n != 5 {
		t.Fatalf("early stop failed: %d", n)
	}
}
