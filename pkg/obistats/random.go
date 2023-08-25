package obistats

import "math/rand"

// SampleIntWithoutReplacement generates a random sample of unique integers without replacement.
//
// Generates a random sample of n unique integers without replacement included in the range [0, max).
//
// Parameters:
//   - n: the number of integers to generate.
//   - max: the maximum value for the generated integers.
//
// Returns:
//   - []int: a slice of integers containing the generated sample.
func SampleIntWithoutReplacement(n, max int) []int {

	draw := make(map[int]int, n)

	for i := 0; i < n; i++ {
		y := rand.Intn(max)
		x, ok := draw[y]
		if ok {
			y = x
		}
		draw[y] = max - 1
		max--
	}

	res := make([]int, 0, n)
	for i := range draw {
		res = append(res, i)
	}

	return res
}
