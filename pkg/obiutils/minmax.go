package obiutils

import "golang.org/x/exp/constraints"

func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func MaxInt(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func MinMaxInt(x, y int) (int, int) {
	if x < y {
		return x, y
	}
	return y, x
}

func MinUInt16(x, y uint16) uint16 {
	if x < y {
		return x
	}
	return y
}

func MaxUInt16(x, y uint16) uint16 {
	if x < y {
		return y
	}
	return x
}

func MinSlice[T constraints.Ordered](vec []T) T {
	if len(vec) == 0 {
		panic("empty slice")
	}
	min := vec[0]
	for _, v := range vec {
		if v < min {
			min = v
		}
	}
	return min
}

func MaxSlice[T constraints.Ordered](vec []T) T {
	if len(vec) == 0 {
		panic("empty slice")
	}
	max := vec[0]
	for _, v := range vec {
		if v > max {
			max = v
		}
	}
	return max
}

func RangeSlice[T constraints.Ordered](vec []T) (min, max T) {
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
