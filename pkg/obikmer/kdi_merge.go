package obikmer

import "container/heap"

// mergeItem represents an element in the min-heap for k-way merge.
type mergeItem struct {
	value uint64
	idx   int // index of the reader that produced this value
}

// mergeHeap implements heap.Interface for k-way merge.
type mergeHeap []mergeItem

func (h mergeHeap) Len() int            { return len(h) }
func (h mergeHeap) Less(i, j int) bool  { return h[i].value < h[j].value }
func (h mergeHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *mergeHeap) Push(x interface{}) { *h = append(*h, x.(mergeItem)) }
func (h *mergeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// KWayMerge performs a k-way merge of multiple sorted KdiReader streams.
// For each unique k-mer value, it reports the value and the number of
// input streams that contained it (count).
type KWayMerge struct {
	h       mergeHeap
	readers []*KdiReader
}

// NewKWayMerge creates a k-way merge from multiple KdiReaders.
// Each reader must produce values in sorted (ascending) order.
func NewKWayMerge(readers []*KdiReader) *KWayMerge {
	m := &KWayMerge{
		h:       make(mergeHeap, 0, len(readers)),
		readers: readers,
	}

	// Initialize heap with first value from each reader
	for i, r := range readers {
		if v, ok := r.Next(); ok {
			m.h = append(m.h, mergeItem{value: v, idx: i})
		}
	}
	heap.Init(&m.h)

	return m
}

// Next returns the next smallest k-mer value, the number of readers
// that contained this value (count), and true.
// Returns (0, 0, false) when all streams are exhausted.
func (m *KWayMerge) Next() (kmer uint64, count int, ok bool) {
	if len(m.h) == 0 {
		return 0, 0, false
	}

	minVal := m.h[0].value
	count = 0

	// Pop all items with the same value
	for len(m.h) > 0 && m.h[0].value == minVal {
		item := heap.Pop(&m.h).(mergeItem)
		count++
		// Advance that reader
		if v, ok := m.readers[item.idx].Next(); ok {
			heap.Push(&m.h, mergeItem{value: v, idx: item.idx})
		}
	}

	return minVal, count, true
}

// Close closes all underlying readers.
func (m *KWayMerge) Close() error {
	var firstErr error
	for _, r := range m.readers {
		if err := r.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
