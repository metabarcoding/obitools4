package goutils

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

func MinMaxInt(x, y int) (int,int) {
	if x < y {
		return x,y
	}
	return y,x
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
