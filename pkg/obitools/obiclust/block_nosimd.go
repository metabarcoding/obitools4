//go:build !goexperiment.simd

package obiclust

// simdBlockWidth is 0 without SIMD: the blocks are not used.
func simdBlockWidth() int { return 0 }

func blockCommon(data []uint8, width int, words []uint16, out []uint8) {
	panic("obiclust: blocks of tables require GOEXPERIMENT=simd")
}
