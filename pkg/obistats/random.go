package obistats

import "math/rand"

func SampleIntWithoutReplacemant(n, max int) []int {

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
