package obikmer

import (
	"math"

	log "github.com/sirupsen/logrus"
)

// DefaultMinimizerSize returns ceil(k / 2.5) as a reasonable default minimizer size.
func DefaultMinimizerSize(k int) int {
	m := int(math.Ceil(float64(k) / 2.5))
	if m < 1 {
		m = 1
	}
	if m >= k {
		m = k - 1
	}
	return m
}

// MinMinimizerSize returns the minimum m such that 4^m >= nworkers,
// i.e. ceil(log(nworkers) / log(4)).
func MinMinimizerSize(nworkers int) int {
	if nworkers <= 1 {
		return 1
	}
	return int(math.Ceil(math.Log(float64(nworkers)) / math.Log(4)))
}

// ValidateMinimizerSize checks and adjusts the minimizer size to satisfy constraints:
// - m >= ceil(log(nworkers)/log(4))
// - 1 <= m < k
func ValidateMinimizerSize(m, k, nworkers int) int {
	minM := MinMinimizerSize(nworkers)
	if m < minM {
		log.Warnf("Minimizer size %d too small for %d workers (4^%d = %d < %d), adjusting to %d",
			m, nworkers, m, 1<<(2*m), nworkers, minM)
		m = minM
	}
	if m < 1 {
		m = 1
	}
	if m >= k {
		m = k - 1
	}
	return m
}
