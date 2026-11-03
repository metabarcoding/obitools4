//go:build goexperiment.simd

package obialign

import "simd"

const haveSIMD = true

func simdLanes() int {
	var probe simd.Uint32s
	return probe.Len()
}

// encodeIUPAC writes in dst the 4-bit IUPAC mask of each base of s, reversed
// if rev is set, followed by zero padding up to len(dst).
func encodeIUPAC(dst []uint32, s []byte, rev bool) {
	n := len(s)
	for x := range dst {
		dst[x] = 0
	}
	if rev {
		for x, c := range s {
			dst[n-1-x] = uint32(iupacMask[c])
		}
		return
	}
	for x, c := range s {
		dst[x] = uint32(iupacMask[c])
	}
}

// encodeBases encodes both sequences once for all the passes of an alignment.
func encodeBases(ws *bandedLCSWorkspace, a, b []byte) {
	vl := simdLanes()
	ws.aRev = growUint32(ws.aRev, len(a)+vl)
	ws.bCode = growUint32(ws.bCode, len(b)+vl)
	encodeIUPAC(ws.aRev, a, true)
	encodeIUPAC(ws.bCode, b, false)
}

// bandedLCSSIMD runs the banded DP along antidiagonals t = i + j. Lane i of
// the buffer of antidiagonal t holds the cell (i, t-i), so that
//
//	diag = buffer[t-2][i-1]   up = buffer[t-1][i-1]   left = buffer[t-1][i]
//
// are plain unaligned loads. Along an antidiagonal the offset j-i = t-2i
// must stay in [kmin, kmax], which gives the window [lo, hi] of rows. The
// vector loop may run past hi; the lanes beyond are then reset to 0
// (unavailable), as is the lane just before the window, because the next
// antidiagonals read them as neighbours. Lane 0 and lane t carry the first
// row and the first column of the matrix.
//
// Requires encodeBases to have been called, lA >= lB >= 1 and kmin <= 0 <=
// kmax.
func bandedLCSSIMD(lA, lB, kmin, kmax int, ws *bandedLCSWorkspace) uint32 {
	vl := simdLanes()
	aRev, bCode := ws.aRev, ws.bCode

	size := lB + 2*vl + 2
	for r := range ws.rows {
		ws.rows[r] = growUint32(ws.rows[r], size)
		clear(ws.rows[r])
	}
	p2, p1, cur := ws.rows[0], ws.rows[1], ws.rows[2]

	p2[0] = lcsLenBase // cell (0,0)
	if kmax >= 1 {
		p1[0] = lcsLenBase - 1 // cell (0,1)
	}
	if kmin <= -1 {
		p1[1] = lcsLenBase - 1 // cell (1,0)
	}

	zero := simd.BroadcastUint32s(0)
	one := simd.BroadcastUint32s(1)
	unit := simd.BroadcastUint32s(lcsScoreUnit)

	for t := 2; t <= lA+lB; t++ {
		lo := max(1, t-lA, (t-kmax+1)>>1)
		hi := min(lB, t-1, (t-kmin)>>1)

		for i := lo; i <= hi; i += vl {
			ca := simd.LoadUint32s(aRev[lA-t+i:])
			cb := simd.LoadUint32s(bCode[i-1:])
			inc := unit.Masked(ca.And(cb).NotEqual(zero))
			d := simd.LoadUint32s(p2[i-1:]).Add(inc)
			u := simd.LoadUint32s(p1[i-1:])
			l := simd.LoadUint32s(p1[i:])
			d.Max(u).Max(l).Sub(one).Store(cur[i:])
		}

		zero.Store(cur[hi+1:])
		if lo > 1 {
			cur[lo-1] = 0
		}
		cur[0] = 0
		if t <= lA && t <= kmax {
			cur[0] = lcsLenBase - uint32(t)
		}
		if t <= lB && -t >= kmin {
			cur[t] = lcsLenBase - uint32(t)
		}

		p2, p1, cur = p1, cur, p2
	}
	return p1[lB]
}
