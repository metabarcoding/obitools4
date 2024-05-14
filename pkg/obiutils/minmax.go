package obiutils

import (
	"golang.org/x/exp/constraints"
)

func Min[T constraints.Ordered](x, y T) T {
	if x < y {
		return x
	}
	return y
}

func Max[T constraints.Ordered](x, y T) T {
	if x < y {
		return y
	}
	return x
}

func MinMax[T constraints.Ordered](x, y T) (T, T) {
	if x < y {
		return x, y
	}
	return y, x
}


func MinMaxSlice[T constraints.Ordered](vec []T) (min, max T) {
	if len(vec) == 0 {
		panic("empty slice")
	}

	min = vec[0]
	max = vec[0]
	for _, v := range vec {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}

	return
}
