//go:build goexperiment.simd

package obidnapat

import "simd"

func vectorLanes() int {
	var probe simd.Uint64s
	return probe.Len()
}

// block reads blk with the vector kernel of vectorLevels = 2 levels. st holds
// r0, r1 and r2 of every lane (3*lanes values), read and updated; seen receives,
// for every lane, the OR of the end bits of the last level over the block.
func (g *vgroup) block(st []uint64, blk []byte, seen []uint64) {
	vectorBlockKernel(g.lanes, g.masks, g.start, g.end, g.allow, st, blk, seen)
}

func vectorBlockKernel(lanes int, masks, startv, endv, allowv, st []uint64, blk []byte, seen []uint64) {
	start, end, allow := simd.LoadUint64s(startv), simd.LoadUint64s(endv), simd.LoadUint64s(allowv)
	r0, r1, r2 := simd.LoadUint64s(st[0:]), simd.LoadUint64s(st[lanes:]), simd.LoadUint64s(st[2*lanes:])
	var acc simd.Uint64s
	for _, b := range blk {
		mk := simd.LoadUint64s(masks[int(b)*lanes:])
		s0 := r0.Add(r0).Or(start)
		s1 := r1.Add(r1).Or(start)
		s2 := r2.Add(r2).Or(start)
		r2 = s2.And(mk).Or(s1.And(allow))
		r1 = s1.And(mk).Or(s0.And(allow))
		r0 = s0.And(mk)
		acc = acc.Or(r2.And(end))
	}
	r0.Store(st[0:])
	r1.Store(st[lanes:])
	r2.Store(st[2*lanes:])
	acc.Store(seen)
}
