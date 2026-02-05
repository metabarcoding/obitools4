package obikmer

import (
	"testing"
)

// TestQuorumAtLeastEdgeCases tests edge cases for QuorumAtLeast
func TestQuorumAtLeastEdgeCases(t *testing.T) {
	k := 5

	// Test group with all empty sets
	emptyGroup := NewKmerSetGroup(k, 3)
	result := emptyGroup.QuorumAtLeast(1)
	if result.Len() != 0 {
		t.Errorf("Empty sets: expected 0 k-mers, got %d", result.Len())
	}

	// Test q <= 0
	group := NewKmerSetGroup(k, 3)
	result = group.QuorumAtLeast(0)
	if result.Len() != 0 {
		t.Errorf("q=0: expected 0 k-mers, got %d", result.Len())
	}

	result = group.QuorumAtLeast(-1)
	if result.Len() != 0 {
		t.Errorf("q=-1: expected 0 k-mers, got %d", result.Len())
	}

	// Test q > n
	group.Get(0).AddKmerCode(1)
	result = group.QuorumAtLeast(10)
	if result.Len() != 0 {
		t.Errorf("q>n: expected 0 k-mers, got %d", result.Len())
	}
}

// TestQuorumAtLeastQ1 tests q=1 (should equal Union)
func TestQuorumAtLeastQ1(t *testing.T) {
	k := 5
	group := NewKmerSetGroup(k, 3)

	// Add different k-mers to each set
	group.Get(0).AddKmerCode(1)
	group.Get(0).AddKmerCode(2)
	group.Get(1).AddKmerCode(2)
	group.Get(1).AddKmerCode(3)
	group.Get(2).AddKmerCode(3)
	group.Get(2).AddKmerCode(4)

	quorum := group.QuorumAtLeast(1)
	union := group.Union()

	if quorum.Len() != union.Len() {
		t.Errorf("QuorumAtLeast(1) length %d != Union length %d", quorum.Len(), union.Len())
	}

	// Check all elements match
	for kmer := uint64(1); kmer <= 4; kmer++ {
		if quorum.Contains(kmer) != union.Contains(kmer) {
			t.Errorf("Mismatch for k-mer %d", kmer)
		}
	}
}

// TestQuorumAtLeastQN tests q=n (should equal Intersect)
func TestQuorumAtLeastQN(t *testing.T) {
	k := 5
	group := NewKmerSetGroup(k, 3)

	// Add some common k-mers and some unique
	for i := 0; i < 3; i++ {
		group.Get(i).AddKmerCode(10) // common to all
		group.Get(i).AddKmerCode(20) // common to all
	}
	group.Get(0).AddKmerCode(1) // unique to set 0
	group.Get(1).AddKmerCode(2) // unique to set 1

	quorum := group.QuorumAtLeast(3)
	intersect := group.Intersect()

	if quorum.Len() != intersect.Len() {
		t.Errorf("QuorumAtLeast(n) length %d != Intersect length %d", quorum.Len(), intersect.Len())
	}

	if quorum.Len() != 2 {
		t.Errorf("Expected 2 common k-mers, got %d", quorum.Len())
	}

	if !quorum.Contains(10) || !quorum.Contains(20) {
		t.Error("Missing common k-mers")
	}

	if quorum.Contains(1) || quorum.Contains(2) {
		t.Error("Unique k-mers should not be in result")
	}
}

// TestQuorumAtLeastGeneral tests general quorum values
func TestQuorumAtLeastGeneral(t *testing.T) {
	k := 5
	group := NewKmerSetGroup(k, 5)

	// Setup: k-mer i appears in i sets (for i=1..5)
	// k-mer 1: in set 0
	// k-mer 2: in sets 0,1
	// k-mer 3: in sets 0,1,2
	// k-mer 4: in sets 0,1,2,3
	// k-mer 5: in sets 0,1,2,3,4 (all)

	for kmer := uint64(1); kmer <= 5; kmer++ {
		for setIdx := 0; setIdx < int(kmer); setIdx++ {
			group.Get(setIdx).AddKmerCode(kmer)
		}
	}

	tests := []struct {
		q        int
		expected map[uint64]bool
	}{
		{1, map[uint64]bool{1: true, 2: true, 3: true, 4: true, 5: true}},
		{2, map[uint64]bool{2: true, 3: true, 4: true, 5: true}},
		{3, map[uint64]bool{3: true, 4: true, 5: true}},
		{4, map[uint64]bool{4: true, 5: true}},
		{5, map[uint64]bool{5: true}},
	}

	for _, tt := range tests {
		result := group.QuorumAtLeast(tt.q)

		if result.Len() != uint64(len(tt.expected)) {
			t.Errorf("q=%d: expected %d k-mers, got %d", tt.q, len(tt.expected), result.Len())
		}

		for kmer := uint64(1); kmer <= 5; kmer++ {
			shouldContain := tt.expected[kmer]
			doesContain := result.Contains(kmer)
			if shouldContain != doesContain {
				t.Errorf("q=%d, k-mer=%d: expected contains=%v, got %v", tt.q, kmer, shouldContain, doesContain)
			}
		}
	}
}

// TestQuorumExactlyBasic tests QuorumExactly basic functionality
func TestQuorumExactlyBasic(t *testing.T) {
	k := 5
	group := NewKmerSetGroup(k, 5)

	// Setup: k-mer i appears in exactly i sets
	for kmer := uint64(1); kmer <= 5; kmer++ {
		for setIdx := 0; setIdx < int(kmer); setIdx++ {
			group.Get(setIdx).AddKmerCode(kmer)
		}
	}

	tests := []struct {
		q        int
		expected []uint64
	}{
		{1, []uint64{1}},
		{2, []uint64{2}},
		{3, []uint64{3}},
		{4, []uint64{4}},
		{5, []uint64{5}},
	}

	for _, tt := range tests {
		result := group.QuorumExactly(tt.q)

		if result.Len() != uint64(len(tt.expected)) {
			t.Errorf("q=%d: expected %d k-mers, got %d", tt.q, len(tt.expected), result.Len())
		}

		for _, kmer := range tt.expected {
			if !result.Contains(kmer) {
				t.Errorf("q=%d: missing k-mer %d", tt.q, kmer)
			}
		}
	}
}

// TestQuorumIdentity tests the mathematical identity: Exactly(q) = AtLeast(q) - AtLeast(q+1)
func TestQuorumIdentity(t *testing.T) {
	k := 5
	group := NewKmerSetGroup(k, 4)

	// Add random distribution
	group.Get(0).AddKmerCode(1)
	group.Get(0).AddKmerCode(2)
	group.Get(0).AddKmerCode(3)

	group.Get(1).AddKmerCode(2)
	group.Get(1).AddKmerCode(3)
	group.Get(1).AddKmerCode(4)

	group.Get(2).AddKmerCode(3)
	group.Get(2).AddKmerCode(4)

	group.Get(3).AddKmerCode(4)

	for q := 1; q <= 4; q++ {
		exactly := group.QuorumExactly(q)
		atLeast := group.QuorumAtLeast(q)
		atLeastPlus1 := group.QuorumAtLeast(q + 1)

		// Verify: every element in exactly(q) is in atLeast(q)
		iter := exactly.Iterator()
		for iter.HasNext() {
			kmer := iter.Next()
			if !atLeast.Contains(kmer) {
				t.Errorf("q=%d: k-mer %d in Exactly but not in AtLeast", q, kmer)
			}
			if atLeastPlus1.Contains(kmer) {
				t.Errorf("q=%d: k-mer %d in Exactly but also in AtLeast(q+1)", q, kmer)
			}
		}
	}
}

// TestQuorumDisjointSets tests quorum on completely disjoint sets
func TestQuorumDisjointSets(t *testing.T) {
	k := 5
	group := NewKmerSetGroup(k, 3)

	// Each set has unique k-mers
	group.Get(0).AddKmerCode(1)
	group.Get(1).AddKmerCode(2)
	group.Get(2).AddKmerCode(3)

	// q=1 should give all
	result := group.QuorumAtLeast(1)
	if result.Len() != 3 {
		t.Errorf("Disjoint sets q=1: expected 3, got %d", result.Len())
	}

	// q=2 should give none
	result = group.QuorumAtLeast(2)
	if result.Len() != 0 {
		t.Errorf("Disjoint sets q=2: expected 0, got %d", result.Len())
	}
}

// TestQuorumIdenticalSets tests quorum on identical sets
func TestQuorumIdenticalSets(t *testing.T) {
	k := 5
	group := NewKmerSetGroup(k, 3)

	// All sets have same k-mers
	for i := 0; i < 3; i++ {
		group.Get(i).AddKmerCode(10)
		group.Get(i).AddKmerCode(20)
		group.Get(i).AddKmerCode(30)
	}

	// Any q <= n should give all k-mers
	for q := 1; q <= 3; q++ {
		result := group.QuorumAtLeast(q)
		if result.Len() != 3 {
			t.Errorf("Identical sets q=%d: expected 3, got %d", q, result.Len())
		}
	}
}

// TestQuorumLargeNumbers tests with large k-mer values
func TestQuorumLargeNumbers(t *testing.T) {
	k := 21
	group := NewKmerSetGroup(k, 3)

	// Use large uint64 values (actual k-mer encodings)
	largeKmers := []uint64{
		0x1234567890ABCDEF,
		0xFEDCBA0987654321,
		0xAAAAAAAAAAAAAAAA,
	}

	// Add to multiple sets
	for i := 0; i < 3; i++ {
		for j := 0; j <= i; j++ {
			group.Get(j).AddKmerCode(largeKmers[i])
		}
	}

	result := group.QuorumAtLeast(2)
	if result.Len() != 2 {
		t.Errorf("Large numbers q=2: expected 2, got %d", result.Len())
	}

	if !result.Contains(largeKmers[1]) || !result.Contains(largeKmers[2]) {
		t.Error("Large numbers: wrong k-mers in result")
	}
}

// TestQuorumAtMostBasic tests QuorumAtMost basic functionality
func TestQuorumAtMostBasic(t *testing.T) {
	k := 5
	group := NewKmerSetGroup(k, 5)

	// Setup: k-mer i appears in exactly i sets
	for kmer := uint64(1); kmer <= 5; kmer++ {
		for setIdx := 0; setIdx < int(kmer); setIdx++ {
			group.Get(setIdx).AddKmerCode(kmer)
		}
	}

	tests := []struct {
		q        int
		expected []uint64
	}{
		{0, []uint64{}},                          // at most 0: none
		{1, []uint64{1}},                         // at most 1: only k-mer 1
		{2, []uint64{1, 2}},                      // at most 2: k-mers 1,2
		{3, []uint64{1, 2, 3}},                   // at most 3: k-mers 1,2,3
		{4, []uint64{1, 2, 3, 4}},                // at most 4: k-mers 1,2,3,4
		{5, []uint64{1, 2, 3, 4, 5}},             // at most 5: all k-mers
		{10, []uint64{1, 2, 3, 4, 5}},            // at most 10: all k-mers
	}

	for _, tt := range tests {
		result := group.QuorumAtMost(tt.q)

		if result.Len() != uint64(len(tt.expected)) {
			t.Errorf("q=%d: expected %d k-mers, got %d", tt.q, len(tt.expected), result.Len())
		}

		for _, kmer := range tt.expected {
			if !result.Contains(kmer) {
				t.Errorf("q=%d: missing k-mer %d", tt.q, kmer)
			}
		}
	}
}

// TestQuorumComplementIdentity tests that AtLeast and AtMost are complementary
func TestQuorumComplementIdentity(t *testing.T) {
	k := 5
	group := NewKmerSetGroup(k, 4)

	// Add random distribution
	group.Get(0).AddKmerCode(1)
	group.Get(0).AddKmerCode(2)
	group.Get(0).AddKmerCode(3)

	group.Get(1).AddKmerCode(2)
	group.Get(1).AddKmerCode(3)
	group.Get(1).AddKmerCode(4)

	group.Get(2).AddKmerCode(3)
	group.Get(2).AddKmerCode(4)

	group.Get(3).AddKmerCode(4)

	union := group.Union()

	for q := 1; q < 4; q++ {
		atMost := group.QuorumAtMost(q)
		atLeast := group.QuorumAtLeast(q + 1)

		// Verify: AtMost(q) ∪ AtLeast(q+1) = Union()
		combined := atMost.Union(atLeast)

		if combined.Len() != union.Len() {
			t.Errorf("q=%d: AtMost(q) ∪ AtLeast(q+1) has %d k-mers, Union has %d",
				q, combined.Len(), union.Len())
		}

		// Verify: AtMost(q) ∩ AtLeast(q+1) = ∅
		overlap := atMost.Intersect(atLeast)
		if overlap.Len() != 0 {
			t.Errorf("q=%d: AtMost(q) and AtLeast(q+1) overlap with %d k-mers",
				q, overlap.Len())
		}
	}
}

// BenchmarkQuorumAtLeast benchmarks quorum operations
func BenchmarkQuorumAtLeast(b *testing.B) {
	k := 21
	n := 10
	group := NewKmerSetGroup(k, n)

	// Populate with realistic data
	for i := 0; i < n; i++ {
		for j := uint64(0); j < 10000; j++ {
			if (j % uint64(n)) <= uint64(i) {
				group.Get(i).AddKmerCode(j)
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = group.QuorumAtLeast(5)
	}
}
