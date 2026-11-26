//go:build !goexperiment.simd

package obidnapat

// Without GOEXPERIMENT=simd every word runs its scalar kernel.

func vectorLanes() int { return 0 }

func (g *vgroup) block(st []uint64, blk []byte, seen []uint64) {
	panic("obidnapat: no vector kernel in this build") // unreachable: no group is built
}
