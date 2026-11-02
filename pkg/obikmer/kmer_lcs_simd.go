//go:build goexperiment.simd

package obikmer

import "simd"

// Each lane of a vector holds the state of one (query, partner) alignment as
//
//	score<<6 | (63 - alignment length)
//
// so that taking the maximum of two cells ranks them by score first and, for
// equal scores, by shortest alignment. With k <= 31 the score is <= 31 and the
// alignment length <= 62, hence the value fits in 11 bits and the low field
// is never exhausted. The three DP moves then all reduce to
//
//	cell(i,j) = max(diag + 64*match, up, left) - 1
//
// (every move lengthens the alignment by one; only a diagonal match adds 1
// to the score).
const (
	lcsScoreUnit = 64
	lcsLenBase   = 63
)

func kmerLCSBatch(query uint64, partners []uint64, k int, out []KmerLCSResult) {
	var probe simd.Uint16s
	lanes := probe.Len()

	var qbase [32]uint16
	for i := 0; i < k; i++ {
		qbase[i] = uint16(query>>(2*uint(k-1-i))) & 3
	}

	// partner bases, transposed: pbase[j*lanes+lane] is base j of the partner in `lane`
	pbase := make([]uint16, k*lanes)
	// one DP row, one vector per column
	row := make([]uint16, (k+1)*lanes)
	one := simd.BroadcastUint16s(1)
	bonus := simd.BroadcastUint16s(lcsScoreUnit)

	for start := 0; start < len(partners); start += lanes {
		chunk := partners[start:min(start+lanes, len(partners))]

		for lane, p := range chunk {
			for j := 0; j < k; j++ {
				pbase[j*lanes+lane] = uint16(p>>(2*uint(k-1-j))) & 3
			}
		}

		for j := 0; j <= k; j++ {
			simd.BroadcastUint16s(uint16(lcsLenBase - j)).Store(row[j*lanes:])
		}

		for i := 1; i <= k; i++ {
			qv := simd.BroadcastUint16s(qbase[i-1])
			diag := simd.LoadUint16s(row)
			left := simd.BroadcastUint16s(uint16(lcsLenBase - i))
			left.Store(row)

			for j := 1; j <= k; j++ {
				up := simd.LoadUint16s(row[j*lanes:])
				pv := simd.LoadUint16s(pbase[(j-1)*lanes:])
				match := qv.Equal(pv)
				best := diag.Add(bonus.Masked(match)).Max(up).Max(left).Sub(one)
				best.Store(row[j*lanes:])
				diag = up
				left = best
			}
		}

		last := row[k*lanes:]
		for lane := range chunk {
			v := int(last[lane])
			out[start+lane] = KmerLCSResult{
				LCS:      v / lcsScoreUnit,
				AlignLen: lcsLenBase - v%lcsScoreUnit,
			}
		}
	}
}
