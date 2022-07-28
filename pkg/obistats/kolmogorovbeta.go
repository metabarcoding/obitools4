package obistats

import (
	"math"
	"sort"

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat/distuv"
)

func BetaKolmogorowDist(data []float64, alpha, beta float64, preordered bool) float64 {
	odata := data
	if !preordered {
		odata = make([]float64, len(data))
		copy(odata,data)
		sort.Float64s(odata)
	}

	distances := make([]float64, len(data))
	B := distuv.Beta{
		Alpha: alpha,
		Beta:  beta,
		Src:   nil,
	}

	s := float64(0.0)
	for i, v := range odata {
		s += v
		distances[i] = math.Abs(B.CDF(s) - 1.0/(float64(i)+1.0))
	}

	return floats.Max(distances)
}
