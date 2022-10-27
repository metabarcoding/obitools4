package goutils

import "sort"

// intRanker is a helper type for the rank function.
type intRanker struct {
	x []int // Data to be ranked.
	r []int // A list of indexes into f that reflects rank order after sorting.
}

// ranker satisfies the sort.Interface without mutating the reference slice, f.
func (r intRanker) Len() int           { return len(r.x) }
func (r intRanker) Less(i, j int) bool { return r.x[r.r[i]] < r.x[r.r[j]] }
func (r intRanker) Swap(i, j int)      { r.r[i], r.r[j] = r.r[j], r.r[i] }

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
