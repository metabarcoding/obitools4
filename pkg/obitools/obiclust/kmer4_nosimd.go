//go:build !goexperiment.simd

package obiclust

// commonWords is the upper bound sumaclust uses for the number of 4-mers
// shared by two sequences: the sum of the minima of the counters, plus the
// minimum of the two overflows.
func commonWords(t1, t2 *table4) int {
	total := 0
	for i := range t1.count {
		total += int(min(t1.count[i], t2.count[i]))
	}
	return total + min(t1.overflow, t2.overflow)
}
