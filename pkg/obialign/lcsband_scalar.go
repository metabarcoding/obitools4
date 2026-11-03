package obialign

import "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"

// bandedLCSScalar is the portable reference of the banded DP, row by row. It
// returns the packed cell (lB, lA). len(a) >= len(b) >= 1.
//
// Cells outside the band hold 0 (unavailable). Row i spans the columns
// [i+kmin, i+kmax] clipped to [1, lA]; the cell just left of that span and
// the one just right of it are reset so that the next row never reads a stale
// value.
func bandedLCSScalar(a, b []byte, kmin, kmax int, ws *bandedLCSWorkspace) uint32 {
	lA, lB := len(a), len(b)

	prev := growUint32(ws.rows[0], lA+2)
	cur := growUint32(ws.rows[1], lA+2)
	ws.rows[0], ws.rows[1] = prev, cur

	for j := 0; j <= lA; j++ {
		prev[j] = 0
		if j <= kmax {
			prev[j] = lcsLenBase - uint32(j)
		}
	}
	prev[lA+1] = 0

	for i := 1; i <= lB; i++ {
		if -i >= kmin {
			cur[0] = lcsLenBase - uint32(i)
		} else {
			cur[0] = 0
		}
		jlo := max(1, i+kmin)
		jhi := min(lA, i+kmax)
		if jlo > 1 {
			cur[jlo-1] = 0
		}
		bi := b[i-1]
		for j := jlo; j <= jhi; j++ {
			d := prev[j-1]
			if obiseq.SameIUPACNuc(a[j-1], bi) {
				d += lcsScoreUnit
			}
			m := max(d, prev[j], cur[j-1], 1)
			cur[j] = m - 1
		}
		if jhi < lA {
			cur[jhi+1] = 0
		}
		prev, cur = cur, prev
	}
	return prev[lA]
}
