package obistats

import (
	"math"
	"math/rand"

	"gonum.org/v1/gonum/mathext"
	"scientificgo.org/special"
)

type BetaBinomial struct {
	N int
	// Alpha is the left shape parameter of the distribution. Alpha must be greater
	// than 0.
	Alpha float64
	// Beta is the right shape parameter of the distribution. Beta must be greater
	// than 0.
	Beta float64

	Src rand.Source
}

func (b BetaBinomial) LogCDFTable(x int) []float64 {
	if x > b.N {
		x = b.N
	}

	tab := make([]float64, x+1)
	tab[0] = 0.0

	for i := 1; i <= x; i++ {
		tab[i] = LogAddExp(tab[i-1], b.LogProb(i))
	}

	return tab
}

// CDF computes the value of the cumulative distribution function at x.
func (b BetaBinomial) LogCDF(x int) float64 {
	if b.Alpha <= 0 || b.Beta <= 0 || b.N <= 0 {
		panic("beta-binomial: negative parameters")
	}

	if x <= 0 {
		return 0
	}
	if x >= b.N {
		return 1
	}

	fn := float64(b.N)
	fx := float64(x)

	lv := Lchoose(b.N, x) + mathext.Lbeta(fx+b.Alpha, fn-fx+b.Beta) - mathext.Lbeta(b.Alpha, b.Beta)
	lv += math.Log(special.HypPFQ(
		[]float64{1, -fx, fn - fx + b.Beta},
		[]float64{fn - fx - 1, 1 - fx - b.Alpha},
		1))
	return lv
}

func (b BetaBinomial) CDF(x int) float64 {
	return math.Exp(b.LogCDF(x))
}

// LogProb computes the value of the neperian logarithm of the probability density function at x.
func (b BetaBinomial) LogProb(x int) float64 {
	if x < 0 || x > b.N {
		return math.Inf(-1)
	}

	if b.Alpha <= 0 || b.Beta <= 0 || b.N <= 0 {
		panic("beta-binomial: negative parameters")
	}

	fn := float64(b.N)
	fx := float64(x)
	return Lchoose(b.N, x) + mathext.Lbeta(fx+b.Alpha, fn-fx+b.Beta) - mathext.Lbeta(b.Alpha, b.Beta)
}

// Prob computes the value of the probability density function at x.
func (b BetaBinomial) Prob(x int) float64 {
	return math.Exp(b.LogProb(x))
}

func (b BetaBinomial) Mean() float64 {
	return float64(b.N) * b.Alpha / (b.Alpha + b.Beta)
}

// Variance returns the variance of the probability distribution.
func (b BetaBinomial) Variance() float64 {
	return float64(b.N) * b.Alpha * b.Beta * (float64(b.N) + b.Alpha + b.Beta) / (b.Alpha + b.Beta) / (b.Alpha + b.Beta) / (b.Alpha + b.Beta + 1)
}

// StdDev returns the standard deviation of the probability distribution.
func (b BetaBinomial) StdDev() float64 {
	return math.Sqrt(b.Variance())
}

// Mode returns the mode of the distribution.
//
// Mode returns NaN if both parameters are less than or equal to 1 as a special case,
// 0 if only Alpha <= 1 and 1 if only Beta <= 1.
func (b BetaBinomial) Mode() float64 {
	if b.Alpha <= 1 {
		if b.Beta <= 1 {
			return math.NaN()
		}
		return 0
	}
	if b.Beta <= 1 {
		return float64(b.N)
	}
	return float64(b.N) * (b.Alpha - 1) / (b.Alpha + b.Beta - 2)
}

func (b BetaBinomial) NumParameters() int {
	return 3
}
