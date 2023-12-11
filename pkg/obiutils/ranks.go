package obiutils

import (
	"sort"
)

// intRanker is a helper type for the rank function.
type intRanker struct {
	x []int // Data to be ranked.
	r []int // A list of indexes into f that reflects rank order after sorting.
}

// ranker satisfies the sort.Interface without mutating the reference slice, f.
func (r intRanker) Len() int           { return len(r.x) }
func (r intRanker) Less(i, j int) bool { return r.x[r.r[i]] < r.x[r.r[j]] }
func (r intRanker) Swap(i, j int)      { r.r[i], r.r[j] = r.r[j], r.r[i] }

// IntOrder sorts a slice of integers and returns a slice
// of indices that represents the order of the sorted
// elements.
//
// `data` is a slice of integers to be ordered.
// Returns a slice of the ordered indices.
func IntOrder(data []int) []int {
	if len(data) == 0 {
		return nil
	}

	r := make([]int, len(data))
	rk := intRanker{
		x: data,
		r: r,
	}

	for i := 0; i < len(data); i++ {
		rk.r[i] = i
	}

	sort.Sort(rk)

	return r
}

func ReverseIntOrder(data []int) []int {
	if len(data) == 0 {
		return nil
	}

	r := make([]int, len(data))
	rk := intRanker{
		x: data,
		r: r,
	}

	for i := 0; i < len(data); i++ {
		rk.r[i] = i
	}

	sort.Sort(sort.Reverse(rk))

	return r
}

type Ranker[T sort.Interface] struct {
	x T     // Data to be ranked.
	r []int // A list of indexes into f that reflects rank order after sorting.
}

// ranker satisfies the sort.Interface without mutating the reference slice, f.
func (r Ranker[_]) Len() int           { return len(r.r) }
func (r Ranker[T]) Less(i, j int) bool { return r.x.Less(r.r[i], r.r[j]) }
func (r Ranker[_]) Swap(i, j int)      { r.r[i], r.r[j] = r.r[j], r.r[i] }

// Order sorts the given data using the provided sort.Interface and returns the sorted indices.
//
// data: The data to be sorted.
// Returns: A slice of integers representing the sorted indices.
func Order[T sort.Interface](data T) []int {
	ldata := data.Len()
	if ldata == 0 {
		return nil
	}
	r := make([]int, ldata)
	rk := Ranker[T]{
		x: data,
		r: r,
	}

	for i := 0; i < ldata; i++ {
		rk.r[i] = i
	}

	sort.Stable(rk)

	return r
}
