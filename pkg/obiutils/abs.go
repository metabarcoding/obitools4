package obiutils

import "golang.org/x/exp/constraints"

// Abs returns the absolute value of x.
//
// It is a generic function that can be used on any signed type.
func Abs[T constraints.Signed](x T) T {
	if x < 0 {
		return -x
	}

	return x
}
