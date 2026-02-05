package obikmer

import (
	"container/heap"

	"github.com/RoaringBitmap/roaring/roaring64"
)

// heapItem represents an element in the min-heap for k-way merge
type heapItem struct {
	value uint64
	idx   int
}

// kmerMinHeap implements heap.Interface for k-way merge algorithm
type kmerMinHeap []heapItem

func (h kmerMinHeap) Len() int           { return len(h) }
func (h kmerMinHeap) Less(i, j int) bool { return h[i].value < h[j].value }
func (h kmerMinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *kmerMinHeap) Push(x interface{}) {
	*h = append(*h, x.(heapItem))
}

func (h *kmerMinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// QuorumAtLeast returns k-mers present in at least q sets
//
// Algorithm: K-way merge with min-heap counting
//
// The algorithm processes all k-mers in sorted order using a min-heap:
//
//  1. Initialize one iterator per non-empty set
//  2. Build a min-heap of (value, set_index) pairs, one per iterator
//  3. While heap is not empty:
//     a. Extract the minimum value v from heap
//     b. Pop ALL heap items with value == v (counting occurrences)
//     c. If count >= q, add v to result
//     d. Advance each popped iterator and re-insert into heap if valid
//
// This ensures each unique k-mer is counted exactly once across all sets.
//
// Time complexity: O(M log N)
//   - M = sum of all set cardinalities (total k-mer occurrences)
//   - N = number of sets
//   - Each k-mer occurrence is inserted/extracted from heap once: O(M) operations
//   - Each heap operation costs O(log N)
//
// Space complexity: O(N)
//   - Heap contains at most N elements (one per set iterator)
//   - Output bitmap size depends on quorum result
//
// Special cases (optimized):
//   - q <= 0: returns empty set
//   - q == 1: delegates to Union() (native OR operations)
//   - q == n: delegates to Intersect() (native AND operations)
//   - q > n: returns empty set (impossible to satisfy)
func (ksg *KmerSetGroup) QuorumAtLeast(q int) *KmerSet {
	n := len(ksg.sets)

	// Edge cases
	if q <= 0 || n == 0 {
		return NewKmerSet(ksg.k)
	}
	if q > n {
		return NewKmerSet(ksg.k)
	}
	if q == 1 {
		return ksg.Union()
	}
	if q == n {
		return ksg.Intersect()
	}

	// Initialize iterators for all non-empty sets
	iterators := make([]roaring64.IntIterable64, 0, n)
	iterIndices := make([]int, 0, n)

	for i, set := range ksg.sets {
		if set.Len() > 0 {
			iter := set.bitmap.Iterator()
			if iter.HasNext() {
				iterators = append(iterators, iter)
				iterIndices = append(iterIndices, i)
			}
		}
	}

	if len(iterators) == 0 {
		return NewKmerSet(ksg.k)
	}

	// Initialize heap with first value from each iterator
	h := make(kmerMinHeap, len(iterators))
	for i, iter := range iterators {
		h[i] = heapItem{value: iter.Next(), idx: i}
	}
	heap.Init(&h)

	// Result bitmap
	result := roaring64.New()

	// K-way merge with counting
	for len(h) > 0 {
		minVal := h[0].value
		count := 0
		activeIndices := make([]int, 0, len(h))

		// Pop all elements with same value (count occurrences)
		for len(h) > 0 && h[0].value == minVal {
			item := heap.Pop(&h).(heapItem)
			count++
			activeIndices = append(activeIndices, item.idx)
		}

		// Add to result if quorum reached
		if count >= q {
			result.Add(minVal)
		}

		// Advance iterators and re-insert into heap
		for _, iterIdx := range activeIndices {
			if iterators[iterIdx].HasNext() {
				heap.Push(&h, heapItem{
					value: iterators[iterIdx].Next(),
					idx:   iterIdx,
				})
			}
		}
	}

	return NewKmerSetFromBitmap(ksg.k, result)
}

// QuorumAtMost returns k-mers present in at most q sets
//
// Algorithm: Uses the mathematical identity
//   AtMost(q) = Union() - AtLeast(q+1)
//
// Proof:
//   - Union() contains all k-mers present in at least 1 set
//   - AtLeast(q+1) contains all k-mers present in q+1 or more sets
//   - Their difference contains only k-mers present in at most q sets
//
// Implementation:
//  1. Compute U = Union()
//  2. Compute A = QuorumAtLeast(q+1)
//  3. Return U - A using bitmap AndNot operation
//
// Time complexity: O(M log N)
//   - Union(): O(M) with native OR operations
//   - QuorumAtLeast(q+1): O(M log N)
//   - AndNot: O(|U|) where |U| <= M
//   - Total: O(M log N)
//
// Space complexity: O(N)
//   - Inherited from QuorumAtLeast heap
//
// Special cases:
//   - q <= 0: returns empty set
//   - q >= n: returns Union() (all k-mers are in at most n sets)
func (ksg *KmerSetGroup) QuorumAtMost(q int) *KmerSet {
	n := len(ksg.sets)

	// Edge cases
	if q <= 0 {
		return NewKmerSet(ksg.k)
	}
	if q >= n {
		return ksg.Union()
	}

	// Compute Union() - AtLeast(q+1)
	union := ksg.Union()
	atLeastQ1 := ksg.QuorumAtLeast(q + 1)

	// Difference: elements in union but not in atLeastQ1
	result := union.bitmap.Clone()
	result.AndNot(atLeastQ1.bitmap)

	return NewKmerSetFromBitmap(ksg.k, result)
}

// QuorumExactly returns k-mers present in exactly q sets
//
// Algorithm: Uses the mathematical identity
//   Exactly(q) = AtLeast(q) - AtLeast(q+1)
//
// Proof:
//   - AtLeast(q) contains all k-mers present in q or more sets
//   - AtLeast(q+1) contains all k-mers present in q+1 or more sets
//   - Their difference contains only k-mers present in exactly q sets
//
// Implementation:
//  1. Compute A = QuorumAtLeast(q)
//  2. Compute B = QuorumAtLeast(q+1)
//  3. Return A - B using bitmap AndNot operation
//
// Time complexity: O(M log N)
//   - Two calls to QuorumAtLeast: 2 * O(M log N)
//   - One AndNot operation: O(|A|) where |A| <= M
//   - Total: O(M log N) since AndNot is dominated by merge operations
//
// Space complexity: O(N)
//   - Inherited from QuorumAtLeast heap
//   - Two temporary bitmaps for intermediate results
//
// Special cases:
//   - q <= 0: returns empty set
//   - q > n: returns empty set (impossible to have k-mer in more than n sets)
func (ksg *KmerSetGroup) QuorumExactly(q int) *KmerSet {
	n := len(ksg.sets)

	// Edge cases
	if q <= 0 || q > n {
		return NewKmerSet(ksg.k)
	}

	// Compute AtLeast(q) - AtLeast(q+1)
	aq := ksg.QuorumAtLeast(q)
	aq1 := ksg.QuorumAtLeast(q + 1)

	// Difference: elements in aq but not in aq1
	result := aq.bitmap.Clone()
	result.AndNot(aq1.bitmap)

	return NewKmerSetFromBitmap(ksg.k, result)
}
