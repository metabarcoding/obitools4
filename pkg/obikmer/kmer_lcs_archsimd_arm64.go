//go:build goexperiment.simd && arm64

package obikmer

import "simd/archsimd"

// neonLanes is the number of 16-bit lanes of a NEON 128-bit register.
const neonLanes = 8

// kmerLCSBatchNEON is kmerLCSBatch written directly on archsimd NEON types
// (see kmer_lcs_simd.go for the cell encoding and recurrence).
func kmerLCSBatchNEON(query uint64, partners []uint64, k int, out []KmerLCSResult) {
	var qbase [32]uint16
	for i := 0; i < k; i++ {
		qbase[i] = uint16(query>>(2*uint(k-1-i))) & 3
	}

	pbase := make([]uint16, k*neonLanes)
	row := make([]uint16, (k+1)*neonLanes)
	one := archsimd.BroadcastUint16x8(1)
	bonus := archsimd.BroadcastUint16x8(lcsScoreUnit)

	for start := 0; start < len(partners); start += neonLanes {
		chunk := partners[start:min(start+neonLanes, len(partners))]

		for lane, p := range chunk {
			for j := 0; j < k; j++ {
				pbase[j*neonLanes+lane] = uint16(p>>(2*uint(k-1-j))) & 3
			}
		}

		for j := 0; j <= k; j++ {
			archsimd.BroadcastUint16x8(uint16(lcsLenBase - j)).Store(row[j*neonLanes:])
		}

		for i := 1; i <= k; i++ {
			qv := archsimd.BroadcastUint16x8(qbase[i-1])
			diag := archsimd.LoadUint16x8(row)
			left := archsimd.BroadcastUint16x8(uint16(lcsLenBase - i))
			left.Store(row)

			for j := 1; j <= k; j++ {
				up := archsimd.LoadUint16x8(row[j*neonLanes:])
				pv := archsimd.LoadUint16x8(pbase[(j-1)*neonLanes:])
				match := qv.Equal(pv)
				best := diag.Add(bonus.Masked(match)).Max(up).Max(left).Sub(one)
				best.Store(row[j*neonLanes:])
				diag = up
				left = best
			}
		}

		last := row[k*neonLanes:]
		for lane := range chunk {
			v := int(last[lane])
			out[start+lane] = KmerLCSResult{
				LCS:      v / lcsScoreUnit,
				AlignLen: lcsLenBase - v%lcsScoreUnit,
			}
		}
	}
}
