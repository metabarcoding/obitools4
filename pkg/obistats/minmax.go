package obistats

func Max[T int64 | int32 | int16 | int8 | int | float32 | float64](data []T) T {
	m := data[0]

	for _, v := range data {
		if v > m {
			m = v
		}
	}

	return m
}

func Min[T int64 | int32 | int16 | int8 | int | uint64 | uint32 | uint16 | uint8 | uint | float32 | float64](data []T) T {
	m := data[0]

	for _, v := range data {
		if v < m {
			m = v
		}
	}

	return m
}

func Mode[T int64 | int32 | int16 | int8 | int | uint64 | uint32 | uint16 | uint8 | uint](data []T) T {
	ds := make(map[T]int)

	for _, v := range data {
		ds[v]++
	}

	md := T(0)
	maxocc := 0

	for v, occ := range ds {
		if occ > maxocc {
			maxocc = occ
			md = v
		}
	}

	return md
}
