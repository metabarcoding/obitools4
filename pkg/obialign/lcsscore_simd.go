//go:build goexperiment.simd

package obialign

import "simd"

func simdLanes16() int {
	var probe simd.Uint16s
	return probe.Len()
}

func growUint16(s []uint16, n int) []uint16 {
	if cap(s) < n {
		return make([]uint16, n)
	}
	return s[:n]
}

func encodeIUPAC16(dst []uint16, s []byte, rev bool) {
	n := len(s)
	clear(dst)
	if rev {
		for x, c := range s {
			dst[n-1-x] = uint16(iupacMask[c])
		}
		return
	}
	for x, c := range s {
		dst[x] = uint16(iupacMask[c])
	}
}

// pruneEvery is the number of antidiagonals between two tests of the upper
// bound that stops a hopeless alignment.
const pruneEvery = 8

// lcsScoreSIMD runs the same banded DP as bandedLCSSIMD on the score only,
// with 16-bit lanes (twice the cells per instruction):
//
//	cell(i,j) = max(diag + match, up, left)
//
// A cell outside the band is 0, which never beats a real cell. The score of a
// path ending at the cell (i, j) of an antidiagonal cannot exceed the score of
// the cell plus min(lB-i, lA-j), the characters left to pair. Every path
// crosses each antidiagonal inside the window, so when the best score of the
// window plus the largest of these remainders is under minLCS the final score
// is under minLCS: -1 is returned without computing the rest of the matrix.
//
// Requires lA >= lB >= 1, kmin <= 0 <= kmax, and bases with non-zero masks.
// Returns the best score of the band, or -1 if it is proven to be under
// minLCS.
func lcsScoreSIMD(a, b []byte, kmin, kmax, minLCS int, ws *bandedLCSWorkspace) int {
	lA, lB := len(a), len(b)
	vl := simdLanes16()

	ws.aRev16 = growUint16(ws.aRev16, lA+vl)
	ws.bCode16 = growUint16(ws.bCode16, lB+vl)
	encodeIUPAC16(ws.aRev16, a, true)
	encodeIUPAC16(ws.bCode16, b, false)
	aRev, bCode := ws.aRev16, ws.bCode16

	size := lB + 2*vl + 2
	for r := range ws.rows16 {
		ws.rows16[r] = growUint16(ws.rows16[r], size)
		clear(ws.rows16[r])
	}
	p2, p1, cur := ws.rows16[0], ws.rows16[1], ws.rows16[2]

	// the first row and the first column hold a score of 0: nothing to set

	zero := simd.BroadcastUint16s(0)
	one := simd.BroadcastUint16s(1)

	for t := 2; t <= lA+lB; t++ {
		lo := max(1, t-lA, (t-kmax+1)>>1)
		hi := min(lB, t-1, (t-kmin)>>1)

		for i := lo; i <= hi; i += vl {
			ca := simd.LoadUint16s(aRev[lA-t+i:])
			cb := simd.LoadUint16s(bCode[i-1:])
			inc := one.Masked(ca.And(cb).NotEqual(zero))
			d := simd.LoadUint16s(p2[i-1:]).Add(inc)
			u := simd.LoadUint16s(p1[i-1:])
			l := simd.LoadUint16s(p1[i:])
			d.Max(u).Max(l).Store(cur[i:])
		}

		zero.Store(cur[hi+1:])
		if lo > 1 {
			cur[lo-1] = 0
		}

		// A diagonal move skips an antidiagonal, but a path always visits
		// one of two consecutive antidiagonals.
		if t%pruneEvery == 0 && minLCS > 0 &&
			scoreBound(p1, t-1, lA, lB, kmin, kmax) < minLCS &&
			scoreBound(cur, t, lA, lB, kmin, kmax) < minLCS {
			return -1
		}

		p2, p1, cur = p1, cur, p2
	}
	return int(p1[lB])
}

// iota16 is 0, 1, 2... to build the row index of the lanes.
var iota16 = func() (t [64]uint16) {
	for i := range t {
		t[i] = uint16(i)
	}
	return t
}()

// scoreBound is an upper bound of the final score of the paths going through
// the antidiagonal t, whose scores are in row. A path through the cell (i, j)
// scores at most the score of the cell plus min(lB-i, lA-j), the characters
// left to pair, so the bound is the maximum of this sum on the cells of the
// window. The sum is computed by vectors on the full chunks of the window.
func scoreBound(row []uint16, t, lA, lB, kmin, kmax int) int {
	lo := max(1, t-lA, (t-kmax+1)>>1)
	hi := min(lB, t-1, (t-kmin)>>1)

	bound := 0
	i := lo
	if hi-lo >= simdLanes16() {
		vl := simdLanes16()
		idx := simd.LoadUint16s(iota16[:]).Add(simd.BroadcastUint16s(uint16(lo)))
		step := simd.BroadcastUint16s(uint16(vl))
		remB := simd.BroadcastUint16s(uint16(lB))
		// lA-t+i is not negative on the window; the 16-bit arithmetic is modular
		offA := simd.BroadcastUint16s(uint16(lA - t))
		acc := simd.BroadcastUint16s(0)
		for ; i+vl-1 <= hi; i += vl {
			left := remB.Sub(idx).Min(idx.Add(offA))
			acc = acc.Max(simd.LoadUint16s(row[i:]).Add(left))
			idx = idx.Add(step)
		}
		var lanes [64]uint16
		acc.Store(lanes[:])
		for _, v := range lanes[:vl] {
			bound = max(bound, int(v))
		}
	}
	for ; i <= hi; i++ {
		bound = max(bound, int(row[i])+min(lB-i, lA-t+i))
	}

	// the paths running along the first row or the first column have a score of 0
	if t <= lA && t <= kmax {
		bound = max(bound, min(lB, lA-t))
	}
	if t <= lB && -t >= kmin {
		bound = max(bound, min(lB-t, lA))
	}
	return bound
}
