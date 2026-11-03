package obikmer

import (
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"
)

// The occurrence budget only bounds the memory: any budget, down to one
// occurrence (every partition then goes alone), gives the same counts.
func TestSeqKmerCounterBudgetDoesNotChangeTheCounts(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	seqs := randomSeqs(rng, 40, 20, 500, true)
	build := func(budget int) []SeqKmerCount {
		c, err := NewSeqKmerCounter(t.TempDir(), 9, 4, 11)
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		if budget > 0 {
			c.SetOccurrenceBudget(budget)
		}
		for i, s := range seqs {
			c.AddSequence(uint32(i), s, i%2 == 0)
		}
		total := 0
		for _, pw := range c.parts {
			total += pw.kmers
		}
		if total == 0 {
			t.Fatal("no occurrence counted")
		}
		got, err := c.CountAtLeast(2)
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	want := build(0)
	if len(want) == 0 {
		t.Fatal("the test needs k-mers shared by several sequences")
	}
	for _, budget := range []int{1, 100, 5000} {
		if got := build(budget); !reflect.DeepEqual(got, want) {
			t.Errorf("budget %d changes the counts", budget)
		}
	}
}

// The occurrences announced for a partition are those that its file holds.
func TestSeqKmerCounterPartitionOccurrenceCount(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	seqs := randomSeqs(rng, 25, 10, 300, true)
	c, err := NewSeqKmerCounter(t.TempDir(), 8, 3, 5)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	for i, s := range seqs {
		c.AddSequence(uint32(i), s, i%3 == 0)
	}
	if err := c.seal(); err != nil {
		t.Fatal(err)
	}
	for p, pw := range c.parts {
		occ, err := c.loadPartition(p, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(occ) != pw.kmers {
			t.Errorf("partition %d: %d occurrences loaded, %d counted when written", p, len(occ), pw.kmers)
		}
	}
}

func TestOccurrenceBudget(t *testing.T) {
	b := newOccurrenceBudget(10)
	if got := b.acquire(4); got != 4 {
		t.Fatalf("acquire(4) = %d", got)
	}
	// 7 does not fit in the 6 left: it must wait for the release
	done := make(chan int)
	go func() { done <- b.acquire(7) }()
	select {
	case <-done:
		t.Fatal("acquire(7) must wait")
	case <-time.After(50 * time.Millisecond):
	}
	b.release(4)
	if got := <-done; got != 7 {
		t.Fatalf("acquire(7) = %d", got)
	}
	b.release(7)

	// a request above the whole budget takes all of it, once everything is free
	if got := b.acquire(1000); got != 10 {
		t.Fatalf("an oversized request must take the whole budget, got %d", got)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); b.release(b.acquire(3)) }()
	time.Sleep(20 * time.Millisecond)
	b.release(10)
	wg.Wait()
}

func TestDefaultPartitions(t *testing.T) {
	for _, tc := range []struct{ kmers, want int }{
		{0, 256}, {5_000_000, 256}, {256 << 20, 256}, {257 << 20, 512}, {526_000_000, 512},
		{600 << 20, 1024}, {1 << 40, 1024},
	} {
		if got := DefaultPartitions(tc.kmers); got != tc.want {
			t.Errorf("DefaultPartitions(%d) = %d, want %d", tc.kmers, got, tc.want)
		}
	}
}
