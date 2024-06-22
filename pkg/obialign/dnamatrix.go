package obialign

import (
	"math"
	"sync"

	log "github.com/sirupsen/logrus"
)

var _FourBitsCount = []float64{
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

var _InitializedDnaScore = false

var _NucPartMatch [32][32]float64
var _NucScorePartMatchMatch [100][100]int
var _NucScorePartMatchMismatch [100][100]int
var _InitDNAScoreMatrixMutex = &sync.Mutex{}

// _MatchRatio calculates the match ratio between two bytes.
//
// It takes two parameters, a and b, which are bytes to be compared.
// The function returns a float64 value representing the match ratio.
func _MatchRatio(a, b byte) float64 {
	// count of common bits
	cm := _FourBitsCount[a&b&15]

	// count of bits in a
	ca := _FourBitsCount[a&15]

	// count of bits in b
	cb := _FourBitsCount[b&15]

	// check if any of the counts is zero
	if cm == 0 || ca == 0 || cb == 0 {
		return float64(0)
	}

	// calculate the match ratio
	return float64(cm) / float64(ca) / float64(cb)
}

// _Logaddexp calculates the logarithm of the sum of exponentials of two given numbers.
//
// Parameters:
//
//	a - the first number (float64)
//	b - the second number (float64)
//
// Returns:
//
//	float64 - the result of the calculation
func _Logaddexp(a, b float64) float64 {
	if a > b {
		a, b = b, a
	}

	return b + math.Log1p(math.Exp(a-b))
}

// _MatchScoreRatio calculates the match score ratio between two bytes.
//
// Parameters:
// - a: the first byte
// - b: the second byte
//
// Returns:
// - float64: the match score ratio when a match is observed
// - float64: the match score ratio when a mismatch is observed
func _MatchScoreRatio(a, b byte) (float64, float64) {

	l2 := math.Log(2)
	l3 := math.Log(3)
	l10 := math.Log(10)
	lalea := math.Log(4)                   // 1 /(change of the random model)
	lE1 := -float64(a)/10*l10 - l3         // log proba of sequencing error on A/3
	lE2 := -float64(b)/10*l10 - l3         // log proba of sequencing error on B/3
	lO1 := math.Log1p(-math.Exp(lE1 + l3)) // log proba no being an error on A
	lO2 := math.Log1p(-math.Exp(lE2 + l3)) // log proba no being an error on B
	lO1O2 := lO1 + lO2
	lE1E2 := lE1 + lE2
	lO1E2 := lO1 + lE2
	lO2E1 := lO2 + lE1

	MM := _Logaddexp(lO1O2, lE1E2+l3)                    // Proba match when match observed
	Mm := _Logaddexp(_Logaddexp(lO1E2, lO2E1), lE1E2+l2) // Proba match when mismatch observed

	return MM + lalea, Mm + lalea
}

func _InitNucPartMatch() {

	for i, a := range _FourBitsBaseCode {
		for j, b := range _FourBitsBaseCode {
			_NucPartMatch[i][j] = _MatchRatio(a, b)
		}
	}
}

func _InitNucScorePartMatch() {
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			MM, Mm := _MatchScoreRatio(byte(i), byte(j))
			_NucScorePartMatchMatch[i][j] = int(MM*10 + 0.5)
			_NucScorePartMatchMismatch[i][j] = int(Mm*10 + 0.5)
		}
	}
}

func _InitDNAScoreMatrix() {
	_InitDNAScoreMatrixMutex.Lock()
	defer _InitDNAScoreMatrixMutex.Unlock()
	if !_InitializedDnaScore {
		log.Info("Initializing the DNA Scoring matrix")

		_InitNucPartMatch()
		_InitNucScorePartMatch()
		_InitializedDnaScore = true
	}
}
