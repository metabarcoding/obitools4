//go:build !goexperiment.simd

package obialign

const haveSIMD = false

// encodeBases and bandedLCSSIMD are never called without GOEXPERIMENT=simd
// (haveSIMD is false); they exist to keep the dispatcher uniform.
func encodeBases(ws *bandedLCSWorkspace, a, b []byte) {}

func bandedLCSSIMD(lA, lB, kmin, kmax int, ws *bandedLCSWorkspace) uint32 {
	return 0
}
