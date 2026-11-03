//go:build goexperiment.simd

package obisimilarity

import "simd"

// simdBlockWidth is the number of sequences in a block: one per lane of a
// vector of bytes.
func simdBlockWidth() int {
	var probe simd.Uint8s
	return min(probe.Len(), 64)
}

// blockCommon computes, for the sequence whose words are given, the sum of
// the minima of the counters with every lane of the block. It is the value of
// commonWords without the overflow term, as the sums fit in 8 bits (see
// SparseWords). out must hold at least `width` bytes.
func blockCommon(data []uint8, width int, words []uint16, out []uint8) {
	acc := simd.BroadcastUint8s(0)
	for _, e := range words {
		w := simd.LoadUint8s(data[int(e>>8)*width:])
		acc = acc.Add(w.Min(simd.BroadcastUint8s(uint8(e))))
	}
	acc.Store(out)
}
