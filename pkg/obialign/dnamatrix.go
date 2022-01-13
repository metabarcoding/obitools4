package obialign

import (
	"math"
)

var __four_bits_count__ = []float64{
	0, // 0000
	1, // 0001
	1, // 0010
	2, // 0011
	1, // 0100
	2, // 0101
	2, // 0110
	3, // 0111
	1, // 1000
	2, // 1001
	2, // 1010
	3, // 1011
	2, // 1100
	3, // 1101
	3, // 1110
	4, // 1111
}

var __initialized_dna_score__ = false

var __nuc_part_match__ [32][32]float64
var __nuc_score_part_match_match__ [100][100]int
var __nuc_score_part_match_mismatch__ [100][100]int

func __match_ratio__(a, b byte) float64 {
	// count of common bits
	cm := __four_bits_count__[a&b&15]

	ca := __four_bits_count__[a&15]
	cb := __four_bits_count__[b&15]

	if cm == 0 || ca == 0 || cb == 0 {
		return float64(0)
	}

	return float64(cm) / float64(ca) / float64(cb)
}

func __logaddexp__(a, b float64) float64 {
	if a > b {
		a, b = b, a
	}

	return b + math.Log1p(math.Exp(a-b))
}

func __match_score_ratio__(a, b byte) (float64, float64) {

	l2 := math.Log(2)
	l3 := math.Log(3)
	l4 := math.Log(4)
	l10 := math.Log(10)
	lE1 := -float64(a)/10*l10 - l4
	lE2 := -float64(b)/10*l10 - l4
	lO1 := math.Log1p(-math.Exp(lE1 + l3))
	lO2 := math.Log1p(-math.Exp(lE2 + l3))
	lO1O2 := lO1 + lO2
	lE1E2 := lE1 + lE2
	lO1E2 := lO1 + lE2
	lO2E1 := lO2 + lE1

	MM := __logaddexp__(lO1O2, lE1E2+l3) + l4
	Mm := __logaddexp__(__logaddexp__(lO1E2, lO2E1), lE1E2+l2) + l4

	return MM, Mm
}

func __init_nuc_part_match__() {

	for i, a := range __four_bits_base_code__ {
		for j, b := range __four_bits_base_code__ {
			__nuc_part_match__[i][j] = __match_ratio__(a, b)
		}
	}
}

func __init_nuc_score_part_match__() {
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			MM, Mm := __match_score_ratio__(byte(i), byte(j))
			__nuc_score_part_match_match__[i][j] = int(MM*10 + 0.5)
			__nuc_score_part_match_mismatch__[i][j] = int(Mm*10 + 0.5)
		}
	}
}

func InitDNAScoreMatrix() {
	if !__initialized_dna_score__ {
		__init_nuc_part_match__()
		__init_nuc_score_part_match__()
		__initialized_dna_score__ = true
	}
}
