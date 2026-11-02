//go:build goexperiment.simd

package obiclust

import "simd"

// commonWords is the upper bound sumaclust uses for the number of 4-mers
// shared by two sequences: the sum of the minima of the counters, plus the
// minimum of the two overflows.
//
// The minima of the 256 counters are computed by vectors of bytes, and summed
// in 16-bit lanes (at most 256*255 spread over the lanes, the lanes being at
// least 8 in number, so that none of them overflows).
func commonWords(t1, t2 *table4) int {
	var probe simd.Uint8s
	vl := probe.Len()

	acc := simd.BroadcastUint16s(0)
	low := simd.BroadcastUint16s(0x00FF)
	for i := 0; i < len(t1.count); i += vl {
		m := simd.LoadUint8s(t1.count[i:]).Min(simd.LoadUint8s(t2.count[i:])).ReshapeToUint16s()
		acc = acc.Add(m.And(low)).Add(m.ShiftAllRight(8))
	}

	// sum the pairs of 16-bit lanes into 32-bit lanes, then add those up
	wide := acc.ReshapeToUint32s()
	sums := wide.And(simd.BroadcastUint32s(0xFFFF)).Add(wide.ShiftAllRight(16))
	var lanes [32]uint32
	sums.Store(lanes[:])
	total := 0
	for _, v := range lanes[:sums.Len()] {
		total += int(v)
	}
	return total + min(t1.overflow, t2.overflow)
}
