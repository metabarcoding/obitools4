package obiutils

import "golang.org/x/exp/constraints"

func Abs[T constraints.Signed](x T) T {
	if x < 0 {
		return -x
	}
	return x
}
