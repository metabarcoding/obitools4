package obistats



func Max[T int64 | int32 | int16 | int8 | int | float32 | float64] (data []T) T {
	m := data[0]

	for _,v := range data {
		if v > m {
			m = v
		}
	}

	return m
}

func Min[T int64 | int32 | int16 | int8 | int | float32 | float64] (data []T) T {
	m := data[0]

	for _,v := range data {
		if v < m {
			m = v
		}
	}

	return m
}