//go:build !goexperiment.simd

package obisimilarity

// commonWords is the upper bound used for the number of 4-mers shared by two
// sequences: the sum of the minima of the counters, plus the minimum of the
// two overflows.
func commonWords(t1, t2 *Table4) int {
	total := 0
	for i := range t1.Count {
		total += int(min(t1.Count[i], t2.Count[i]))
	}
	return total + min(t1.Overflow, t2.Overflow)
}
